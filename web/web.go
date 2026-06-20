package web

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/fmotalleb/north_outage/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var server = echo.New()

func RegisterEndpoint(register func(*echo.Echo)) {
	register(server)
}

func init() {
	server.Use(middleware.RequestLogger())
	server.Use(middleware.Recover())
}

func Start(ctx context.Context, cfg *config.Config) error {
	if cfg.HTTPListenAddr == "" {
		return nil
	}
	server.Server = &http.Server{
		ReadTimeout:       time.Minute,
		ReadHeaderTimeout: time.Minute,
		IdleTimeout:       time.Minute,
		WriteTimeout:      time.Minute,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	if err := server.Start(cfg.HTTPListenAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
