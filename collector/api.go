package collector

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fmotalleb/go-jalali"
	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/internal/template"
	"github.com/fmotalleb/north_outage/models"
)

var defaultCityMap = map[int]string{
	2:  "ساری",
	3:  "ساری",
	4:  "ساری",
	5:  "ساری",
	6:  "ساری",
	7:  "میاندرود",
	13: "بابل",
	14: "نکا",
	21: "گلوگاه",
	22: "بهشهر",
	23: "بهشهر",
	25: "بابل",
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
	61: "بابل",
	62: "بابل",
	64: "بابل",
	65: "بابل",
	66: "بابل",
	67: "بابل",
	68: "بابل",
	71: "آمل",
	72: "آمل",
	73: "آمل",
	74: "آمل",
	75: "آمل",
	76: "آمل",
	84: "نکا",
	85: "بابلسر",
	86: "فریدونکنار",
	87: "ساری",
}

const defaultBodyTemplate = `{"fromDate":"{{ now | jFormat "2006/01/02" | fanum }}","toDate":"{{ now | dateModify "24h" | jFormat "2006/01/02" | fanum }}","city":-1,"pgds":""}`

func fetchData(ctx context.Context) ([]models.Event, error) {
	logger := log.Of(ctx)
	cfg, err := config.Get(ctx)
	if err != nil {
		logger.Error("failed to fetch config from context", zap.Error(err))
		return nil, err
	}
	collectorCfg := cfg.CollectorConfig

	ctx, cancel := context.WithTimeout(ctx, collectorCfg.Timeout)
	defer cancel()

	bodyStr, err := template.EvaluateTemplate(defaultBodyTemplate, nil)
	if err != nil {
		logger.Error("failed to evaluate body template", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, collectorCfg.Endpoint, bytes.NewBufferString(bodyStr))
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return nil, err
	}

	client := &http.Client{}
	transport := &http.Transport{}
	if collectorCfg.Proxy != nil {
		transport.Proxy = http.ProxyURL(collectorCfg.Proxy)
	}
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: !collectorCfg.SSLVerify,
	}
	client.Transport = transport
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to send request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()
	var response OutageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logger.Error("failed to parse response", zap.Error(err))
		return nil, err
	}
	events := normalize(response, logger, collectorCfg)
	return events, nil
}

func normalize(response OutageResponse, logger *zap.Logger, collectorCfg config.Collector) []models.Event {
	events := make([]models.Event, 0, len(response.OutageList))
	for _, v := range response.OutageList {
		city, ok := defaultCityMap[v.City]
		if !ok {
			logger.Error("city id is not found in city mapper", zap.Any("event", v))
			continue
		}
		loc, err := time.LoadLocation("Asia/Tehran")
		if err != nil {
			logger.Warn("failed to load Asia/Tehran timezone, falling back to UTC", zap.Error(err))
			loc = time.UTC
		}
		date, err := jalali.ParseInLocation("2006/01/02 15:04", v.OutageDate+" "+v.OutageTime, loc)
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
