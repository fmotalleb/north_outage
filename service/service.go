package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/fmotalleb/go-tools/broadcast"
	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/models"
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
	ec := make(chan models.Event, eventsChannelBufferSize)
	wg := new(sync.WaitGroup)
	wg.Go(
		func() {
			err := web.Start(ctx, cfg)
			if err != nil {
				l.Error("api server collapsed", zap.Error(err))
				panic(fmt.Errorf("api server unrecoverable exception: %w", err))
			}
		},
	)

	bc := broadcast.NewBroadcaster[models.Notification](l)
	wg.Go(
		func() {
			ch := eventToNotificationTransformer(ctx, db, ec)
			bc.BindTo(ch)
		},
	)
	wg.Go(
		func() {
			err := startCollector(ctx, cfg, ec)
			if err != nil {
				l.Error("scheduler service collapsed", zap.Error(err))
				panic(fmt.Errorf("scheduler service unrecoverable exception: %w", err))
			}
		},
	)

	// wg.Go(
	// 	func() {
	// 		_, ch := bc.Subscribe(notificationBufferSize)
	// 		err := telegram.Run(ctx, cfg, ch)
	// 		if err != nil {
	// 			l.Error("telegram service collapsed", zap.Error(err))
	// 			panic(fmt.Errorf("telegram service unrecoverable exception: %w", err))
	// 		}
	// 	},
	// )
	wg.Wait()
	return nil
}

func eventToNotificationTransformer(ctx context.Context, db *gorm.DB, events <-chan models.Event) <-chan models.Notification {
	notifications := make(chan models.Notification, eventsChannelBufferSize)
	db = db.Table("listeners")
	go func() {
		for {
			select {
			case ev := <-events:
				listeners := selectMatchingListeners(db, ev)
				for _, l := range listeners {
					notifications <- models.Notification{
						Listener: &l,
						Event:    &ev,
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return notifications
}

func selectMatchingListeners(db *gorm.DB, e models.Event) []models.Listener {
	listeners := make([]models.Listener, 0)
	db.Where(
		"city = ? AND INSTR(?, search_term) > 0",
		e.City,
		e.Address,
	).Find(&listeners)
	return listeners
}
