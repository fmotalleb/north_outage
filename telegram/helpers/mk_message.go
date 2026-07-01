package helpers

import (
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func MakeMessage(update *models.Update) *bot.SendMessageParams {
	mp := new(bot.SendMessageParams)
	if update == nil {
		return mp
	}
	input := update.Message
	if input == nil && update.CallbackQuery != nil {
		input = update.CallbackQuery.Message.Message
	}
	if input == nil {
		return mp
	}
	mp.ChatID = input.Chat.ID
	if input.MessageThreadID != 0 {
		mp.MessageThreadID = input.MessageThreadID
	}
	mp.ParseMode = models.ParseModeHTML
	mp.Text = "If you see this message there is a bug in the application, please report to @fmotalleb"
	return mp
}
