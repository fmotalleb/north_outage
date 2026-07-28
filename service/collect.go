package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/fmotalleb/go-tools/log"
	"github.com/robfig/cron/v3"

	"github.com/fmotalleb/north_outage/collector"
	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/models"
)

// startCollector schedules and runs the periodic event collection job.
func startCollector(ctx context.Context, cfg *config.Config, ec chan models.Event) error {
	logger := log.FromContext(ctx).Named("CollectorScheduler")

	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(cfg.CollectCycle)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %s: %w", cfg.CollectCycle, err)
	}

	nextRun := schedule.Next(time.Now())
	timeUntilNext := time.Until(nextRun)

	cron := cron.New(cron.WithParser(parser))

	jobID, err := cron.AddFunc(cfg.CollectCycle, makeCollectFunc(ctx, cfg, ec))
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
		go makeCollectFunc(ctx, cfg, ec)()
	}

	// Start cron scheduler and block until context is canceled.
	go cron.Start()
	<-ctx.Done()
	if innerCtx := cron.Stop(); innerCtx != nil {
		<-innerCtx.Done()
	}
	return nil
}

// makeCollectFunc wraps collector.CollectAndStore in a cron-compatible function.
func makeCollectFunc(ctx context.Context, cfg *config.Config, ec chan models.Event) func() {
	logger := log.FromContext(ctx).Named("CollectorJob")
	return func() {
		logger.Debug("collector cycle started")
		if events, err := collector.CollectAndStore(ctx, cfg); err != nil {
			logger.Error("collector cycle failed", zap.Error(err))
		} else {
			go func() {
				for _, e := range events {
					ec <- e
				}
			}()
		}
	}
}
