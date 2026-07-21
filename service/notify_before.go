package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/models"
)

// startNotifyBeforeScheduler listens for collection triggers and schedules
// one-shot timers for each event that needs a pre-start notification.
//
// When a trigger is received (after a collection cycle), it scans the database
// for events starting within the [now, now + 4*leadTime] window and creates a
// one-shot time.Timer for each new event. The timer fires at (event.Start -
// leadTime), dispatches notifications to matching listeners, and is then
// automatically removed from the schedule — no polling required.
func startNotifyBeforeScheduler(
	ctx context.Context,
	cfg *config.Config,
	trigger <-chan struct{},
	nc chan<- models.Notification,
) error {
	logger := log.FromContext(ctx).Named("NotifyBeforeScheduler")
	leadTime := cfg.NotifyBefore

	s := &notifyScheduler{
		timers:   make(map[string]*time.Timer),
		nc:       nc,
		leadTime: leadTime,
		logger:   logger,
	}

	// Initial scan at boot — schedule any events already in the window.
	logger.Debug("performing initial scan at boot", zap.Duration("lead_time", leadTime))
	s.scanAndSchedule(ctx)

	for {
		select {
		case <-trigger:
			logger.Debug("collection trigger received, scheduling events")
			s.scanAndSchedule(ctx)
		case <-ctx.Done():
			// Stop all pending timers on shutdown.
			s.mu.Lock()
			for hash, t := range s.timers {
				if t != nil {
					t.Stop()
				}
				delete(s.timers, hash)
			}
			s.mu.Unlock()
			logger.Debug("context canceled, stopped all pending notification timers")
			return nil
		}
	}
}

// notifyScheduler holds one-shot timers for pre-start notifications.
type notifyScheduler struct {
	mu       sync.Mutex
	timers   map[string]*time.Timer // event hash -> one-shot timer (nil after fire)
	nc       chan<- models.Notification
	leadTime time.Duration
	logger   *zap.Logger
}

// scanAndSchedule queries the database for events starting within the lead-time
// window and schedules a one-shot timer for each new event.
func (s *notifyScheduler) scanAndSchedule(ctx context.Context) {
	now := time.Now()
	windowEnd := now.Add(4 * s.leadTime)

	db := database.Get()
	var events []models.Event
	if err := db.Where("start_at >= ? AND start_at <= ?", now, windowEnd).Find(&events).Error; err != nil {
		s.logger.Warn("failed to query upcoming events", zap.Error(err))
		return
	}

	s.logger.Debug("query returned events",
		zap.Int("count", len(events)),
		zap.Time("now", now),
		zap.Time("window_end", windowEnd),
	)

	newlyScheduled := 0
	for i := range events {
		ev := events[i]

		s.mu.Lock()
		if _, exists := s.timers[ev.Hash]; exists {
			s.mu.Unlock()
			s.logger.Debug("event already scheduled, skipping",
				zap.String("hash", ev.Hash),
				zap.String("city", ev.City),
			)
			continue
		}

		fireTime := ev.Start.Add(-s.leadTime)

		if fireTime.Before(now) {
			// Already due — dispatch immediately without a timer.
			s.timers[ev.Hash] = nil // mark as handled
			s.mu.Unlock()
			s.logger.Debug("event already due, dispatching immediately",
				zap.String("hash", ev.Hash),
				zap.String("city", ev.City),
				zap.Time("event_start", ev.Start),
			)
			s.dispatchForEvent(ctx, ev)
			// Clean up the marker — no timer to track.
			s.mu.Lock()
			delete(s.timers, ev.Hash)
			s.mu.Unlock()
			continue
		}

		// Schedule a one-shot timer. Capture ev by copy for the closure.
		event := ev
		timer := time.AfterFunc(time.Until(fireTime), func() {
			s.dispatchForEvent(ctx, event)
			s.mu.Lock()
			delete(s.timers, event.Hash)
			s.mu.Unlock()
			s.logger.Debug("notification timer fired and cleaned up",
				zap.String("hash", event.Hash),
				zap.String("city", event.City),
				zap.Time("event_start", event.Start),
			)
		})
		s.timers[ev.Hash] = timer
		s.mu.Unlock()
		newlyScheduled++

		s.logger.Debug("scheduled pre-start notification",
			zap.String("hash", ev.Hash),
			zap.String("city", ev.City),
			zap.Time("event_start", ev.Start),
			zap.Time("fire_at", fireTime),
			zap.Duration("until_fire", time.Until(fireTime)),
			zap.Int("total_scheduled", len(s.timers)),
		)
	}

	if newlyScheduled > 0 {
		s.logger.Debug("new events scheduled for pre-start notification",
			zap.Int("new_count", newlyScheduled),
			zap.Int("total_scheduled", len(s.timers)),
		)
	} else {
		s.logger.Debug("no new events to schedule",
			zap.Int("total_scheduled", len(s.timers)),
		)
	}
}

// dispatchForEvent looks up all listeners that match the event and sends a
// Notification for each.
func (s *notifyScheduler) dispatchForEvent(ctx context.Context, ev models.Event) {
	db := database.Get()
	// Fetch the latest event from the database to ensure it still exists.
	var fresh models.Event
	if err := db.Where("hash = ?", ev.Hash).First(&fresh).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("failed to fetch event for pre-start notification",
				zap.String("hash", ev.Hash), zap.Error(err),
			)
		} else {
			s.logger.Debug("event not found in DB, skipping notification",
				zap.String("hash", ev.Hash),
			)
		}
		return
	}
	ev = fresh

	listeners := getListeners(db)
	matched := 0
	for i := range listeners {
		l := listeners[i]
		if ev.City != l.City || !strings.Contains(ev.Address, l.SearchTerm) {
			continue
		}
		matched++
		s.logger.Debug("dispatching notification to listener",
			zap.String("event_hash", ev.Hash),
			zap.Uint("listener_id", l.ID),
			zap.String("city", l.City),
			zap.String("search_term", l.SearchTerm),
			zap.Int64("telegram_cid", l.TelegramCID),
			zap.String("mattermost_cid", l.MattermostCID),
		)
		select {
		case s.nc <- models.Notification{Listener: &l, Event: &ev}:
		case <-ctx.Done():
			return
		}
	}

	if matched > 0 {
		s.logger.Debug("dispatched notifications for event",
			zap.String("hash", ev.Hash),
			zap.Int("matched_listeners", matched),
		)
	} else {
		s.logger.Debug("no matching listeners found for event",
			zap.String("hash", ev.Hash),
			zap.String("city", ev.City),
		)
	}
}
