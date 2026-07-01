package handlers

import (
	"context"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/telegram/helpers"
	"github.com/fmotalleb/north_outage/telegram/template"
)

func registerHelpHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, "start")
	}, help)
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, "help")
	}, help)
}

func help(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	chat := update.Message.Chat
	l := log.Of(ctx).
		Named("help").
		With(zap.Any("chat", chat))
	mp := helpers.MakeMessage(update)
	out, err := template.EvaluateTemplate(template.Help, nil, update)
	if err != nil {
		l.Error("failed to generate help message")
	} else {
		mp.Text = out
	}

	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send help message", zap.Error(err))
		return
	}
	l.Debug("message sent", zap.Int("id", msg.ID))
}
