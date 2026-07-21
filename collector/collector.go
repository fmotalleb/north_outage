package collector

import (
	"context"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/models"
)

func Collect(ctx context.Context) ([]models.Event, error) {
	l := log.FromContext(ctx).Named("collector")
	ctx = log.WithLogger(ctx, l)
	l.Info("starting to collect")
	events, err := fetchData(ctx)
	if err != nil {
		l.Error("failed to fetch outage data", zap.Error(err))
		return nil, err
	}
	l.Info("finished collecting", zap.Int("items", len(events)))
	return events, nil
}
