package telegram

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
)

func TestIsPollingTimeout(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "client timeout during getUpdates",
			err: fmt.Errorf("error get updates, %w",
				fmt.Errorf("error do request for method getUpdates, %w",
					&url.Error{Op: "Post", URL: "https://api.telegram.org/bot.../getUpdates", Err: context.DeadlineExceeded},
				),
			),
			want: true,
		},
		{
			name: "deadline exceeded on the request context",
			err: fmt.Errorf("error get updates, %w",
				fmt.Errorf("error do request for method getUpdates, %w", context.DeadlineExceeded),
			),
			want: true,
		},
		{
			name: "non-timeout getUpdates failure",
			err:  fmt.Errorf("error get updates, error do request for method getUpdates: 401 Unauthorized"),
			want: false,
		},
		{
			name: "timeout on a different method",
			err: fmt.Errorf("error do request for method sendMessage, %w",
				&url.Error{Op: "Post", URL: "https://api.telegram.org/bot.../sendMessage", Err: context.DeadlineExceeded},
			),
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("something else"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPollingTimeout(tt.err); got != tt.want {
				t.Fatalf("isPollingTimeout(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
