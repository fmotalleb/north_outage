package collector

import (
	"context"

	"github.com/fmotalleb/go-tools/log"

	"github.com/fmotalleb/north_outage/collector/scrapper"
	"github.com/fmotalleb/north_outage/models"
)

func Collect(ctx context.Context) ([]models.Event, error) {
	l := log.FromContext(ctx).Named("collector")
	ctx = log.WithLogger(ctx, l)
	l.Info("starting to collect")
	events, err := scrapper.Run(ctx)
	if err != nil {
		return nil, err
	}
	return events, nil
}
