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

	l := log.Of(ctx).Named("myList")

	db := database.Get()
	var listeners []im.Listener
	err := db.Where("telegram_cid = ?", chatID).Find(&listeners).Error

	mp := helpers.MakeMessage(update)

	if err != nil {
		l.Error("failed to query listeners", zap.Error(err))
		mp.Text = "خطا در دریافت داده"
	} else if len(listeners) == 0 {
		mp.Text = "📭 شما هیچ آیتمی را مانیتور نکردید.\n\nبرای افزودن، یک پیام متنی با آدرس موردنظر بفرستید یا از دستور /search استفاده کنید."
	} else {
		// Limit to maxListItems
		if len(listeners) > maxListItems {
			listeners = listeners[:maxListItems]
		}

		data := map[string]any{
			"listeners": listeners,
		}
		var out string
		out, err = message.EvaluateMessageTemplate(message.List, data, update)
		if err != nil {
			l.Error("failed to evaluate list template", zap.Error(err), zap.Any("chat", chat))
			mp.Text = "خطایی در نمایش خروجی پیش اومده"
		} else {
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
	l.Debug("sent mylist", zap.Int("id", msg.ID))
}

func buildMyListKeyboard(listeners []im.Listener) [][]models.InlineKeyboardButton {
	n := len(listeners)
	if n > maxListItems {
		n = maxListItems
	}

	buttons := make([][]models.InlineKeyboardButton, 0, n+1)
	for i := 0; i < n; i++ {
		li := listeners[i]
		// Button label via template
		data := map[string]any{
			"City": li.City,
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
	l := log.Of(ctx).Named("removeListener")

	// Parse the listener ID
	var id uint
	n, err := fmt.Sscanf(update.CallbackQuery.Data, "rm_listener:%d", &id)
	if err != nil || n != 1 || id == 0 {
		l.Error("failed to parse listener ID", zap.String("data", update.CallbackQuery.Data))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "خطا: شناسه نامعتبر",
			ShowAlert:       true,
		})
		return
	}

	chatID := update.CallbackQuery.Message.Message.Chat.ID

	// Delete the listener, scoped to this chat
	db := database.Get()
	result := db.Where("id = ? AND telegram_cid = ?", id, chatID).Delete(&im.Listener{})
	if result.RowsAffected == 0 {
		l.Error("listener not found or not owned by this chat",
			zap.Uint("id", id), zap.Int64("chat", chatID))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "این آیتم یافت نشد یا متعلق به شما نیست",
			ShowAlert:       true,
		})
		return
	}

	// Answer callback — no alert, just a quick toast
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	// Refresh the list
	var listeners []im.Listener
	err = db.Where("telegram_cid = ?", chatID).Find(&listeners).Error

	msg := update.CallbackQuery.Message.Message

	if err != nil {
		l.Error("failed to refresh listeners after removal", zap.Error(err))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "خطا در دریافت داده",
			ShowAlert:       true,
		})
		return
	}

	if len(listeners) == 0 {
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
			ShowAlert:       true,
		})
		return
	}

	editParams := &bot.EditMessageTextParams{
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: buildMyListKeyboard(listeners),
		},
	}
	_, _ = b.EditMessageText(ctx, editParams)
}

func clearAll(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	chat := update.Message.Chat
	chatID := chat.ID

	l := log.Of(ctx).Named("clearAll")

	db := database.Get()
	var count int64
	db.Model(&im.Listener{}).Where("telegram_cid = ?", chatID).Count(&count)

	mp := helpers.MakeMessage(update)

	if count == 0 {
		mp.Text = "📭 شما هیچ آیتمی برای پاک کردن ندارید."
		msg, err := b.SendMessage(ctx, mp)
		if err != nil {
			l.Error("failed to send empty clear", zap.Error(err))
			return
		}
		l.Debug("sent empty clear", zap.Int("id", msg.ID))
		return
	}

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
	l.Debug("sent clear confirmation", zap.Int("id", msg.ID))
}

func confirmClear(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.CallbackQuery == nil {
		return
	}
	l := log.Of(ctx).Named("confirmClear")

	chatID := update.CallbackQuery.Message.Message.Chat.ID

	// Answer callback
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	db := database.Get()
	result := db.Where("telegram_cid = ?", chatID).Delete(&im.Listener{})

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

	l.Debug("cleared all listeners", zap.Int64("chat", chatID), zap.Int64("count", result.RowsAffected))
}

func cancelClear(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.CallbackQuery == nil {
		return
	}

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "حذف لغو شد ✓",
		ShowAlert:       true,
	})

	// Remove the keyboard
	msg := update.CallbackQuery.Message.Message
	if msg != nil {
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

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "بسته شد ✓",
		ShowAlert:       true,
	})

	// Remove the keyboard
	msg := update.CallbackQuery.Message.Message
	if msg != nil {
		editParams := &bot.EditMessageReplyMarkupParams{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.ID,
			ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: make([][]models.InlineKeyboardButton, 0)},
		}
		_, _ = b.EditMessageReplyMarkup(ctx, editParams)
	}
}
