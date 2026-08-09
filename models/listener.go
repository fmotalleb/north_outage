package models

type Listener struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// ChatID
	TelegramCID int64 `gorm:"not null" json:"-"`
	// ThreadID
	TelegramTID int64 `gorm:"null" json:"-"`

	// Search Text (sql `like`)
	SearchTerm string `gorm:"type:text;not null;uniqueIndex:idx_listener_term_city" json:"search_term"`
	City       string `gorm:"size:255;not null;uniqueIndex:idx_listener_term_city" json:"city"`
}
