package service

import (
	"context"
	"fmt"

	"github.com/fmotalleb/go-tools/broadcast"
	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/mattermost"
	"github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram"
	"github.com/fmotalleb/north_outage/weather"
	"github.com/fmotalleb/north_outage/web"
)

const (
	eventsChannelBufferSize = 300
	notificationBufferSize  = 30
)

func Serve(ctx context.Context) error {
	ctx, l := log.AsNamedChild(ctx, "Serve")
	cfg, err := config.Get(ctx)
	if err != nil {
		return err
	}
	ctx = config.Attach(ctx, cfg)
	db, err := database.Connect(cfg.DatabaseConnection)
	if err != nil {
		return err
	}
	if err = db.AutoMigrate(&models.Listener{}, &models.Event{}, &models.Notification{}, &models.ScheduledNotification{}, &models.NotificationMute{}); err != nil {
		return err
	}
	l.Info("config initialized", zap.Any("cfg", cfg))
	weather.Init(cfg.Weather.Proxy)
	mattermost.Setup(ctx, cfg)
	ec := make(chan models.Event, eventsChannelBufferSize)

	// The shared context passed to every service is the signal-derived
	// context created at the entry point. No service derives a cancellable
	// context of its own, so a failing service can never cancel the context
	// used by the others. Fatal service errors are reported through errCh
	// and exit the process, but they never cancel the shared context.
	wg := new(errgroup.Group)
	errCh := make(chan error, 1)
	run := func(fn func() error) {
		wg.Go(func() error {
			if err := fn(); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return err
			}
			return nil
		})
	}

	run(func() error {
		err := web.Start(ctx, cfg)
		if err != nil {
			l.Error("api server collapsed", zap.Error(err))
			return fmt.Errorf("api server unrecoverable exception: %w", err)
		}
		return nil
	})

	bc := broadcast.NewBroadcaster[models.Notification](l)

	run(func() error {
		notifications := make(chan models.Notification, eventsChannelBufferSize)
		go eventToNotificationTransformer(ctx, db, ec, notifications)
		bc.BindTo(notifications)
		return nil
	})

	run(func() error {
		err := startCollector(ctx, cfg, ec)
		if err != nil {
			l.Error("scheduler service collapsed", zap.Error(err))
			return fmt.Errorf("scheduler service unrecoverable exception: %w", err)
		}
		return nil
	})

	if cfg.Telegram.BotKey != "" {
		run(func() error {
			_, ch := bc.Subscribe(notificationBufferSize)
			err := telegram.Run(ctx, cfg, ch)
			if err != nil {
				l.Error("telegram service collapsed", zap.Error(err))
				return fmt.Errorf("telegram service unrecoverable exception: %w", err)
			}
			return nil
		})
	}
	if cfg.Mattermost.BotToken != "" && cfg.Mattermost.ServerURL != "" {
		run(func() error {
			_, ch := bc.Subscribe(notificationBufferSize)
			err := mattermost.Run(ctx, cfg, ch)
			if err != nil {
				l.Error("mattermost service collapsed", zap.Error(err))
				return fmt.Errorf("mattermost service unrecoverable exception: %w", err)
			}
			return nil
		})
	}

	select {
	case <-ctx.Done():
		l.Info("received shutdown signal")
		return wg.Wait()
	case err := <-errCh:
		l.Error("a service collapsed, shutting down", zap.Error(err))
		return err
	}
}
