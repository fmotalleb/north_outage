package template

import (
	"testing"
	"time"
)

func TestRelTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		t    time.Time
		want string
	}{
		{"72m future", now.Add(72*time.Minute + time.Second), "۱ ساعت و ۱۲ دقیقه دیگه"},
		{"30m future", now.Add(30*time.Minute + time.Second), "۳۰ دقیقه دیگه"},
		{"1h past", now.Add(-1*time.Hour - time.Second), "۱ ساعت پیش"},
		{"1h5m future", now.Add(65*time.Minute + time.Second), "۱ ساعت و ۵ دقیقه دیگه"},
		{"2d future", now.Add(2*24*time.Hour + time.Second), "۲ روز دیگه"},
		{"2d3h future", now.Add(51*time.Hour + time.Second), "۲ روز و ۳ ساعت دیگه"},
		{"30s future", now.Add(30 * time.Second), "لحظاتی دیگه"},
		{"30s past", now.Add(-30 * time.Second), "لحظاتی پیش"},
		{"exact 1h", now.Add(60*time.Minute + time.Second), "۱ ساعت دیگه"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := relTime(c.t)
			if got != c.want {
				t.Fatalf("relTime(%s) = %q, want %q", c.name, got, c.want)
			}
		})
	}
}
