package handlers

import (
	"context"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

func SetupHandlers(ctx context.Context, b *bot.Bot) {
	_, l := log.AsNamedChild(ctx, "SetupHandlers")
	l.Debug("registering telegram handlers")
	registerHelpHandlers(b)
	registerSearchHandlers(b)
	registerTextHandlers(b)
	registerVersionHandlers(b)
	registerListenHandlers(b)
	registerMyListHandlers(b)
	registerEventsHandlers(b)
	registerNotificationsHandlers(b)
	registerMenuCommands(ctx, b)
	l.Debug("all telegram handlers registered")
}

// registerMenuCommands registers the bot command list and sets the chat
// menu button so users can access core commands from the "/" button.
func registerMenuCommands(ctx context.Context, b *bot.Bot) {
	_, l := log.AsNamedChild(ctx, "registerMenuCommands")

	commands := []models.BotCommand{
		{Command: "start", Description: "شروع و راهنمای ربات"},
		{Command: "search", Description: "جستجوی قطعی برق"},
		{Command: "events", Description: "رویدادهای فعال و آتی"},
		{Command: "list", Description: "لیست مانیتورهای شما"},
		{Command: "notifications", Description: "اعلان‌های دریافتی"},
		{Command: "clear", Description: "پاک کردن همه مانیتورها"},
	}

	_, err := b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: commands,
	})
	if err != nil {
		l.Error("failed to set bot commands", zap.Error(err))
	}

	_, err = b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
		MenuButton: &models.MenuButtonCommands{
			Type: models.MenuButtonTypeCommands,
		},
	})
	if err != nil {
		l.Error("failed to set chat menu button", zap.Error(err))
	}

	l.Debug("menu commands registered")
}
