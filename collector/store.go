package collector

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/fmotalleb/go-tools/log"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/models"
)

// CollectAndStore runs the collector, deduplicates results, persists new
// events, and runs garbage collection on expired events.
func CollectAndStore(ctx context.Context, cfg *config.Config) ([]models.Event, error) {
	logger := log.FromContext(ctx).Named("CollectorCycle")
	ctx, cancel := context.WithTimeout(ctx, cfg.CollectTimeout)
	defer cancel()

	db := database.Get()
	logger.Info("starting data collection")

	events, err := Collect(ctx)
	if err != nil {
		return nil, fmt.Errorf("collector failed: %w", err)
	}

	// Deduplicate against database hashes.
	existingHashes := make(map[string]struct{})
	var storedHashes []string
	db.Table("events").Select("hash").Find(&storedHashes)
	for _, h := range storedHashes {
		existingHashes[h] = struct{}{}
	}
	newEvents := make([]models.Event, 0)
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, ev := range events {
			if _, exists := existingHashes[ev.Hash]; exists {
				continue
			}
			if err = tx.Create(&ev).Error; err != nil {
				return err
			}
			newEvents = append(newEvents, ev)
			existingHashes[ev.Hash] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist events: %w", err)
	}

	if gcErr := RunEventsGC(cfg.RotateAfter, db); gcErr != nil {
		logger.Warn("event garbage collection failed", zap.Error(gcErr))
	}
	return newEvents, nil
}

// RunEventsGC removes expired events from the database where end_at is
// older than the given maxAge.
func RunEventsGC(maxAge time.Duration, db *gorm.DB) error {
	cutoff := time.Now().UTC().Add(-maxAge)

	return db.
		Where(`"end_at" <= ?`, cutoff).
		Unscoped().
		Delete(&models.Event{}).
		Error
}
