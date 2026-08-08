package handlers

import (
	"context"
	"fmt"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/internal/template"
	im "github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram/autodelete"
	"github.com/fmotalleb/north_outage/telegram/helpers"
	"github.com/fmotalleb/north_outage/telegram/message"
)

const maxListItems = 20

func registerMyListHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, "mylist") || isCommand(update, "list")
	}, myList)
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, "clear")
	}, clearAll)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "rm_listener:", bot.MatchTypePrefix, removeListener)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "close_list", bot.MatchTypeExact, closeList)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "confirm_clear", bot.MatchTypeExact, confirmClear)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "cancel_clear", bot.MatchTypeExact, cancelClear)
}

func myList(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	chat := update.Message.Chat
	chatID := chat.ID
	ctx, l := log.AsNamedChild(ctx, "myList")
	l = l.With(zap.Any("chat", chat))
	ctx = log.WithLogger(ctx, l)

	l.Debug("querying listeners from db", zap.Int64("chat_id", chatID))
	db := database.Get()
	var listeners []im.Listener
	err := db.Where("telegram_c_id = ?", chatID).Find(&listeners).Error

	mp := helpers.MakeMessage(update)

	if err != nil {
		l.Error("failed to query listeners", zap.Error(err))
		mp.Text = "خطا در دریافت داده"
	} else if len(listeners) == 0 {
		l.Debug("no listeners found for chat")
		mp.Text = "📭 شما هیچ آیتمی را مانیتور نکردید.\n\nبرای افزودن، یک پیام متنی با آدرس موردنظر بفرستید یا از دستور /search استفاده کنید."
	} else {
		l.Debug("listeners retrieved", zap.Int("count", len(listeners)))
		// Limit to maxListItems
		if len(listeners) > maxListItems {
			listeners = listeners[:maxListItems]
			l.Debug("listeners truncated to max", zap.Int("max", maxListItems))
		}

		data := map[string]any{
			"listeners": listeners,
		}
		var out string
		out, err = message.EvaluateMessageTemplate(message.List, data, update)
		if err != nil {
			l.Error("failed to evaluate list template", zap.Error(err))
			mp.Text = "خطایی در نمایش خروجی پیش اومده"
		} else {
			l.Debug("list template evaluated successfully")
			mp.Text = out
			mp.ReplyMarkup = &models.InlineKeyboardMarkup{
				InlineKeyboard: buildMyListKeyboard(listeners),
			}
		}
	}

	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send mylist", zap.Error(err))
		return
	}
	l.Debug("mylist sent", zap.Int("msg_id", msg.ID), zap.Int("listener_count", len(listeners)))
	if mp.ReplyMarkup != nil {
		autodelete.Schedule(ctx, b, chatID, msg.ID, update.Message.ID)
	}
}

func buildMyListKeyboard(listeners []im.Listener) [][]models.InlineKeyboardButton {
	n := len(listeners)
	if n > maxListItems {
		n = maxListItems
	}

	buttons := make([][]models.InlineKeyboardButton, 0, n+1)
	for i := 0; i < n; i++ {
		li := listeners[i]
		// Button label via template: show both city and search term
		data := map[string]any{
			"City":       li.City,
			"SearchTerm": li.SearchTerm,
		}
		label, _ := template.EvaluateTemplate(message.RemoveBtn, data)
		buttons = append(buttons, []models.InlineKeyboardButton{
			{
				Text:         label,
				CallbackData: fmt.Sprintf("rm_listener:%d", li.ID),
			},
		})
	}
	// Close button
	buttons = append(buttons, []models.InlineKeyboardButton{
		{
			Text:         "✖️ بستن",
			CallbackData: "close_list",
		},
	})
	return buttons
}

func removeListener(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.CallbackQuery == nil {
		return
	}
	chat := update.CallbackQuery.Message.Message.Chat
	ctx, l := log.AsNamedChild(ctx, "removeListener")
	l = l.With(zap.Any("chat", chat))
	ctx = log.WithLogger(ctx, l)

	// Parse the listener ID
	var id uint
	n, err := fmt.Sscanf(update.CallbackQuery.Data, "rm_listener:%d", &id)
	if err != nil || n != 1 || id == 0 {
		l.Error("failed to parse listener ID from callback", zap.String("data", update.CallbackQuery.Data))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "خطا: شناسه نامعتبر",
			ShowAlert:       false,
		})
		return
	}

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	l.Debug("removing listener", zap.Uint("listener_id", id), zap.Int64("chat_id", chatID))

	// Delete the listener, scoped to this chat
	db := database.Get()
	result := db.Where("id = ? AND telegram_c_id = ?", id, chatID).Delete(&im.Listener{})
	if result.RowsAffected == 0 {
		l.Error("listener not found or not owned by this chat",
			zap.Uint("listener_id", id), zap.Int64("chat_id", chatID))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "این آیتم یافت نشد یا متعلق به شما نیست",
			ShowAlert:       false,
		})
		return
	}
	l.Debug("listener deleted", zap.Uint("listener_id", id), zap.Int64("rows_affected", result.RowsAffected))

	// Answer callback — no alert, just a quick toast
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	// Refresh the list
	l.Debug("refreshing listener list after removal")
	var listeners []im.Listener
	err = db.Where("telegram_c_id = ?", chatID).Find(&listeners).Error

	msg := update.CallbackQuery.Message.Message

	if err != nil {
		l.Error("failed to refresh listeners after removal", zap.Error(err))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "خطا در دریافت داده",
			ShowAlert:       false,
		})
		return
	}

	if len(listeners) == 0 {
		l.Debug("all listeners removed — updating message")
		// All removed — update text and remove keyboard
		editParams := &bot.EditMessageTextParams{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.ID,
			Text:        "✅ همه آیتم‌ها حذف شدند.\n\nبرای افزودن، از /search استفاده کنید.",
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: make([][]models.InlineKeyboardButton, 0)},
		}
		_, _ = b.EditMessageText(ctx, editParams)
		return
	}

	// Limit to maxListItems
	if len(listeners) > maxListItems {
		listeners = listeners[:maxListItems]
	}

	// Build updated message via template
	data := map[string]any{
		"listeners": listeners,
	}
	text, err := message.EvaluateMessageTemplate(message.List, data, update)
	if err != nil {
		l.Error("failed to evaluate list template", zap.Error(err))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "خطایی در نمایش خروجی پیش اومده",
			ShowAlert:       false,
		})
		return
	}
	l.Debug("list template re-evaluated after removal")

	editParams := &bot.EditMessageTextParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: buildMyListKeyboard(listeners),
		},
	}
	_, _ = b.EditMessageText(ctx, editParams)
	l.Debug("list message updated after removal")
}

