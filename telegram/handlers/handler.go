package handlers

import (
	"context"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
)

func SetupHandlers(ctx context.Context, b *bot.Bot) {
	l := log.Of(ctx).Named("SetupHandlers")
	l.Debug("registering telegram handlers")
	registerHelpHandlers(b)
	registerSearchHandlers(b)
	registerTextHandlers(b)
	registerVersionHandlers(b)
	registerListenHandlers(b)
	registerMyListHandlers(b)
	l.Debug("all telegram handlers registered")
}
