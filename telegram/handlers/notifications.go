package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/internal/template"
	im "github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram/autodelete"
	"github.com/fmotalleb/north_outage/telegram/helpers"
	"github.com/fmotalleb/north_outage/telegram/message"
	"gorm.io/gorm"
)

const notificationsListHeader = "📢"

// notificationRow is a single user-facing notification entry in the
// /notifications list. Rows are deduplicated per listener+event pair.
type notificationRow struct {
	ListenerID uint
	EventID    uint
	Event      *im.Event
	Message    string
	SentAt     time.Time
	Muted      bool
}

func registerNotificationsHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, "notifications")
	}, notifications)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "mute:", bot.MatchTypePrefix, muteNotification)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "unmute:", bot.MatchTypePrefix, unmuteNotification)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "close_notifications", bot.MatchTypeExact, closeNotifications)
}

// notifications lists the user's recent notifications and lets them mute or
// unmute each one (DND).
func notifications(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	chat := update.Message.Chat
	chatID := chat.ID
	ctx, l := log.AsNamedChild(ctx, "notifications")
	l = l.With(zap.Any("chat", chat))
	ctx = log.WithLogger(ctx, l)

	mp := helpers.MakeMessage(update)
	rows, err := notificationRows(database.Get().WithContext(ctx), ctx, chatID)
	if err != nil {
		l.Error("failed to query notifications", zap.Error(err))
		mp.Text = "خطا در دریافت اعلان‌ها"
	} else if len(rows) == 0 {
		l.Debug("no notifications found for chat")
		mp.Text = "📭 هنوز اعلانی برای شما ثبت نشده است.\n\nبرای شروع، از /search استفاده کنید."
	} else {
		l.Debug("notifications retrieved", zap.Int("count", len(rows)))
		if len(rows) > maxListItems {
			rows = rows[:maxListItems]
		}
		data := map[string]any{
			"notifications": rows,
		}
		out, err := template.EvaluateTemplate(message.Notifications, data)
		if err != nil {
			l.Error("failed to evaluate notifications template", zap.Error(err))
			mp.Text = "خطایی در نمایش خروجی پیش اومده"
		} else {
			mp.Text = out
			mp.ReplyMarkup = &models.InlineKeyboardMarkup{
				InlineKeyboard: buildNotificationsKeyboard(rows),
			}
		}
	}

	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send notifications", zap.Error(err))
		return
	}
	l.Debug("notifications sent", zap.Int("msg_id", msg.ID), zap.Int("notification_count", len(rows)))
	if mp.ReplyMarkup != nil {
		autodelete.Schedule(ctx, b, chatID, msg.ID)
	}
}

func muteNotification(ctx context.Context, b *bot.Bot, update *models.Update) {
	toggleNotificationMute(ctx, b, update, true)
}

func unmuteNotification(ctx context.Context, b *bot.Bot, update *models.Update) {
	toggleNotificationMute(ctx, b, update, false)
}

func toggleNotificationMute(ctx context.Context, b *bot.Bot, update *models.Update, mute bool) {
	if update == nil || update.CallbackQuery == nil {
		return
	}
	cq := update.CallbackQuery
	msg := cq.Message.Message
	if msg == nil {
		return
	}
	chatID := msg.Chat.ID
	ctx, l := log.AsNamedChild(ctx, "toggleNotificationMute")
	l = l.With(zap.Any("chat", msg.Chat))
	ctx = log.WithLogger(ctx, l)

	prefix := "mute:"
	if !mute {
		prefix = "unmute:"
	}
	listenerID, eventID, ok := parseMuteTarget(strings.TrimPrefix(cq.Data, prefix))
	if !ok {
		l.Error("failed to parse listener/event ID from callback",
			zap.String("data", cq.Data))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: cq.ID,
			Text:            "شناسه نامعتبر",
			ShowAlert:       false,
		})
		return
	}

	db := database.Get().WithContext(ctx)
	var listener im.Listener
	if err := db.First(&listener, listenerID).Error; err != nil {
		l.Error("listener not found", zap.Uint("listener_id", listenerID), zap.Error(err))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: cq.ID,
			Text:            "این آیتم یافت نشد",
			ShowAlert:       false,
		})
		return
	}
	if listener.TelegramCID != chatID {
		l.Error("listener not owned by chat",
			zap.Uint("listener_id", listenerID), zap.Int64("chat_id", chatID))
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: cq.ID,
			Text:            "این آیتم متعلق به شما نیست",
			ShowAlert:       false,
		})
		return
	}

	if mute {
		m := im.NotificationMute{
			ListenerID: listenerID,
			EventID:    eventID,
			MutedAt:    time.Now(),
		}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "listener_id"}, {Name: "event_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"muted_at"}),
		}).Create(&m).Error; err != nil {
			l.Error("failed to create notification mute", zap.Error(err))
		}
		// Cancel any already-scheduled upcoming reminder for this pair.
		if err := db.Where("listener_id = ? AND event_id = ?", listenerID, eventID).
			Delete(&im.ScheduledNotification{}).Error; err != nil {
			l.Error("failed to delete scheduled notification", zap.Error(err))
		}
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: cq.ID,
			Text:            "🔕 اعلان این رویداد خاموش شد",
			ShowAlert:       false,
		})
	} else {
		if err := db.Where("listener_id = ? AND event_id = ?", listenerID, eventID).
			Delete(&im.NotificationMute{}).Error; err != nil {
			l.Error("failed to delete notification mute", zap.Error(err))
		}
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: cq.ID,
			Text:            "🔔 اعلان این رویداد روشن شد",
			ShowAlert:       false,
		})
	}

	// Update the message: re-render the whole list, or toggle the single
	// button on a standalone notification message.
	if strings.HasPrefix(msg.Text, notificationsListHeader) {
		if err := renderNotificationsList(ctx, b, msg, l); err != nil {
			l.Error("failed to refresh notifications list", zap.Error(err))
		}
	} else {
		editNotificationsButton(ctx, b, msg, listenerID, eventID, mute)
	}
}

