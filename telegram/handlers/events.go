package handlers

import (
	"context"
	"sort"
	"time"

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

const eventsCommand = "events"

func registerEventsHandlers(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool {
		return isCommand(update, eventsCommand)
	}, events)
}

// events lists the user's monitored outages that are currently active or
// upcoming (i.e. events whose end time has not passed yet).
func events(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	chat := update.Message.Chat
	chatID := chat.ID
	ctx, l := log.AsNamedChild(ctx, "events")
	l = l.With(zap.Any("chat", chat))
	ctx = log.WithLogger(ctx, l)

	db := database.Get()
	var listeners []im.Listener
	if err := db.Where("telegram_c_id = ?", chatID).Find(&listeners).Error; err != nil {
		l.Error("failed to query listeners", zap.Error(err))
	}

	evs := make([]im.Event, 0)
	seen := make(map[uint]struct{})
	now := time.Now()
	for _, li := range listeners {
		var found []im.Event
		if err := db.Table("events").
			Where("city = ? AND address LIKE ? AND end_at >= ?", li.City, "%"+li.SearchTerm+"%", now).
			Order("start_at ASC").
			Limit(maxListItems).
			Find(&found).Error; err != nil {
			l.Error("failed to query events", zap.Error(err), zap.Uint("listener_id", li.ID))
			continue
		}
		for _, ev := range found {
			if _, ok := seen[ev.ID]; ok {
				continue
			}
			seen[ev.ID] = struct{}{}
			evs = append(evs, ev)
		}
	}
	sort.Slice(evs, func(i, j int) bool { return evs[i].Start.Before(evs[j].Start) })

	mp := helpers.MakeMessage(update)
	data := map[string]any{
		"events": evs,
	}
	out, err := message.EvaluateMessageTemplate(message.Events, data, update)
	if err != nil {
		l.Error("failed to evaluate events template", zap.Error(err))
		mp.Text = "خطایی در نمایش خروجی پیش اومده"
	} else {
		mp.Text = out
	}

	msg, err := b.SendMessage(ctx, mp)
	if err != nil {
		l.Error("failed to send events", zap.Error(err))
		return
	}
	l.Debug("events sent", zap.Int("msg_id", msg.ID), zap.Int("event_count", len(evs)))
	autodelete.Schedule(ctx, b, chatID, msg.ID)
}
