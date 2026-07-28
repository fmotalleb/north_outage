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

	l.Debug("creating telegram bot client")
	var opts []bot.Option
	client := httpClient(tel.Proxy)

	hc := bot.WithHTTPClient(time.Second*30, client)
	opts = append(opts, hc)

	b, err := bot.New(tel.BotKey, opts...)
	if err != nil {
		l.Error("failed to connect to telegram bot", zap.Error(err))
		return err
	}
	l.Debug("telegram bot connected, setting up handlers")
	handlers.SetupHandlers(ctx, b)
	go bindToChannel(ctx, b, nc)
	l.Debug("starting telegram bot polling")
	b.Start(ctx)
	return nil
}

func bindToChannel(ctx context.Context, b *bot.Bot, nc <-chan im.Notification) {
	l := log.Of(ctx).Named("binder")
	l.Debug("notification binder started")
	for {
		select {
		case n := <-nc:
			l.Debug("notification received from channel",
				zap.String("message", n.Message),
				zap.Int64("chat_id", n.Listener.TelegramCID),
			)
			sp := new(bot.SendMessageParams)
			sp.ChatID = n.Listener.TelegramCID
			sp.MessageThreadID = int(n.Listener.TelegramTID)
			cfg, err := config.Get(ctx)
			notifyWeather := err == nil && cfg.Weather.Notify
			sp.Text = formatNotification(ctx, n.Message, n.Event, notifyWeather)
			l.Debug("sending notification message to telegram")
			m, err := b.SendMessage(ctx, sp)
			if err != nil {
				l.Error("failed to send notification to telegram", zap.Error(err))
			} else {
				l.Debug("notification sent to telegram", zap.Int("message_id", m.ID))
			}
		case <-ctx.Done():
			l.Debug("notification binder shutting down")
			return
		}
	}
}

func formatNotification(ctx context.Context, msg string, ev *im.Event, notifyWeather bool) string {
	l := log.FromContext(ctx).Named("formatNotification")
	data := map[string]any{
		"message": msg,
		"event":   ev,
	}

	if notifyWeather {
		if w := weather.FormatWeatherLine(weather.GetWeather(ctx, ev.City, ev.Start, ev.End)); w != "" {
			data["weather"] = w
			l.Debug("weather data appended to notification", zap.String("city", ev.City))
		}
	}

	l.Debug("evaluating notification template")
	out, err := template.EvaluateTemplate(message.Notification, data)
	if err != nil {
		l.Error("failed to evaluate notification template", zap.Error(err))
		return ""
	}
	l.Debug("notification template evaluated")
	return out
}
