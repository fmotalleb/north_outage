package telegram

import (
	"context"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"go.opentelemetry.io/otel/attribute"

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
		if user := userFromUpdate(update); user != nil {
			attrs := []attribute.KeyValue{
				attribute.Int64("telegram.user_id", user.ID),
			}
			if user.Username != "" {
				attrs = append(attrs, attribute.String("telegram.username", user.Username))
			}
			if user.FirstName != "" {
				attrs = append(attrs, attribute.String("telegram.first_name", user.FirstName))
			}
			if user.LastName != "" {
				attrs = append(attrs, attribute.String("telegram.last_name", user.LastName))
			}
			span.SetAttributes(attrs...)
		}
		next(ctx, b, update)
	}
}

// userFromUpdate returns the user who triggered the update, or nil.
func userFromUpdate(update *tgmodels.Update) *tgmodels.User {
	switch {
	case update == nil:
		return nil
	case update.Message != nil:
		return update.Message.From
	case update.EditedMessage != nil:
		return update.EditedMessage.From
	case update.ChannelPost != nil:
		return update.ChannelPost.From
	case update.EditedChannelPost != nil:
		return update.EditedChannelPost.From
	case update.CallbackQuery != nil:
		u := update.CallbackQuery.From
		return &u
	}
	return nil
}
