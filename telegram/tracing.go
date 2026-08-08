package telegram

import (
	"context"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"

	"github.com/fmotalleb/north_outage/internal/otel"
)

// tracingMiddleware starts a span per incoming update so every bot
// interaction (message, callback query, channel post, ...) shows up in the
// trace. It uses the rate-sampled process-global provider, so telegram
// interactions follow TRACING_RATE.
func tracingMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		name := "telegram.update"
		switch {
		case update == nil:
		case update.Message != nil:
			name = "telegram.message"
		case update.EditedMessage != nil:
			name = "telegram.edited_message"
		case update.ChannelPost != nil:
			name = "telegram.channel_post"
		case update.EditedChannelPost != nil:
			name = "telegram.edited_channel_post"
		case update.CallbackQuery != nil:
			name = "telegram.callback_query"
		}

		ctx, span := otel.Tracer("north_outage.telegram").Start(ctx, name)
		defer span.End()
		next(ctx, b, update)
	}
}
