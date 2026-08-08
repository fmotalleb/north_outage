// Package autodelete removes Telegram messages (search results, lists,
// confirmations) after a configurable time. Deletions are scheduled through
// a shared go-scheduler instance (default callback mode); the scheduler
// itself is the task manager, so no additional bookkeeping is needed here.
package autodelete

import (
	"context"
	"time"

	"github.com/fmotalleb/go-scheduler"
	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/config"
)

const defaultTTL = time.Minute

// sched is the shared scheduler. It owns all scheduled deletions.
var sched *scheduler.Scheduler[scheduler.Callback]

// Init creates the shared scheduler using the default callback option. It is
// safe to call more than once; only the first call takes effect.
func Init(ctx context.Context) {
	if sched != nil {
		return
	}
	sched = scheduler.NewCallback(ctx)
}

// Schedule deletes the given message after the configured TTL.
func Schedule(ctx context.Context, b *bot.Bot, chatID int64, messageID int) {
	if sched == nil || b == nil || chatID == 0 || messageID == 0 {
		return
	}
	ttl := defaultTTL
	if cfg, err := config.Get(ctx); err == nil && cfg.Telegram.MessageTTL > 0 {
		ttl = cfg.Telegram.MessageTTL
	}
	if _, err := sched.Add(time.Now().Add(ttl), func(ctx context.Context) {
		if _, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: messageID,
		}); err != nil {
			_, l := log.AsNamedChild(ctx, "AutoDelete")
			l.Debug("failed to delete stale message",
				zap.Int64("chat_id", chatID),
				zap.Int("message_id", messageID),
				zap.Error(err),
			)
		}
	}); err != nil {
		_, l := log.AsNamedChild(ctx, "AutoDelete")
		l.Debug("failed to schedule message deletion",
			zap.Int64("chat_id", chatID),
			zap.Int("message_id", messageID),
			zap.Error(err),
		)
	}
}