func closeNotifications(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.CallbackQuery == nil {
		return
	}
	ctx, l := log.AsNamedChild(ctx, "closeNotifications")
	l = l.With(zap.Any("chat", update.CallbackQuery.Message.Message.Chat))
	ctx = log.WithLogger(ctx, l)
	l.Debug("user closed notifications list")

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "بسته شد ✓",
		ShowAlert:       false,
	})

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

// notificationRows returns the user's notifications, deduplicated per
// listener+event pair, most recent first, along with their mute state.
func notificationRows(db *gorm.DB, ctx context.Context, chatID int64) ([]notificationRow, error) {
	var notifs []im.Notification
	if err := db.Model(&im.Notification{}).
		Joins("JOIN listeners ON listeners.id = notifications.listener_id").
		Where("listeners.telegram_c_id = ?", chatID).
		Preload("Event").
		Order("notifications.sent_at DESC").
		Limit(200).
		Find(&notifs).Error; err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(notifs))
	rows := make([]notificationRow, 0, len(notifs))
	for _, n := range notifs {
		key := fmt.Sprintf("%d:%d", n.ListenerID, n.EventID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		rows = append(rows, notificationRow{
			ListenerID: n.ListenerID,
			EventID:    n.EventID,
			Event:      n.Event,
			Message:    n.Message,
			SentAt:     n.SentAt,
		})
	}

	var mutes []im.NotificationMute
	if err := db.Joins("JOIN listeners ON listeners.id = notification_mutes.listener_id").
		Where("listeners.telegram_c_id = ?", chatID).
		Find(&mutes).Error; err != nil {
		return nil, err
	}
	muted := make(map[string]struct{}, len(mutes))
	for _, mm := range mutes {
		muted[fmt.Sprintf("%d:%d", mm.ListenerID, mm.EventID)] = struct{}{}
	}
	for i := range rows {
		if _, ok := muted[fmt.Sprintf("%d:%d", rows[i].ListenerID, rows[i].EventID)]; ok {
			rows[i].Muted = true
		}
	}
	return rows, nil
}

// parseMuteTarget parses "listenerID:eventID" from a mute/unmute callback
// payload. Both IDs must be present and non-zero.
func parseMuteTarget(data string) (listenerID, eventID uint, ok bool) {
	parts := strings.Split(data, ":")
	if len(parts) != 2 {
		return 0, 0, false
	}
	var li, ev uint
	if _, err := fmt.Sscan(parts[0], &li); err != nil {
		return 0, 0, false
	}
	if _, err := fmt.Sscan(parts[1], &ev); err != nil {
		return 0, 0, false
	}
	if li == 0 || ev == 0 {
		return 0, 0, false
	}
	return li, ev, true
}

func buildNotificationsKeyboard(rows []notificationRow) [][]models.InlineKeyboardButton {
	buttons := make([][]models.InlineKeyboardButton, 0, len(rows)+1)
	for _, r := range rows {
		if r.Muted {
			buttons = append(buttons, []models.InlineKeyboardButton{
				{
					Text:         "🔔 فعال کردن",
					CallbackData: fmt.Sprintf("unmute:%d:%d", r.ListenerID, r.EventID),
				},
			})
		} else {
			buttons = append(buttons, []models.InlineKeyboardButton{
				{
					Text:         "🔕 غیرفعال کردن",
					CallbackData: fmt.Sprintf("mute:%d:%d", r.ListenerID, r.EventID),
				},
			})
		}
	}
	buttons = append(buttons, []models.InlineKeyboardButton{
		{
			Text:         "✖️ بستن",
			CallbackData: "close_notifications",
		},
	})
	return buttons
}

func renderNotificationsList(ctx context.Context, b *bot.Bot, msg *models.Message, l *zap.Logger) error {
	rows, err := notificationRows(database.Get().WithContext(ctx), ctx, msg.Chat.ID)
	if err != nil {
		return err
	}
	if len(rows) > maxListItems {
		rows = rows[:maxListItems]
	}
	data := map[string]any{
		"notifications": rows,
	}
	text, err := template.EvaluateTemplate(message.Notifications, data)
	if err != nil {
		return err
	}
	editParams := &bot.EditMessageTextParams{
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: buildNotificationsKeyboard(rows)},
	}
	_, err = b.EditMessageText(ctx, editParams)
	return err
}

// editNotificationsButton toggles the mute button on a standalone
// notification message after the user muted/unmuted it.
func editNotificationsButton(ctx context.Context, b *bot.Bot, msg *models.Message, listenerID, eventID uint, muted bool) {
	text := "🔕 غیرفعال کردن"
	data := fmt.Sprintf("mute:%d:%d", listenerID, eventID)
	if muted {
		text = "🔔 فعال کردن"
		data = fmt.Sprintf("unmute:%d:%d", listenerID, eventID)
	}
	editParams := &bot.EditMessageReplyMarkupParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         text,
						CallbackData: data,
					},
				},
			},
		},
	}
	_, _ = b.EditMessageReplyMarkup(ctx, editParams)
}
