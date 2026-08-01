package web

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

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
	ctx, l := log.AsNamedChild(ctx, "Web")
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
	// Stop the HTTP server when the shared context is canceled (i.e. on the
	// sys signal at the app entry point). Without this server.Start blocks
	// forever and the process would never shut down cleanly.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			l.Error("web server shutdown failed", zap.Error(err))
		}
	}()
	if err := server.Start(cfg.HTTPListenAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
