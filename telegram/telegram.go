package telegram

import (
	"context"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/internal/template"
	im "github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram/handlers"
	"github.com/fmotalleb/north_outage/telegram/message"
	"github.com/fmotalleb/north_outage/weather"

	"github.com/go-telegram/bot"
)

func Run(ctx context.Context, cfg *config.Config, nc <-chan im.Notification) error {
	l := log.Of(ctx).Named("Telegram")
	ctx = log.WithLogger(ctx, l)
	tel := cfg.Telegram
	if tel.BotKey == "" {
		l.Warn("telegram bot token is not set")
		return nil
	}

	var opts []bot.Option
	client := httpClient(tel.Proxy)

	hc := bot.WithHTTPClient(time.Second*30, client)
	opts = append(opts, hc)

	b, err := bot.New(tel.BotKey, opts...)
	if err != nil {
		l.Error("failed to connect to telegram bot", zap.Error(err))
		return err
	}
	handlers.SetupHandlers(ctx, b)
	go bindToChannel(ctx, b, nc)
	b.Start(ctx)
	return nil
}

func bindToChannel(ctx context.Context, b *bot.Bot, nc <-chan im.Notification) {
	l := log.Of(ctx).Named("binder")
	for {
		select {
		case n := <-nc:
			l.Debug("notification received", zap.Any("event", n))
			sp := new(bot.SendMessageParams)
			sp.ChatID = n.Listener.TelegramCID
			sp.MessageThreadID = int(n.Listener.TelegramTID)
			cfg, err := config.Get(ctx)
			notifyWeather := err == nil && cfg.Weather.Notify
			sp.Text = formatNotification(ctx, n.Message, n.Event, notifyWeather)
			m, err := b.SendMessage(ctx, sp)
			if err != nil {
				l.Error("failed to send message to telegram", zap.Error(err))
			} else {
				l.Debug("telegram message sent", zap.Int("message_id", m.ID))
			}
		case <-ctx.Done():
			return
		}
	}
}

func formatNotification(ctx context.Context, msg string, ev *im.Event, notifyWeather bool) string {
	data := map[string]any{
		"message": msg,
		"event":   ev,
	}

	if notifyWeather {
		if w := weather.FormatWeatherLine(weather.GetWeather(ctx, ev.City, ev.Start, ev.End)); w != "" {
			data["weather"] = w
		}
	}

	out, err := template.EvaluateTemplate(message.Notification, data)
	if err != nil {
		log.FromContext(ctx).Named("formatNotification").
			Error("failed to evaluate notification template", zap.Error(err))
		return ""
	}
	return out
}
