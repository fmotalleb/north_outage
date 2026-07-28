package models

import "time"

type Notification struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Foreign keys to Listener and Event
	ListenerID uint `gorm:"not null;index" json:"listener_id"`
	EventID    uint `gorm:"not null;index" json:"event_id"`

	// Notification payload
	Message string `gorm:"type:text;not null" json:"message"`

	// When the notification was sent
	SentAt time.Time `gorm:"not null" json:"sent_at"`

	// GORM auto-managed timestamp
	CreatedAt time.Time `json:"created_at"`

	// Relationships (preloaded on retrieval)
	Listener *Listener `gorm:"foreignKey:ListenerID;constraint:OnDelete:CASCADE" json:"listener,omitempty"`
	Event    *Event    `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE" json:"event,omitempty"`
}