func clearAll(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	chat := update.Message.Chat
	chatID := chat.ID
	ctx, l := log.AsNamedChild(ctx, "clearAll")
	l = l.With(zap.Any("chat", chat))
	ctx = log.WithLogger(ctx, l)

	db := database.Get()
	var count int64
	l.Debug("counting listeners for chat", zap.Int64("chat_id", chatID))
	db.Model(&im.Listener{}).Where("telegram_c_id = ?", chatID).Count(&count)

	mp := helpers.MakeMessage(update)

	if count == 0 {
		l.Debug("no listeners to clear")
		mp.Text = "📭 شما هیچ آیتمی برای پاک کردن ندارید."
		msg, err := b.SendMessage(ctx, mp)
		if err != nil {
			l.Error("failed to send empty clear message", zap.Error(err))
			return
		}
		l.Debug("empty clear message sent", zap.Int("msg_id", msg.ID))
		return
	}

	l.Debug("showing clear confirmation", zap.Int64("listener_count", count))
	data := map[string]any{
		"count": count,
	}
	out, err := message.EvaluateMessageTemplate(message.ClearConfirm, data, update)
	if err != nil {
		l.Error("failed to evaluate clear confirmation template", zap.Error(err))
		mp.Text = "خطایی در نمایش پیام پیش اومده"
	} else {
		mp.Text = out
	}
	mp.ReplyMarkup = &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "✅ بله، همه را حذف کن",
					CallbackData: "confirm_clear",
				},
			},
			{
				{
					Text:         "❌ خیر، منصرف شدم",
					CallbackData: "cancel_clear",
				},
			},
		},
	}

	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send clear confirmation", zap.Error(err))
		return
	}
	l.Debug("clear confirmation sent", zap.Int("msg_id", msg.ID))
	if mp.ReplyMarkup != nil {
		autodelete.Schedule(ctx, b, chatID, msg.ID, update.Message.ID)
	}
}

func confirmClear(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.CallbackQuery == nil {
		return
	}
	ctx, l := log.AsNamedChild(ctx, "confirmClear")
	l = l.With(zap.Any("chat", update.CallbackQuery.Message.Message.Chat))
	ctx = log.WithLogger(ctx, l)
	l.Debug("user confirmed clear-all")

	chatID := update.CallbackQuery.Message.Message.Chat.ID

	// Answer callback
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	db := database.Get()
	l.Debug("deleting all listeners for chat", zap.Int64("chat_id", chatID))
	result := db.Where("telegram_c_id = ?", chatID).Delete(&im.Listener{})
	l.Debug("listeners deleted", zap.Int64("rows_affected", result.RowsAffected))

	msg := update.CallbackQuery.Message.Message

	data := map[string]any{
		"count": result.RowsAffected,
	}
	text, err := message.EvaluateMessageTemplate(message.ClearDone, data, update)
	if err != nil {
		l.Error("failed to evaluate clear done template", zap.Error(err))
	}

	editParams := &bot.EditMessageTextParams{
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: make([][]models.InlineKeyboardButton, 0)},
	}
	_, _ = b.EditMessageText(ctx, editParams)

	l.Debug("clear-all done, message updated")
}

func cancelClear(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.CallbackQuery == nil {
		return
	}
	ctx, l := log.AsNamedChild(ctx, "cancelClear")
	l = l.With(zap.Any("chat", update.CallbackQuery.Message.Message.Chat))
	ctx = log.WithLogger(ctx, l)
	l.Debug("user cancelled clear-all")

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "حذف لغو شد ✓",
		ShowAlert:       false,
	})

	// Remove the keyboard
	msg := update.CallbackQuery.Message.Message
	if msg != nil {
		l.Debug("removing inline keyboard on cancel")
		editParams := &bot.EditMessageReplyMarkupParams{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.ID,
			ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: make([][]models.InlineKeyboardButton, 0)},
		}
		_, _ = b.EditMessageReplyMarkup(ctx, editParams)
	}
}

func closeList(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.CallbackQuery == nil {
		return
	}
	ctx, l := log.AsNamedChild(ctx, "closeList")
	l = l.With(zap.Any("chat", update.CallbackQuery.Message.Message.Chat))
	ctx = log.WithLogger(ctx, l)
	l.Debug("user closed list")

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "بسته شد ✓",
		ShowAlert:       false,
	})

	// Remove the keyboard
	msg := update.CallbackQuery.Message.Message
	if msg != nil {
		l.Debug("removing inline keyboard on close")
		editParams := &bot.EditMessageReplyMarkupParams{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.ID,
			ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: make([][]models.InlineKeyboardButton, 0)},
		}
		_, _ = b.EditMessageReplyMarkup(ctx, editParams)
	}
}
