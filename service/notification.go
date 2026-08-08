package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/fmotalleb/go-scheduler"
	"github.com/fmotalleb/go-scheduler/worker"
	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/models"
)

// NotificationStorage persists scheduled notifications via GORM so they
// survive service restarts.
type NotificationStorage struct {
	db *gorm.DB
}

// Add stores a notification to be sent at (or after) the given time.
// Returns the auto-increment ID assigned by the database.
func (n *NotificationStorage) Add(when time.Time, task models.Notification) (int, error) {
	sn := models.ScheduledNotification{
		ScheduledAt: when,
		ListenerID:  task.Listener.ID,
		EventID:     task.Event.ID,
		Message:     task.Message,
	}
	if err := n.db.Create(&sn).Error; err != nil {
		return 0, err
	}
	return int(sn.ID), nil
}

// Remove deletes a previously stored notification by its ID and returns
// the original Notification value. Returns an error if the ID is not found.
func (n *NotificationStorage) Remove(id int) (models.Notification, error) {
	var zero models.Notification

	var sn models.ScheduledNotification
	if err := n.db.First(&sn, id).Error; err != nil {
		return zero, err
	}

	if err := n.db.Delete(&sn).Error; err != nil {
		return zero, err
	}

	return models.Notification{
		Listener: &models.Listener{ID: sn.ListenerID},
		Event:    &models.Event{ID: sn.EventID},
		Message:  sn.Message,
	}, nil
}

// PopBefore retrieves and removes all notifications scheduled strictly
// before the given time. The returned values have their Listener and
// Event relationships preloaded.
func (n *NotificationStorage) PopBefore(deadline time.Time) ([]models.Notification, error) {
	var items []models.ScheduledNotification
	if err := n.db.
		Where("scheduled_at < ?", deadline).
		Order("id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	// Collect IDs for deletion.
	ids := make([]uint, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}

	// Delete atomically in the same transaction context.
	if err := n.db.Delete(&models.ScheduledNotification{}, "id IN ?", ids).Error; err != nil {
		return nil, err
	}

	// Convert to notifications with preloaded relationships.
	result := make([]models.Notification, len(items))
	for i, item := range items {
		listener := new(models.Listener)
		if err := n.db.First(listener, item.ListenerID).Error; err != nil {
			return nil, err
		}

		event := new(models.Event)
		if err := n.db.First(event, item.EventID).Error; err != nil {
			return nil, err
		}

		result[i] = models.Notification{
			Listener: listener,
			Event:    event,
			Message:  item.Message,
		}
	}

	return result, nil
}

// DeleteBefore removes all stored notifications scheduled strictly before
// the given time. Used once at boot to forget notifications whose send time
// elapsed while the service was down.
func (n *NotificationStorage) DeleteBefore(deadline time.Time) error {
	return n.db.Where("scheduled_at < ?", deadline).Delete(&models.ScheduledNotification{}).Error
}

// Close is a no-op — the GORM connection is managed externally and reused
// across the application.
func (n *NotificationStorage) Close() {}

func eventToNotificationTransformer(ctx context.Context, db *gorm.DB, events <-chan models.Event, notifications chan models.Notification) {
	ctx, l := log.AsNamedChild(ctx, "NotificationTransformer")
	cfg, _ := config.Get(ctx)
	// Keep a full-featured db handle for mute lookups; `db` is reassigned to
	// the listeners table below for the cached listener queries.
	muteDB := db
	worker := worker.NewWorkerPool(
		context.WithoutCancel(ctx),
		func(ctx context.Context, n models.Notification) {
			if isMuted(muteDB, n.Listener.ID, n.Event.ID) {
				l.Debug("skipping muted notification",
					zap.Uint("listener_id", n.Listener.ID),
					zap.Uint("event_id", n.Event.ID),
				)
				return
			}
			select {
			case <-ctx.Done():
			case notifications <- n:
			}
		},
		4,
		100,
	)
	ns := &NotificationStorage{
		db,
	}
	// Forget scheduled notifications whose send time passed while the
	// service was down, so the scheduler does not fire them on startup.
	if err := ns.DeleteBefore(time.Now()); err != nil {
		l.Error("failed to purge expired scheduled notifications", zap.Error(err))
	}
	storage := scheduler.WithStorage(ns)
	sc := scheduler.New(
		ctx,
		worker,
		storage,
		scheduler.WithTickerCycle[models.Notification](time.Minute),
		scheduler.WithLogger[models.Notification](func(err error) {
			l.Error("scheduler error", zap.Error(err))
		}),
	)

	db = db.Table("listeners")

	// Close the notifications channel only after sc.Close has drained the
	// worker pool, otherwise a worker could send on a closed channel.
	// Registered first so it runs last (defers run LIFO).
	defer close(notifications)
	defer sc.Close()
	for {
		select {
		case ev := <-events:
			for _, li := range getListeners(db) {
				if ev.City != li.City {
					continue
				}
				if strings.Contains(ev.Address, li.SearchTerm) {
					if isMuted(muteDB, li.ID, ev.ID) {
						l.Debug("skipping muted event for listener",
							zap.Uint("listener_id", li.ID),
							zap.Uint("event_id", ev.ID),
						)
						continue
					}
					select {
					case <-ctx.Done():
						return
					case notifications <- models.Notification{
						Listener: &li,
						Event:    &ev,
						Message:  models.NotificationNewData,
					}:
					}
					sc.Add(ev.Start.Add(cfg.NotifyBefore*-1), models.Notification{
						Listener: &li,
						Event:    &ev,
						Message:  models.NotificationUpcoming,
					})
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

var (
	listenersCacheMu    sync.Mutex
	listenersCache      []models.Listener
	listenersCacheExpAt time.Time
)

// isMuted reports whether the given listener+event pair has been muted
// (DND) by the user.
func isMuted(db *gorm.DB, listenerID, eventID uint) bool {
	var count int64
	db.Model(&models.NotificationMute{}).
		Where("listener_id = ? AND event_id = ?", listenerID, eventID).
		Count(&count)
	return count > 0
}

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
