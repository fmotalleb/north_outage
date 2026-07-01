package mattermost

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fmotalleb/go-tools/log"

	"github.com/fmotalleb/north_outage/config"
	im "github.com/fmotalleb/north_outage/models"
)

var activeCfg *config.Config

func Setup(ctx context.Context, cfg *config.Config) {
	activeCfg = cfg
	registerRoutes(ctx, cfg)
}

func Run(ctx context.Context, cfg *config.Config, nc <-chan im.Notification) error {
	l := log.Of(ctx).Named("Mattermost")
	ctx = log.WithLogger(ctx, l)
	if cfg.Mattermost.BotToken == "" || cfg.Mattermost.ServerURL == "" {
		l.Warn("mattermost bot token or server url is not set")
		return nil
	}

	client := &http.Client{
		Timeout: cfg.Mattermost.Timeout,
	}
	go bindToChannel(ctx, l, client, cfg, nc)
	<-ctx.Done()
	return nil
}

func apiBase() *url.URL {
	if activeCfg == nil {
		return nil
	}
	if activeCfg.Mattermost.ServerURL == "" {
		return nil
	}
	u, err := url.Parse(activeCfg.Mattermost.ServerURL)
	if err != nil {
		return nil
	}
	return u
}

func publicBase() *url.URL {
	if activeCfg == nil {
		return nil
	}
	if activeCfg.Mattermost.PublicURL == "" {
		return nil
	}
	u, err := url.Parse(activeCfg.Mattermost.PublicURL)
	if err != nil {
		return nil
	}
	return u
}

func joinURL(base *url.URL, path string) string {
	if base == nil {
		return ""
	}
	cp := *base
	cp.Path = strings.TrimRight(cp.Path, "/") + path
	return cp.String()
}

func clientWithTimeout(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{Timeout: timeout}
}
