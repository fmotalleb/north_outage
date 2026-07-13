package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fmotalleb/go-tools/broadcast"
	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

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
	l := log.FromContext(ctx).Named("Serve")
	cfg, err := config.Get(ctx)
	if err != nil {
		return err
	}
	ctx = config.Attach(ctx, cfg)
	db, err := database.Connect(cfg.DatabaseConnection)
	if err != nil {
		return err
	}
	if err = db.AutoMigrate(&models.Listener{}, &models.Event{}); err != nil {
		return err
	}
	l.Info("config initialized", zap.Any("cfg", cfg))
	weather.Init(cfg.Weather.Proxy)
	mattermost.Setup(ctx, cfg)
	ec := make(chan models.Event, eventsChannelBufferSize)
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(
		func() error {
			err := web.Start(ctx, cfg)
			if err != nil {
				l.Error("api server collapsed", zap.Error(err))
				return fmt.Errorf("api server unrecoverable exception: %w", err)
			}
			return nil
		},
	)

	bc := broadcast.NewBroadcaster[models.Notification](l)
	wg.Go(
		func() error {
			ch := eventToNotificationTransformer(ctx, db, ec)
			bc.BindTo(ch)
			return nil
		},
	)

	// notifyTrigger signals the notify-before scheduler to re-scan the database
	// after each successful collection cycle.
	notifyTrigger := make(chan struct{}, 1)

	wg.Go(
		func() error {
			err := startCollector(ctx, cfg, ec, notifyTrigger)
			if err != nil {
				l.Error("scheduler service collapsed", zap.Error(err))
				return fmt.Errorf("scheduler service unrecoverable exception: %w", err)
			}
			return nil
		},
	)

	wg.Go(
		func() error {
			// Create a dedicated channel for pre-start notifications and register it
			// with the broadcaster so messages are forwarded to Telegram / Mattermost.
			preNotifyCh := make(chan models.Notification, notificationBufferSize)
			bc.BindTo(preNotifyCh)
			err := startNotifyBeforeScheduler(ctx, cfg, notifyTrigger, preNotifyCh)
			if err != nil {
				l.Error("notify-before scheduler collapsed", zap.Error(err))
				return fmt.Errorf("notify-before scheduler unrecoverable exception: %w", err)
			}
			return nil
		},
	)
	if cfg.Telegram.BotKey != "" {
		wg.Go(
			func() error {
				_, ch := bc.Subscribe(notificationBufferSize)
				err := telegram.Run(ctx, cfg, ch)
				if err != nil {
					l.Error("telegram service collapsed", zap.Error(err))
					return fmt.Errorf("telegram service unrecoverable exception: %w", err)
				}
				return nil
			},
		)
	}
	if cfg.Mattermost.BotToken != "" && cfg.Mattermost.ServerURL != "" {
		wg.Go(
			func() error {
				_, ch := bc.Subscribe(notificationBufferSize)
				err := mattermost.Run(ctx, cfg, ch)
				if err != nil {
					l.Error("mattermost service collapsed", zap.Error(err))
					return fmt.Errorf("mattermost service unrecoverable exception: %w", err)
				}
				return nil
			},
		)
	}

	return wg.Wait()
}

func eventToNotificationTransformer(ctx context.Context, db *gorm.DB, events <-chan models.Event) <-chan models.Notification {
	notifications := make(chan models.Notification, eventsChannelBufferSize)
	db = db.Table("listeners")
	go func() {
		for {
			select {
			case ev := <-events:
				for _, l := range getListeners(db) {
					if ev.City != l.City {
						continue
					}
					if strings.Contains(ev.Address, l.SearchTerm) {
						notifications <- models.Notification{
							Listener: &l,
							Event:    &ev,
						}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return notifications
}

var (
	listenersCacheMu    sync.Mutex
	listenersCache      []models.Listener
	listenersCacheExpAt time.Time
)

func getListeners(db *gorm.DB) []models.Listener {
	listenersCacheMu.Lock()
	defer listenersCacheMu.Unlock()

	now := time.Now()
	if listenersCache != nil && now.Before(listenersCacheExpAt) {
		return listenersCache
	}

	listeners := make([]models.Listener, 0)
	db.Find(&listeners)
	listenersCache = listeners
	listenersCacheExpAt = now.Add(10 * time.Second)
	return listeners
}
