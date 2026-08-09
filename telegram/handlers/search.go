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
	"github.com/fmotalleb/north_outage/telegram/autodelete"
	"github.com/fmotalleb/north_outage/telegram/helpers"
	"github.com/fmotalleb/north_outage/telegram/message"
)

const (
	maxSearchResult = 10
	maxQueryLen     = 20
	searchCommand   = "search"
)

func registerSearchHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, searchCommand)
	}, search)
}

func registerTextHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		if update == nil || update.Message == nil {
			return false
		}
		text := strings.TrimSpace(update.Message.Text)
		return text != "" && !strings.HasPrefix(text, "/")
	}, searchByText)
}

// search is the /search command handler.
func search(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	chat := update.Message.Chat
	ctx, l := log.AsNamedChild(ctx, "search")
	l = l.With(zap.Any("chat", chat))
	ctx = log.WithLogger(ctx, l)

	query := commandArgument(update.Message.Text, searchCommand)
	if query == "" {
		l.Debug("empty /search command — prompting user")
		mp := helpers.MakeMessage(update)
		mp.Text = "عبارت جستجو را بعد از /search بنویس."
		msg, err := b.SendMessage(ctx, mp)
		if err != nil {
			l.Error("failed to send empty search prompt", zap.Error(err))
			return
		}
		l.Debug("sent empty-search prompt", zap.Int("msg_id", msg.ID))
		return
	}
	l.Debug("handling /search command", zap.String("query", query))
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
	chat := update.Message.Chat
	ctx, l := log.AsNamedChild(ctx, "searchByText")
	l = l.With(zap.Any("chat", chat))
	ctx = log.WithLogger(ctx, l)
	l.Debug("treating plain text as search", zap.String("query", query))
	handleSearch(ctx, b, update, query)
}

// handleSearch is the shared search logic used by both /search and plain text.
func handleSearch(ctx context.Context, b *bot.Bot, update *models.Update, query string) {
	query = truncateQuery(query)
	ctx, l := log.AsNamedChild(ctx, "handleSearch")
	l.Debug("fetching events matching query", zap.String("query", query))
	events, err := fetchEvents(query)
	if err != nil {
		l.Error("failed to fetch events from db", zap.Error(err), zap.String("query", query))
	} else {
		l.Debug("events fetched", zap.Int("count", len(events)), zap.String("query", query))
	}

	mp := helpers.MakeMessage(update)
	if err != nil {
		mp.Text = "خطا در دریافت داده"
	} else {
		data := map[string]any{
			"results": events,
		}
		var out string
		out, err = message.EvaluateMessageTemplate(message.Search, data, update)
		if err != nil {
			l.Error("failed to evaluate search template", zap.Error(err))
			mp.Text = renderErrorText
		} else {
			l.Debug("search template evaluated", zap.Int("result_count", len(events)))
			mp.Text = out
		}
	}

	if len(events) > 0 {
		l.Debug("building inline keyboard", zap.Int("button_count", len(events)))
		mp.ReplyMarkup = &models.InlineKeyboardMarkup{
			InlineKeyboard: buildBtns(query, events),
		}
	}

	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send search results", zap.Error(err))
		return
	}
	l.Debug("search results sent", zap.Int("msg_id", msg.ID), zap.Int("result_count", len(events)))
	if len(events) > 0 {
		autodelete.Schedule(ctx, b, msg.Chat.ID, msg.ID, update.Message.ID)
	}
}

// truncateQuery limits the search query to maxQueryLen characters (runes) so
// long free-text messages cannot produce unbounded LIKE queries.
func truncateQuery(query string) string {
	runes := []rune(strings.TrimSpace(query))
	if len(runes) <= maxQueryLen {
		return string(runes)
	}
	return string(runes[:maxQueryLen])
}

func fetchEvents(search string) ([]im.Event, error) {
	search = strings.TrimSpace(search)
	out := make([]im.Event, 0, maxSearchResult)
	err := database.Get().
		Table("events").
		Where("address LIKE ?", "%"+search+"%").
		Limit(maxSearchResult).
		Find(&out).Error
	return out, err
}

// buildBtns builds inline keyboard buttons for search results.
// Each button has a simple "+ بهم خبر بده" label (the listener is created
// using the search term, not the full address, so showing address was misleading).
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

		btn := models.InlineKeyboardButton{
			Text:         "➕ بهم خبر بده",
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
