package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/fmotalleb/north_outage/internal/template"
	im "github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram/message"
)

func TestParseMuteTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   string
		wantLI uint
		wantEI uint
		wantOK bool
	}{
		{name: "valid", data: "12:34", wantLI: 12, wantEI: 34, wantOK: true},
		{name: "zero listener", data: "0:34", wantOK: false},
		{name: "zero event", data: "12:0", wantOK: false},
		{name: "missing event", data: "12", wantOK: false},
		{name: "extra colon", data: "12:34:56", wantOK: false},
		{name: "empty", data: "", wantOK: false},
		{name: "non numeric", data: "abc:def", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			li, ei, ok := parseMuteTarget(tt.data)
			if ok != tt.wantOK {
				t.Fatalf("parseMuteTarget(%q) ok = %v, want %v", tt.data, ok, tt.wantOK)
			}
			if tt.wantOK && (li != tt.wantLI || ei != tt.wantEI) {
				t.Fatalf("parseMuteTarget(%q) = (%d, %d), want (%d, %d)",
					tt.data, li, ei, tt.wantLI, tt.wantEI)
			}
		})
	}
}

func TestNotificationsTemplate(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 8, 9, 9, 30, 0, 0, time.UTC)
	rows := []notificationRow{
		{
			ListenerID: 1,
			EventID:    2,
			Event: &im.Event{
				City:    "ساری",
				Address: "خیابان فردوسی",
				Start:   start,
				End:     start.Add(2 * time.Hour),
			},
			Message: "دیتای جدید",
			SentAt:  time.Now(),
			Muted:   true,
		},
		{
			ListenerID: 3,
			EventID:    4,
			Event: &im.Event{
				City:    "آمل",
				Address: "بلوار امام",
				Start:   start.Add(24 * time.Hour),
				End:     start.Add(25 * time.Hour),
			},
			Message: "به زودی",
			SentAt:  time.Now(),
		},
	}

	out, err := template.EvaluateTemplate(message.Notifications, map[string]any{
		"notifications": rows,
	})
	if err != nil {
		t.Fatalf("failed to evaluate notifications template: %v", err)
	}
	for _, want := range []string{"ساری", "آمل", "دیتای جدید", "غیرفعال شده"} {
		if !strings.Contains(out, want) {
			t.Fatalf("notifications template output missing %q:\n%s", want, out)
		}
	}
}
