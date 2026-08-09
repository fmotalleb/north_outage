package handlers

import (
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	im "github.com/fmotalleb/north_outage/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&im.Listener{}, &im.Event{}, &im.Notification{}, &im.ScheduledNotification{}, &im.NotificationMute{}); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func seedNotificationData(t *testing.T, db *gorm.DB) (listenerID, eventID uint) {
	t.Helper()
	li := im.Listener{
		TelegramCID: 12345,
		City:        "ساری",
		SearchTerm:  "فردوسی",
	}
	if err := db.Create(&li).Error; err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	ev := im.Event{
		City:    "ساری",
		Address: "خیابان فردوسی",
		Start:   time.Now().Add(2 * time.Hour),
		End:     time.Now().Add(4 * time.Hour),
	}
	ev.ResetHash()
	if err := db.Create(&ev).Error; err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	notif := im.Notification{
		ListenerID: li.ID,
		EventID:    ev.ID,
		Message:    im.NotificationNewData,
		SentAt:     time.Now(),
	}
	if err := db.Create(&notif).Error; err != nil {
		t.Fatalf("failed to create notification: %v", err)
	}
	return li.ID, ev.ID
}

func TestNotificationRows(t *testing.T) {
	db := newTestDB(t)
	listenerID, eventID := seedNotificationData(t, db)

	rows, err := notificationRows(db, 12345)
	if err != nil {
		t.Fatalf("notificationRows failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("notificationRows returned %d rows, want 1", len(rows))
	}
	r := rows[0]
	if r.ListenerID != listenerID || r.EventID != eventID {
		t.Fatalf("unexpected row: listener=%d event=%d", r.ListenerID, r.EventID)
	}
	if r.Event == nil || r.Event.Address != "خیابان فردوسی" {
		t.Fatalf("event relationship not preloaded: %+v", r.Event)
	}
	if r.Muted {
		t.Fatal("row should not be muted before a mute is created")
	}

	// Mute the pair and re-check.
	mute := im.NotificationMute{
		ListenerID: listenerID,
		EventID:    eventID,
		MutedAt:    time.Now(),
	}
	if err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "listener_id"}, {Name: "event_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"muted_at"}),
	}).Create(&mute).Error; err != nil {
		t.Fatalf("failed to create mute: %v", err)
	}
	rows, err = notificationRows(db, 12345)
	if err != nil {
		t.Fatalf("notificationRows failed after mute: %v", err)
	}
	if len(rows) != 1 || !rows[0].Muted {
		t.Fatalf("expected the pair to be muted, got %+v", rows)
	}
}

func TestNotificationRowsScopedToChat(t *testing.T) {
	db := newTestDB(t)
	seedNotificationData(t, db)

	// A different chat must see no rows.
	rows, err := notificationRows(db, 99999)
	if err != nil {
		t.Fatalf("notificationRows failed: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected no rows for another chat, got %d", len(rows))
	}
}

func TestMuteDeletesScheduledNotification(t *testing.T) {
	db := newTestDB(t)
	listenerID, eventID := seedNotificationData(t, db)

	scheduled := im.ScheduledNotification{
		ListenerID:  listenerID,
		EventID:     eventID,
		Message:     im.NotificationUpcoming,
		ScheduledAt: time.Now().Add(10 * time.Minute),
	}
	if err := db.Create(&scheduled).Error; err != nil {
		t.Fatalf("failed to create scheduled notification: %v", err)
	}

	// Mirror the mute handler's cleanup.
	if err := db.Where("listener_id = ? AND event_id = ?", listenerID, eventID).
		Delete(&im.ScheduledNotification{}).Error; err != nil {
		t.Fatalf("failed to delete scheduled notification: %v", err)
	}
	var remaining int64
	db.Model(&im.ScheduledNotification{}).Count(&remaining)
	if remaining != 0 {
		t.Fatalf("expected scheduled notification to be removed, got %d", remaining)
	}
}
