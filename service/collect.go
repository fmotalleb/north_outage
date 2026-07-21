package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/fmotalleb/go-tools/log"
	"github.com/robfig/cron/v3"

	"github.com/fmotalleb/north_outage/collector"
	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/models"
)

// StartCollector schedules and runs the periodic event collection job.
// The notifyTrigger channel is signaled after every successful collection so
// the notify-before scheduler can scan for newly arrived events.
func startCollector(ctx context.Context, cfg *config.Config, ec chan models.Event, notifyTrigger chan<- struct{}) error {
	logger := log.FromContext(ctx).Named("CollectorScheduler")

	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(cfg.CollectCycle)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %s: %w", cfg.CollectCycle, err)
	}

	nextRun := schedule.Next(time.Now())
	timeUntilNext := time.Until(nextRun)

	scheduler := cron.New(cron.WithParser(parser))

	jobID, err := scheduler.AddFunc(cfg.CollectCycle, makeCollectFunc(ctx, cfg, ec, notifyTrigger))
	if err != nil {
		logger.Error("failed to register collector job", zap.Error(err))
		return err
	}

	logger.Info("collector job scheduled",
		zap.Int("jobID", int(jobID)),
		zap.Time("nextRun", nextRun),
		zap.Duration("timeUntilNext", timeUntilNext),
	)

	// Trigger immediate collection if threshold is met.
	if *cfg.CollectOnStart && timeUntilNext > cfg.CollectOnStartThreshold {
		logger.Info("triggering immediate collection on start",
			zap.Duration("threshold", cfg.CollectOnStartThreshold),
			zap.Duration("timeUntilNext", timeUntilNext),
		)
		go makeCollectFunc(ctx, cfg, ec, notifyTrigger)()
	}

	// Start cron scheduler and block until context is canceled.
	go scheduler.Start()
	<-ctx.Done()
	if innerCtx := scheduler.Stop(); innerCtx != nil {
		<-innerCtx.Done()
	}
	return nil
}

// makeCollectFunc wraps collectAndStore in a cron-compatible function.
// After a successful collection it also signals the notifyTrigger so that the
// notify-before scheduler re-scans the database for newly stored events.
func makeCollectFunc(ctx context.Context, cfg *config.Config, ec chan models.Event, notifyTrigger chan<- struct{}) func() {
	logger := log.FromContext(ctx).Named("CollectorJob")
	return func() {
		logger.Debug("collector cycle started")
		if events, err := collectAndStore(ctx, cfg); err != nil {
			logger.Error("collector cycle failed", zap.Error(err))
		} else {
			go func() {
				for _, e := range events {
					ec <- e
				}
				// Signal the notify-before scheduler after all new events are published.
				select {
				case notifyTrigger <- struct{}{}:
				default:
				}
			}()
		}
	}
}

// collectAndStore runs the collector, deduplicates results, persists new events, and triggers GC.
func collectAndStore(ctx context.Context, cfg *config.Config) ([]models.Event, error) {
	logger := log.FromContext(ctx).Named("CollectorCycle")
	ctx, cancel := context.WithTimeout(ctx, cfg.CollectTimeout)
	defer cancel()

	db := database.Get()
	logger.Info("starting data collection")

	events, err := collector.Collect(ctx)
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

	if gcErr := runEventsGC(cfg.RotateAfter, db); gcErr != nil {
		logger.Warn("event garbage collection failed", zap.Error(gcErr))
	}
	return newEvents, nil
}

// runEventsGC removes expired events from the database based on maxAge.
func runEventsGC(maxAge time.Duration, db *gorm.DB) error {
	cutoff := time.Now().UTC().Add(-maxAge)

	return db.
		Where(`"end_at" <= ?`, cutoff).
		Unscoped().
		Delete(&models.Event{}).
		Error
}
