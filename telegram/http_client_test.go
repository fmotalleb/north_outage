package telegram

import (
	"net/http"
	"net/url"
	"testing"
)

func TestIsGetUpdatesRequest(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "getUpdates polling call",
			url:  "https://api.telegram.org/bot123:ABC/getUpdates",
			want: true,
		},
		{
			name: "sendMessage call",
			url:  "https://api.telegram.org/bot123:ABC/sendMessage",
			want: false,
		},
		{
			name: "getMe call",
			url:  "https://api.telegram.org/bot123:ABC/getMe",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("url.Parse(%q): %v", tt.url, err)
			}
			req := &http.Request{URL: u}
			if got := isGetUpdatesRequest(req); got != tt.want {
				t.Fatalf("isGetUpdatesRequest(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
