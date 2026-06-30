package scrapper

import (
	"context"
	"errors"
	"fmt"

	"github.com/fmotalleb/go-tools/decoder"
	"github.com/fmotalleb/go-tools/log"
	"github.com/fmotalleb/scrapper-go/engine"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/models"
)

// Run executes the configured collector engine and returns a deduplicated list of events.
func Run(ctx context.Context) ([]models.Event, error) {
	logger := log.FromContext(ctx).Named("collector")

	cfg, err := config.Get(ctx)
	if err != nil {
		return nil, err
	}

	rawResult, err := engine.ExecuteConfig(ctx, cfg.CollectorConfig)
	if err != nil {
		logger.Error("collector execution failed", zap.Error(err))
		return nil, err
	}

	logger.Info("collector finished successfully")

	events, reshapeErr := transformResult(rawResult)
	if reshapeErr != nil {
		// Preserve detailed error reporting but keep the main flow working.
		if multiErr, ok := reshapeErr.(interface{ Unwrap() []error }); ok {
			logger.Error("transform produced some errors (ignored)", zap.Errors("errors", multiErr.Unwrap()))
		} else {
			logger.Error("transform produced some errors (ignored)", zap.Error(reshapeErr))
		}
	}

	return events, nil
}

// transformResult converts engine output into []models.Event, deduplicating by Hash.
func transformResult(data map[string]any) ([]models.Event, error) {
	seen := make(map[string]struct{})
	events := make([]models.Event, 0)
	var errs []error

	for mapName, value := range data {
		list, ok := value.([]map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf(
				"invalid data type for %q: expected []map[string]any, got %T",
				mapName, value,
			))
			continue
		}

		rows := make([]map[string]string, len(list))
		if err := decoder.Decode(&rows, list); err != nil {
			errs = append(errs, fmt.Errorf("failed to decode %q: %w", mapName, err))
			continue
		}

		for _, row := range rows {
			ev, ok := normalize(mapName, row)
			if !ok {
				errs = append(errs, fmt.Errorf(
					"failed to normalize entry for %q: %v",
					mapName, row,
				))
				continue
			}

			if _, exists := seen[ev.Hash]; exists {
				continue
			}

			seen[ev.Hash] = struct{}{}
			events = append(events, *ev)
		}
	}

	if len(errs) > 0 {
		return events, errors.Join(errs...)
	}

	return events, nil
}
