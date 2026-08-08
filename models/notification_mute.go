package models

import "time"

// NotificationMute records a DND-style mute for a specific listener+event
// pair. While a row exists for a pair, no notification (new data or upcoming)
// is delivered for that pair.
type NotificationMute struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ListenerID uint      `gorm:"not null;uniqueIndex:idx_mute_listener_event" json:"listener_id"`
	EventID    uint      `gorm:"not null;uniqueIndex:idx_mute_listener_event" json:"event_id"`
	MutedAt    time.Time `gorm:"not null" json:"muted_at"`

	// Relationships (preloaded on retrieval)
	Listener *Listener `gorm:"foreignKey:ListenerID;constraint:OnDelete:CASCADE" json:"listener,omitempty"`
	Event    *Event    `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE" json:"event,omitempty"`
}
