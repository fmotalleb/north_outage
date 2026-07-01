package handlers

import (
	"context"

	"github.com/fmotalleb/go-tools/git"
	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/telegram/helpers"
)

func registerVersionHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, "version")
	}, version)
}

func version(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	chat := update.Message.Chat
	l := log.Of(ctx).
		Named("version").
		With(zap.Any("chat", chat))
	mp := helpers.MakeMessage(update)

	mp.Text = git.String()

	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send version message", zap.Error(err))
		return
	}
	l.Debug("message sent", zap.Int("id", msg.ID))
}
