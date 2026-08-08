package autodelete

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-telegram/bot"

	"github.com/fmotalleb/north_outage/config"
)

func TestScheduleDeletesAfterTTL(t *testing.T) {
	var (
		mu    sync.Mutex
		calls []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		mu.Lock()
		calls = append(calls, r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":true}`))
	}))
	defer srv.Close()

	b, err := bot.New("test-token",
		bot.WithServerURL(srv.URL),
		bot.WithSkipGetMe(),
	)
	if err != nil {
		t.Fatalf("failed to create bot: %v", err)
	}

	ctx := context.Background()
	cfg := &config.Config{}
	cfg.Telegram.MessageTTL = 150 * time.Millisecond
	ctx = config.Attach(ctx, cfg)

	Init(ctx)

	Schedule(ctx, b, 123, 456)

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(calls)
		mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calls) == 0 {
		t.Fatal("expected a deleteMessage API call after TTL")
	}
	if !strings.HasSuffix(calls[0], "/deleteMessage") {
		t.Fatalf("unexpected API call: %s", calls[0])
	}
}

// TestScheduleGuards ensures Schedule is a no-op before Init or without a
// valid message identity.
func TestScheduleGuards(t *testing.T) {
	Schedule(context.Background(), nil, 123, 456)

	Init(context.Background())

	Schedule(context.Background(), nil, 0, 5)
	Schedule(context.Background(), nil, 5, 0)
}
