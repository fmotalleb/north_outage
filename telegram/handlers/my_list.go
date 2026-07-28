package handlers

import (
	"context"
	"fmt"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/database"
	im "github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram/helpers"
)

const maxListItems = 20

func registerMyListHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, "mylist") || isCommand(update, "list")
	}, myList)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "rm_listener:", bot.MatchTypePrefix, removeListener)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "close_list", bot.MatchTypeExact, closeList)
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
	db.Where("telegram_cid = ?", chatID).Find(&listeners)

	mp := helpers.MakeMessage(update)

	if len(listeners) == 0 {
		mp.Text = "📭 شما هیچ آیتمی را مانیتور نکردید.\n\nبرای افزودن، یک پیام متنی با آدرس موردنظر بفرستید یا از دستور /search استفاده کنید."
		msg, err := b.SendMessage(ctx, mp)
		if err != nil {
			l.Error("failed to send empty list", zap.Error(err))
			return
		}
		l.Debug("sent empty list", zap.Int("id", msg.ID))
		return
	}

	text := "📋 <b>لیست مانیتور شده‌های شما</b>:\n\n"
	for i, li := range listeners {
		if i >= maxListItems {
			text += fmt.Sprintf("\n… و %d مورد دیگر", len(listeners)-maxListItems)
			break
		}
		text += fmt.Sprintf("%d. 🏙 <b>%s</b>\n   🔍 %s\n\n", i+1, li.City, li.SearchTerm)
	}
	text += "👇 برای حذف هر آیتم، دکمه مربوطه را بزنید."

	mp.Text = text
	mp.ReplyMarkup = &models.InlineKeyboardMarkup{
		InlineKeyboard: buildMyListKeyboard(listeners),
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
		label := fmt.Sprintf("❌ %s", li.City)
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
	db.Where("telegram_cid = ?", chatID).Find(&listeners)

	msg := update.CallbackQuery.Message.Message

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

	// Build updated message
	text := "📋 <b>لیست مانیتور شده‌های شما</b>:\n\n"
	for i, li := range listeners {
		if i >= maxListItems {
			text += fmt.Sprintf("\n… و %d مورد دیگر", len(listeners)-maxListItems)
			break
		}
		text += fmt.Sprintf("%d. 🏙 <b>%s</b>\n   🔍 %s\n\n", i+1, li.City, li.SearchTerm)
	}
	text += "👇 برای حذف هر آیتم، دکمه مربوطه را بزنید."

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
