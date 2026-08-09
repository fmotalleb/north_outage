package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/memory"
	im "github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram/helpers"
)

type listenReq struct {
	city   string
	search string
}

var mem = memory.NewRuntime[listenReq]()

const memoryTTL = time.Minute

func registerListenHandlers(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "listen:", bot.MatchTypePrefix, listen)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "cancel", bot.MatchTypeExact, cancelListen)
}

func listen(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.CallbackQuery == nil {
		return
	}
	chat := update.CallbackQuery.Message.Message.Chat
	ctx, l := log.AsNamedChild(ctx, "listen")
	l = l.With(zap.Any("chat", chat))
	ctx = log.WithLogger(ctx, l)

	if update.CallbackQuery.Message.Message == nil {
		l.Error("callback query message missing")
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "این دکمه دیگر معتبر نیست",
			ShowAlert:       false,
		})
		return
	}

	key := strings.TrimPrefix(update.CallbackQuery.Data, "listen:")
	l.Debug("listen callback triggered", zap.String("key", key))

	cp := &bot.AnswerCallbackQueryParams{
		ShowAlert:       false,
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "لطفا کمی صبر کنید",
	}
	_, _ = b.AnswerCallbackQuery(ctx, cp)

	mp := helpers.MakeMessage(update)
	data, ok := mem.Pop(key)
	if !ok {
		l.Error("listen request expired or not found in memory", zap.String("key", key))
		responseError(ctx, l, mp, b)
		return
	}
	l.Debug("listen request found in memory", zap.String("city", data.city), zap.String("search", data.search))

	db := database.Get()
	listener := new(im.Listener)
	listener.City = data.city
	listener.SearchTerm = data.search
	listener.TelegramCID = update.CallbackQuery.Message.Message.Chat.ID
	listener.TelegramTID = int64(update.CallbackQuery.Message.Message.MessageThreadID)
	mp.Text = "در صورت دریافت اطلاعات جدید و همچنین بیست دقیقه قبل از قطعی بهت اطلاع میدم"

	l.Debug("saving listener to db", zap.Any("listener", listener))
	if err := db.Save(listener).Error; err != nil {
		l.Error("failed to save listener to db", zap.Any("request", listener), zap.Error(err))
		mp.Text = "خطا در ذخیره\u200cسازی داده، احتمالا داری آیتم تکراری ذخیره میکنی"
	} else {
		l.Debug("listener saved successfully", zap.Uint("listener_id", listener.ID))
	}

	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send listen response", zap.Error(err))
		return
	}
	l.Debug("listen response sent", zap.Int("msg_id", msg.ID), zap.Uint("listener_id", listener.ID))
}

func responseError(ctx context.Context, l *zap.Logger, mp *bot.SendMessageParams, b *bot.Bot) {
	l.Debug("sending expiry error to user")
	mp.Text = "درخواست منقضی شد لطفا مجددا جست و جو کنید"
	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send expiry error message", zap.Error(err))
		return
	}
	l.Debug("expiry error sent", zap.Int("msg_id", msg.ID))
}

func cancelListen(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.CallbackQuery == nil {
		return
	}
	chat := update.CallbackQuery.Message.Message.Chat
	ctx, l := log.AsNamedChild(ctx, "cancelListen")
	l = l.With(zap.Any("chat", chat))
	ctx = log.WithLogger(ctx, l)
	l.Debug("user canceled search")

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "لغو شد ✓",
		ShowAlert:       false,
	})

	// Remove the inline keyboard from the original message.
	msg := update.CallbackQuery.Message.Message
	if msg != nil {
		l.Debug("removing inline keyboard from message", zap.Int("msg_id", msg.ID))
		editParams := &bot.EditMessageReplyMarkupParams{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.ID,
			ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: make([][]models.InlineKeyboardButton, 0)},
		}
		_, _ = b.EditMessageReplyMarkup(ctx, editParams)
	}
}

func createRequest(search, city string) string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Errorf("random uuid generation failed: %w", err))
	}
	req := listenReq{
		city:   city,
		search: search,
	}
	key := uuid.String()
	mem.Put(key, req, memoryTTL)
	return key
}
