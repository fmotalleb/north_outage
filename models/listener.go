package models

type Listener struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// ChatID
	TelegramCID int64 `gorm:"not null" json:"-"`
	// ThreadID
	TelegramTID int64 `gorm:"null" json:"-"`

	SearchTerm string `gorm:"type:text;not null" json:"search_term"`
	City       string `gorm:"size:255;not null" json:"city"`
}
