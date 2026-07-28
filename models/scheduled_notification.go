package models

import "time"

// ScheduledNotification persists pending scheduled notifications in the
// database so they survive service restarts. The scheduler's Storage
// implementation reads/writes this table.
type ScheduledNotification struct {
	ID          uint      `gorm:"primaryKey"`
	ScheduledAt time.Time `gorm:"not null;index"`
	ListenerID  uint      `gorm:"not null"`
	EventID     uint      `gorm:"not null"`
	Message     string    `gorm:"type:text;not null"`
	CreatedAt   time.Time
}
