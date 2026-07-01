package handlers

import (
	"context"

	"github.com/go-telegram/bot"
)

func SetupHandlers(ctx context.Context, b *bot.Bot) {
	registerHelpHandlers(b)
	registerSearchHandlers(b)
	registerVersionHandlers(b)
	registerListenHandlers(b)
}
