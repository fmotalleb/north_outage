package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Event struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Hash    string `gorm:"index:idx_event_hash,unique" json:"unique_hash"`
	City    string `gorm:"size:255;not null" json:"city"`
	Address string `gorm:"type:text;not null" json:"address"`

	Start     time.Time `json:"start_at"`
	End       time.Time `json:"end_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (e *Event) ResetHash() {
	data := fmt.Sprintf("%s|%s|%d|%d",
		e.City,
		e.Address,
		e.Start.UnixNano(),
		e.End.UnixNano(),
	)

	hash := sha256.Sum256([]byte(data))
	e.Hash = hex.EncodeToString(hash[:])
}

var clockMapper = map[int]string{
	0:  "🕛",
	1:  "🕐",
	2:  "🕑",
	3:  "🕒",
	4:  "🕓",
	5:  "🕔",
	6:  "🕕",
	7:  "🕖",
	8:  "🕗",
	9:  "🕘",
	10: "🕙",
	11: "🕚",
}

func getClockOf(hour int) string {
	return clockMapper[hour%12]
}

func (e *Event) StartClock() string {
	return getClockOf(e.Start.Hour())
}

func (e *Event) EndClock() string {
	return getClockOf(e.End.Hour())
}
