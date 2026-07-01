package handlers

import (
	"strings"

	"github.com/go-telegram/bot/models"
)

func isCommand(update *models.Update, command string) bool {
	msg := messageFromUpdate(update)
	if msg == nil {
		return false
	}
	return commandName(msg.Text) == command
}

func commandArgument(text, command string) string {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return ""
	}
	if !strings.HasPrefix(fields[0], "/") {
		return ""
	}

	token := strings.TrimPrefix(fields[0], "/")
	if idx := strings.IndexByte(token, '@'); idx >= 0 {
		token = token[:idx]
	}
	if token != command {
		return ""
	}

	if len(fields) < 2 {
		return ""
	}
	return strings.Join(fields[1:], " ")
}

func commandName(text string) string {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return ""
	}
	if !strings.HasPrefix(fields[0], "/") {
		return ""
	}

	token := strings.TrimPrefix(fields[0], "/")
	if idx := strings.IndexByte(token, '@'); idx >= 0 {
		token = token[:idx]
	}

	return token
}

func messageFromUpdate(update *models.Update) *models.Message {
	if update == nil {
		return nil
	}
	if update.Message != nil {
		return update.Message
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Message
	}
	return nil
}
