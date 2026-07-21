package handlers

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fmotalleb/go-tools/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/database"
	im "github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram/helpers"
	"github.com/fmotalleb/north_outage/telegram/message"
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

func registerTextHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		msg := messageFromUpdate(update)
		if msg == nil {
			return false
		}
		text := strings.TrimSpace(msg.Text)
		return text != "" && !strings.HasPrefix(text, "/")
	}, searchByText)
}

// search is the /search command handler.
func search(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	query := commandArgument(update.Message.Text, searchCommand)
	if query == "" {
		mp := helpers.MakeMessage(update)
		mp.Text = "عبارت جستجو را بعد از /search بنویس."
		msg, err := b.SendMessage(ctx, mp)
		if err != nil {
			log.Of(ctx).Error("failed to send empty search prompt", zap.Error(err))
			return
		}
		log.Of(ctx).Debug("sent prompt", zap.Int("id", msg.ID))
		return
	}
	handleSearch(ctx, b, update, query)
}

// searchByText handles plain text messages as search queries.
func searchByText(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	query := strings.TrimSpace(update.Message.Text)
	if query == "" {
		return
	}
	log.Of(ctx).Debug("treating plain text as search", zap.String("query", query))
	handleSearch(ctx, b, update, query)
}

// handleSearch is the shared search logic used by both /search and plain text.
func handleSearch(ctx context.Context, b *bot.Bot, update *models.Update, query string) {
	l := log.Of(ctx).Named("search")
	events, err := fetchEvents(query)

	mp := helpers.MakeMessage(update)
	if err != nil {
		l.Error("failed to fetch data from db", zap.Error(err))
		mp.Text = "خطا در دریافت داده"
	} else {
		data := map[string]any{
			"results": events,
		}
		var out string
		out, err = message.EvaluateMessageTemplate(message.Search, data, update)
		if err != nil {
			l.Error("failed to evaluate template", zap.Error(err), zap.Any("chat", update.Message.Chat))
			mp.Text = "خطایی در نمایش خروجی پیش اومده"
		} else {
			mp.Text = out
		}
	}

	if len(events) > 0 {
		mp.ReplyMarkup = &models.InlineKeyboardMarkup{
			InlineKeyboard: buildBtns(query, events),
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

// truncate trims a string to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

// buildBtns builds inline keyboard buttons for search results.
// Each button shows the city and a short address snippet, deduplicated by city.
// A cancel button is appended at the bottom.
func buildBtns(search string, events []im.Event) [][]models.InlineKeyboardButton {
	buttons := make([][]models.InlineKeyboardButton, 0)
	seen := make(map[string]struct{})
	for _, ev := range events {
		if len(buttons) >= maxSearchResult {
			break
		}
		if _, ok := seen[ev.City]; ok {
			continue
		}
		seen[ev.City] = struct{}{}

		// Button label: city + truncated address
		label := fmt.Sprintf("🔍 %s | %s", ev.City, truncate(ev.Address, 20))
		btn := models.InlineKeyboardButton{
			Text:         label,
			CallbackData: "listen:" + createRequest(search, ev.City),
		}
		buttons = append(buttons, []models.InlineKeyboardButton{btn})
	}

	// Cancel button at the bottom
	buttons = append(buttons, []models.InlineKeyboardButton{
		{
			Text:         "❌ لغو",
			CallbackData: "cancel",
		},
	})

	return buttons
}
