package handlers

import (
	"context"
	"strings"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/database"
	im "github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram/helpers"
	"github.com/fmotalleb/north_outage/telegram/template"
)

const (
	maxSearchResult = 10
	searchCommand   = "search"
)

func registerSearchHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, searchCommand)
	}, search)
}

func search(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	l := log.Of(ctx).Named("search")
	input := update.Message
	search := commandArgument(input.Text, searchCommand)
	if search == "" {
		mp := helpers.MakeMessage(update)
		mp.Text = "عبارت جستجو را بعد از /search بنویس."
		msg, err := b.SendMessage(ctx, mp)
		if err != nil {
			l.Error("failed to send empty search prompt", zap.Error(err))
			return
		}
		l.Debug("sent prompt", zap.Int("id", msg.ID))
		return
	}

	events, err := fetchEvents(search)

	mp := helpers.MakeMessage(update)
	if err != nil {
		l.Error("failed to fetch data from db", zap.Error(err))
		mp.Text = "خطا در دریافت داده"
	} else {
		data := map[string]any{
			"results": events,
		}
		var out string
		out, err = template.EvaluateTemplate(template.Search, data, update)
		if err != nil {
			l.Error("failed to evaluate template", zap.Error(err), zap.Any("chat", update.Message.Chat))
			mp.Text = "خطایی در نمایش خروجی پیش اومده"
		} else {
			mp.Text = out
		}
	}

	if len(events) > 0 {
		mp.ReplyMarkup = &models.InlineKeyboardMarkup{
			InlineKeyboard: buildBtns(search, events),
		}
	}

	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send search results", zap.Error(err))
		return
	}
	l.Debug("sent message", zap.Int("id", msg.ID))
}

func fetchEvents(search string) ([]im.Event, error) {
	search = strings.TrimSpace(search)
	out := make([]im.Event, 0, maxSearchResult)
	err := database.Get().
		Table("events").
		Where("address LIKE ?", "%"+search+"%").
		Limit(maxSearchResult).
		Find(&out).Error
	// sortEvents(out)
	return out, err
}

func buildBtns(search string, events []im.Event) [][]models.InlineKeyboardButton {
	cityButtons := make([][]models.InlineKeyboardButton, 0)
	seen := make(map[string]struct{})
	for _, ev := range events {
		if _, ok := seen[ev.City]; ok {
			continue
		}
		seen[ev.City] = struct{}{}

		btn := models.InlineKeyboardButton{
			Text:         "🔍 " + ev.City,                             // Glass icon + city name
			CallbackData: "listen:" + createRequest(search, ev.City), // will be handled by callback
		}
		cityButtons = append(cityButtons, []models.InlineKeyboardButton{btn})
	}
	return cityButtons
}
