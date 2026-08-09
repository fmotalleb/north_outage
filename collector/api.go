package collector

import (
	"bytes"
	"cmp"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fmotalleb/go-jalali"
	"github.com/fmotalleb/go-tools/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/internal/otel"
	"github.com/fmotalleb/north_outage/internal/template"
	"github.com/fmotalleb/north_outage/models"
)

const (
	citySari  = "ساری"
	cityBabol = "بابل"
)

var defaultCityMap = map[int]string{
	2:  citySari,
	3:  citySari,
	4:  citySari,
	5:  citySari,
	6:  citySari,
	7:  "میاندرود",
	13: cityBabol,
	14: "نکا",
	21: "گلوگاه",
	22: "بهشهر",
	23: "بهشهر",
	25: cityBabol,
	26: "بهشهر",
	31: "قائمشهر",
	32: "قائمشهر",
	33: "سیمرغ",
	34: "قائمشهر",
	42: "سوادکوه",
	43: "سوادکوه شمالی",
	44: "سوادکوه",
	46: "سوادکوه",
	51: "جویبار",
	52: "جویبار",
	53: "بابلسر",
	61: cityBabol,
	62: cityBabol,
	64: cityBabol,
	65: cityBabol,
	66: cityBabol,
	67: cityBabol,
	68: cityBabol,
	71: "آمل",
	72: "آمل",
	73: "آمل",
	74: "آمل",
	75: "آمل",
	76: "آمل",
	84: "نکا",
	85: "بابلسر",
	86: "فریدونکنار",
	87: citySari,
}

const defaultBodyTemplate = `{"fromDate":"{{ now | jFormat "2006/01/02" | faNum }}","toDate":"{{ now | dateModify "24h" | jFormat "2006/01/02" | faNum }}","city":-1,"pgds":""}`

func fetchData(ctx context.Context) ([]models.Event, error) {
	ctx, logger := log.AsNamedChild(ctx, "fetchData")
	cfg, err := config.Get(ctx)
	if err != nil {
		logger.Error("failed to fetch config from context", zap.Error(err))
		return nil, err
	}
	collectorCfg := cfg.CollectorConfig

	ctx, span := otel.CollectorTracer("north_outage.collector").Start(ctx, "collector.fetch",
		trace.WithAttributes(attribute.String("collector.endpoint", collectorCfg.Endpoint)),
	)
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, collectorCfg.Timeout)
	defer cancel()

	bodyStr, err := template.EvaluateTemplate(defaultBodyTemplate, nil)
	if err != nil {
		logger.Error("failed to evaluate body template", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to evaluate body template")
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, collectorCfg.Endpoint, bytes.NewBufferString(bodyStr))
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return nil, err
	}

	transport := &http.Transport{}
	if collectorCfg.Proxy != nil {
		transport.Proxy = http.ProxyURL(collectorCfg.Proxy)
	}
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: !collectorCfg.SSLVerify,
	}
	client := &http.Client{Transport: otelhttp.NewTransport(transport, otelhttp.WithTracerProvider(otel.CollectorProvider()))}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to send request", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to send request")
		return nil, err
	}
	defer resp.Body.Close()
	var response OutageResponse
	_, parseSpan := otel.CollectorTracer("north_outage.collector").Start(ctx, "collector.parse")
	err = json.NewDecoder(resp.Body).Decode(&response)
	parseSpan.SetAttributes(attribute.Int("collector.raw_events", len(response.OutageList)))
	if err != nil {
		logger.Error("failed to parse response", zap.Error(err))
		parseSpan.RecordError(err)
		parseSpan.SetStatus(codes.Error, "failed to parse response")
		parseSpan.End()
		return nil, err
	}
	parseSpan.End()
	span.SetAttributes(attribute.Int("collector.events", len(response.OutageList)))
	events := normalize(ctx, response, logger, collectorCfg)
	return events, nil
}

func normalize(ctx context.Context, response OutageResponse, logger *zap.Logger, collectorCfg config.Collector) []models.Event {
	_, span := otel.CollectorTracer("north_outage.collector").Start(ctx, "collector.mapper")
	defer span.End()
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		logger.Warn("failed to load Asia/Tehran timezone, falling back to UTC", zap.Error(err))
		loc = time.UTC
	}
	events := make([]models.Event, 0, len(response.OutageList))
	for _, v := range response.OutageList {
		city, ok := defaultCityMap[v.City]
		if !ok {
			logger.Error("city id is not found in city mapper", zap.Any("event", v))
			continue
		}
		// sometimes outages omit time, maybe unplanned ones?
		ot := cmp.Or(v.OutageTime, "00:00")
		date, err := jalali.ParseInLocation("2006/01/02 15:04", v.OutageDate+" "+ot, loc)
		if err != nil {
			logger.Error("failed to parse jalali start date", zap.Error(err), zap.Any("event", v))
			continue
		}
		duration := collectorCfg.PlannedDuration
		if !v.IsPlanned {
			duration = collectorCfg.UnPlannedDuration
		}
		start := date.ToGregorian()
		if start.Before(time.Now()) {
			logger.Debug("event is happened already; skipping", zap.Any("event", v))
			continue
		}
		ev := &models.Event{
			City:    city,
			Address: persianFixer(v.Address),
			Start:   start,
			End:     start.Add(duration),
		}
		ev.ResetHash()
		events = append(events, *ev)
	}
	span.SetAttributes(attribute.Int("collector.mapped_events", len(events)))
	return events
}

type OutageResponse struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	OutageList []Outage `json:"outageList"`
}

type Outage struct {
	RegDate      string `json:"reg_date"`
	Registerer   string `json:"registerer"`
	ReasonOutage string `json:"reason_outage"`
	OutageDate   string `json:"outage_date"`
	OutageTime   string `json:"outage_time"`
	IsPlanned    bool   `json:"is_planned"`
	Address      string `json:"address"`
	OutageNumber int64  `json:"outage_number"`
	City         int    `json:"city"`
}
