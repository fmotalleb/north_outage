package mattermost

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fmotalleb/go-tools/git"
	"github.com/fmotalleb/go-tools/log"
	tgmodels "github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/internal/template"
	nmodels "github.com/fmotalleb/north_outage/models"
	"github.com/fmotalleb/north_outage/telegram/message"
	"github.com/fmotalleb/north_outage/weather"
	"github.com/labstack/echo/v4"
)

type slashCommandRequest struct {
	Token       string `json:"token"`
	TeamID      string `json:"team_id"`
	TeamDomain  string `json:"team_domain"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Command     string `json:"command"`
	Text        string `json:"text"`
	ResponseURL string `json:"response_url"`
	TriggerID   string `json:"trigger_id"`
	PostID      string `json:"post_id"`
}

type actionRequest struct {
	UserID    string         `json:"user_id"`
	PostID    string         `json:"post_id"`
	ChannelID string         `json:"channel_id"`
	TeamID    string         `json:"team_id"`
	Context   map[string]any `json:"context"`
}

type mattermostResponse struct {
	ResponseType string         `json:"response_type"`
	Text         string         `json:"text,omitempty"`
	Username     string         `json:"username,omitempty"`
	IconURL      string         `json:"icon_url,omitempty"`
	Attachments  []mmAttachment `json:"attachments,omitempty"`
	Props        map[string]any `json:"props,omitempty"`
}

type mmAttachment struct {
	Fallback string     `json:"fallback,omitempty"`
	Color    string     `json:"color,omitempty"`
	Pretext  string     `json:"pretext,omitempty"`
	Text     string     `json:"text,omitempty"`
	Actions  []mmAction `json:"actions,omitempty"`
}

type mmAction struct {
	Name        string        `json:"name"`
	Text        string        `json:"text,omitempty"`
	Type        string        `json:"type,omitempty"`
	Integration mmIntegration `json:"integration"`
}

type mmIntegration struct {
	URL     string         `json:"url"`
	Context map[string]any `json:"context"`
}

func commandHandler(c echo.Context) error {
	req, err := parseSlashRequest(c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, mattermostResponse{
			ResponseType: "ephemeral",
			Text:         "invalid command payload",
		})
	}

	l := log.Of(c.Request().Context()).Named("commandHandler")
	l.Debug("mm slash command",
		zap.String("cmd", req.Command),
		zap.String("text", req.Text),
		zap.String("user", req.UserName),
		zap.String("channel", req.ChannelName),
	)

	if cfg := activeCfg; cfg != nil && cfg.Mattermost.CommandToken != "" && req.Token != cfg.Mattermost.CommandToken {
		l.Warn("unauthorized mm command", zap.String("token", req.Token))
		return c.JSON(http.StatusUnauthorized, mattermostResponse{
			ResponseType: "ephemeral",
			Text:         "unauthorized",
		})
	}

	cmd := commandName(req.Command)
	l.Debug("dispatching mm command", zap.String("command", cmd))

	switch cmd {
	case "help", "start":
		return c.JSON(http.StatusOK, makeResponse(req, helpText(req)))
	case "version":
		return c.JSON(http.StatusOK, makeResponse(req, git.String()))
	case "search":
		return c.JSON(http.StatusOK, handleSearch(req))
	case "listen":
		return c.JSON(http.StatusOK, handleDirectListen(req))
	default:
		return c.JSON(http.StatusOK, makeResponse(req, helpText(req)))
	}
}

func actionHandler(c echo.Context) error {
	req, err := parseActionRequest(c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, mattermostResponse{
			ResponseType: "ephemeral",
			Text:         "invalid action payload",
		})
	}

	action := strings.ToLower(toString(req.Context["action"]))
	l := log.FromContext(c.Request().Context()).Named("actionHandler")
	l.Debug("mm action received",
		zap.String("action", action),
		zap.String("user", req.UserID),
		zap.String("channel", req.ChannelID),
	)

	switch action {
	case "listen":
		return c.JSON(http.StatusOK, handleListenAction(req))
	default:
		return c.JSON(http.StatusOK, mattermostResponse{
			ResponseType: "ephemeral",
			Text:         "unknown action",
		})
	}
}

func parseSlashRequest(r *http.Request) (*slashCommandRequest, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	req := &slashCommandRequest{
		Token:       r.FormValue("token"),
		TeamID:      r.FormValue("team_id"),
		TeamDomain:  r.FormValue("team_domain"),
		ChannelID:   r.FormValue("channel_id"),
		ChannelName: r.FormValue("channel_name"),
		UserID:      r.FormValue("user_id"),
		UserName:    r.FormValue("user_name"),
		Command:     r.FormValue("command"),
		Text:        strings.TrimSpace(r.FormValue("text")),
		ResponseURL: r.FormValue("response_url"),
		TriggerID:   r.FormValue("trigger_id"),
		PostID:      r.FormValue("post_id"),
	}
	return req, nil
}

func parseActionRequest(r *http.Request) (*actionRequest, error) {
	payload, err := readActionPayload(r)
	if err != nil {
		return nil, err
	}
	req := new(actionRequest)
	if err := json.Unmarshal(payload, req); err != nil {
		return nil, err
	}
	return req, nil
}

func readActionPayload(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if len(body) > 0 && !strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		return body, nil
	}

	r.Body = io.NopCloser(bytes.NewReader(body))
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	payload := r.FormValue("payload")
	if payload == "" {
		if len(body) > 0 {
			return body, nil
		}
		return nil, errors.New("empty payload")
	}
	return []byte(payload), nil
}

func makeResponse(req *slashCommandRequest, text string) mattermostResponse {
	return mattermostResponse{
		ResponseType: "in_channel",
		Text:         text,
	}
}

func helpText(req *slashCommandRequest) string {
	update := syntheticUpdate(req.UserName, req.ChannelName, "")
	out, err := message.EvaluateMessageTemplate(message.Help, nil, update)
	if err != nil || out == "" {
		return "Use /search <text> to find outages."
	}
	return out
}

func handleSearch(req *slashCommandRequest) mattermostResponse {
	search := req.Text
	if search == "" {
		return mattermostResponse{
			ResponseType: "ephemeral",
			Text:         "Use /search <address>.",
		}
	}

	events, err := fetchEvents(search)
	if err != nil {
		return mattermostResponse{
			ResponseType: "ephemeral",
			Text:         "failed to fetch outage data",
		}
	}

	update := syntheticUpdate(req.UserName, req.ChannelName, search)
	data := map[string]any{"results": events}
	out, err := message.EvaluateMessageTemplate(message.Search, data, update)
	if err != nil || out == "" {
		out = "موردی یافت نشد"
	}

	resp := makeResponse(req, out)
	if len(events) > 0 {
		resp.Attachments = []mmAttachment{{
			Fallback: "listen options",
			Text:     "Select a city to start monitoring.",
			Actions:  buildActions(search, events),
		}}
	}
	return resp
}

func handleDirectListen(req *slashCommandRequest) mattermostResponse {
	parts := strings.SplitN(req.Text, "|", 2)
	if len(parts) == 2 {
		city := strings.TrimSpace(parts[0])
		search := strings.TrimSpace(parts[1])
		if city != "" && search != "" {
			return storeListener(req.ChannelID, req.PostID, city, search)
		}
	}

	fields := strings.Fields(req.Text)
	if len(fields) < 2 {
		return mattermostResponse{
			ResponseType: "ephemeral",
			Text:         "Use /listen <city> <search text>.",
		}
	}
	city := fields[0]
	search := strings.Join(fields[1:], " ")
	return storeListener(req.ChannelID, req.PostID, city, search)
}

func handleListenAction(req *actionRequest) mattermostResponse {
	city := toString(req.Context["city"])
	search := toString(req.Context["search"])
	if city == "" || search == "" {
		return mattermostResponse{
			ResponseType: "ephemeral",
			Text:         "This action expired. Please search again.",
		}
	}
	return storeListener(req.ChannelID, req.PostID, city, search)
}

func storeListener(channelID, rootID, city, search string) mattermostResponse {
	db := database.Get()
	listener := new(nmodels.Listener)
	listener.City = city
	listener.SearchTerm = search
	listener.MattermostCID = channelID
	listener.MattermostRID = rootID
	if err := db.Save(listener).Error; err != nil {
		return mattermostResponse{
			ResponseType: "ephemeral",
			Text:         "failed to save listener",
		}
	}
	return mattermostResponse{
		ResponseType: "ephemeral",
		Text:         "I will notify this channel when that outage appears.",
	}
}

func fetchEvents(search string) ([]nmodels.Event, error) {
	out := make([]nmodels.Event, 0, 10)
	err := database.Get().
		Table("events").
		Where("address LIKE ?", "%"+search+"%").
		Limit(10).
		Find(&out).Error
	return out, err
}

func buildActions(search string, events []nmodels.Event) []mmAction {
	base := joinURL(publicBase(), "/api/mattermost/action")
	if base == "" {
		return nil
	}
	seen := make(map[string]struct{})
	actions := make([]mmAction, 0, len(events))
	for _, ev := range events {
		if _, ok := seen[ev.City]; ok {
			continue
		}
		seen[ev.City] = struct{}{}
		actions = append(actions, mmAction{
			Name: "🔍 " + ev.City,
			Type: "button",
			Integration: mmIntegration{
				URL: base,
				Context: map[string]any{
					"action": "listen",
					"city":   ev.City,
					"search": search,
				},
			},
		})
	}
	return actions
}

func commandName(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	cmd = strings.TrimPrefix(cmd, "/")
	if idx := strings.IndexByte(cmd, '@'); idx >= 0 {
		cmd = cmd[:idx]
	}
	return cmd
}

func syntheticUpdate(userName, channelName, text string) *tgmodels.Update {
	name := channelName
	if name == "" {
		name = userName
	}
	if name == "" {
		name = "Mattermost"
	}
	return &tgmodels.Update{
		Message: &tgmodels.Message{
			Text: text,
			Chat: tgmodels.Chat{
				Title:     name,
				FirstName: name,
			},
		},
	}
}

func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprint(v)
	}
}

func formatMMNotification(ctx context.Context, ev *nmodels.Event, notifyWeather bool) string {
	data := map[string]any{
		"event": ev,
	}

	if notifyWeather {
		if w := weather.FormatWeatherLine(weather.GetWeather(ctx, ev.City, ev.Start, ev.End)); w != "" {
			data["weather"] = w
		}
	}

	out, err := template.EvaluateTemplate(message.MMNotification, data)
	if err != nil {
		log.FromContext(ctx).Named("formatMMNotification").
			Error("failed to evaluate mattermost notification template", zap.Error(err))
		return ""
	}
	return out
}

func bindToChannel(ctx context.Context, l *zap.Logger, client *http.Client, cfg *config.Config, nc <-chan nmodels.Notification) {
	base := apiBase()
	if base == nil {
		l.Warn("mattermost server url is not configured")
		return
	}
	postURL := joinURL(base, "/api/v4/posts")
	l.Debug("mattermost notification binder started")
	for {
		select {
		case n := <-nc:
			if n.Listener == nil || n.Event == nil {
				l.Debug("skipping notification with nil listener/event")
				continue
			}
			if n.Listener.MattermostCID == "" {
				l.Debug("skipping notification without mattermost channel")
				continue
			}
			l.Debug("sending mattermost notification",
				zap.String("channel", n.Listener.MattermostCID),
				zap.String("message", n.Message),
			)
			notifyWeather := cfg.Weather.Notify
			body := map[string]any{
				"channel_id": n.Listener.MattermostCID,
				"message":    formatMMNotification(ctx, n.Event, notifyWeather),
			}
			if n.Listener.MattermostRID != "" {
				body["root_id"] = n.Listener.MattermostRID
				l.Debug("notification has root_id", zap.String("root_id", n.Listener.MattermostRID))
			}
			payload, err := json.Marshal(body)
			if err != nil {
				l.Error("failed to marshal mattermost post body", zap.Error(err))
				continue
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, postURL, bytes.NewReader(payload))
			if err != nil {
				l.Error("failed to create mattermost HTTP request", zap.Error(err))
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+cfg.Mattermost.BotToken)
			resp, err := client.Do(req)
			if err != nil {
				l.Error("failed to POST mattermost notification", zap.Error(err))
				continue
			}
			l.Debug("mattermost notification posted", zap.Int("status", resp.StatusCode))
			_ = resp.Body.Close()
		case <-ctx.Done():
			l.Debug("mattermost notification binder shutting down")
			return
		}
	}
}
