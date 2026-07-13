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

// startNotifyBeforeScheduler runs a goroutine that scans for upcoming events
// and sends a notification to users NotifyBefore minutes before each event starts.
//
// The scheduler ticks every minute. On each tick and whenever the trigger
// channel receives a signal (e.g. after a collection cycle), it queries the
// database for events whose start time falls within the notify window and
// schedules one-shot callbacks. When a callback fires, matching listeners are
// looked up and a Notification is pushed onto nc.
func startNotifyBeforeScheduler(
	ctx context.Context,
	cfg *config.Config,
	trigger <-chan struct{},
	nc chan<- models.Notification,
) error {
	logger := log.FromContext(ctx).Named("NotifyBeforeScheduler")
	leadTime := cfg.NotifyBefore

	s := &notifyScheduler{
		scheduled: make(map[string]time.Time),
		nc:        nc,
		logger:    logger,
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Initial scan at boot — fire any events that are already due immediately.
	logger.Debug("performing initial scan at boot", zap.Duration("lead_time", leadTime))
	s.scanAndSchedule(ctx, leadTime)
	s.fireDue(ctx)

	for {
		select {
		case <-ticker.C:
			logger.Debug("ticker ticked, firing due and re-scanning")
			s.fireDue(ctx)
			s.scanAndSchedule(ctx, leadTime)
		case <-trigger:
			logger.Debug("collection trigger received, re-scanning and firing due")
			s.scanAndSchedule(ctx, leadTime)
			s.fireDue(ctx)
		case <-ctx.Done():
			logger.Debug("context cancelled, stopping notify-before scheduler")
			return nil
		}
	}
}

// notifyScheduler holds the map of events scheduled for pre-start notification.
type notifyScheduler struct {
	mu        sync.Mutex
	scheduled map[string]time.Time // event hash -> fire time (start - notifyBefore)
	nc        chan<- models.Notification
	logger    *zap.Logger
}

// scanAndSchedule queries the database for events starting within the lead time
// window and registers them in the scheduled map if not already present.
func (s *notifyScheduler) scanAndSchedule(ctx context.Context, leadTime time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	windowStart := now.Add(-leadTime)
	windowEnd := now.Add(leadTime)

	db := database.Get()
	var events []models.Event
	if err := db.Where("start_at >= ? AND start_at <= ?", windowStart, windowEnd).Find(&events).Error; err != nil {
		s.logger.Warn("failed to query upcoming events", zap.Error(err))
		return
	}

	s.logger.Debug("query returned events",
		zap.Int("count", len(events)),
		zap.Time("window_start", windowStart),
		zap.Time("window_end", windowEnd),
	)

	newlyScheduled := 0
	for i := range events {
		ev := events[i]
		if _, exists := s.scheduled[ev.Hash]; exists {
			s.logger.Debug("event already scheduled, skipping",
				zap.String("hash", ev.Hash),
				zap.String("city", ev.City),
				zap.Time("event_start", ev.Start),
			)
			continue
		}
		fireTime := ev.Start.Add(-leadTime)
		if fireTime.Before(now) {
			fireTime = now
		}
		s.scheduled[ev.Hash] = fireTime
		newlyScheduled++
		s.logger.Debug("scheduled pre-start notification",
			zap.String("hash", ev.Hash),
			zap.String("city", ev.City),
			zap.Time("event_start", ev.Start),
			zap.Time("fire_at", fireTime),
			zap.Duration("until_fire", time.Until(fireTime)),
			zap.Int("scheduled_count", len(s.scheduled)),
		)
	}

	if newlyScheduled > 0 {
		s.logger.Debug("new events scheduled for pre-start notification",
			zap.Int("new_count", newlyScheduled),
			zap.Int("total_scheduled", len(s.scheduled)),
		)
	} else {
		s.logger.Debug("no new events to schedule",
			zap.Int("total_scheduled", len(s.scheduled)),
		)
	}
}

// fireDue iterates the scheduled map and dispatches notifications for every
// event whose fire time has arrived.
func (s *notifyScheduler) fireDue(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	db := database.Get()

	fired := 0
	for hash, fireTime := range s.scheduled {
		if now.Before(fireTime) {
			continue
		}

		// Fetch the full event from the database.
		var ev models.Event
		if err := db.Where("hash = ?", hash).First(&ev).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Warn("failed to fetch event for pre-start notification",
					zap.String("hash", hash), zap.Error(err),
				)
			} else {
				s.logger.Debug("event not found in DB, removing from schedule",
					zap.String("hash", hash),
					zap.Time("fire_time", fireTime),
				)
			}
			delete(s.scheduled, hash)
			continue
		}

		delete(s.scheduled, hash)
		fired++
		s.logger.Debug("firing pre-start notification",
			zap.String("hash", hash),
			zap.String("city", ev.City),
			zap.String("address", ev.Address),
			zap.Time("event_start", ev.Start),
			zap.Time("fire_time", fireTime),
		)
		s.dispatchForEvent(ctx, ev)
	}

	if fired > 0 {
		s.logger.Debug("pre-start notifications fired",
			zap.Int("fired_count", fired),
			zap.Int("remaining_scheduled", len(s.scheduled)),
		)
	} else if len(s.scheduled) > 0 {
		s.logger.Debug("no events due yet",
			zap.Int("pending", len(s.scheduled)),
		)
	}
}

// dispatchForEvent looks up all listeners that match the event and sends a
// Notification for each.
func (s *notifyScheduler) dispatchForEvent(ctx context.Context, ev models.Event) {
	listeners := getListeners(database.Get())
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
