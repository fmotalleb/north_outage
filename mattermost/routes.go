package mattermost

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/fmotalleb/north_outage/config"
	"github.com/fmotalleb/north_outage/web"
)

func registerRoutes(_ context.Context, _ *config.Config) {
	web.RegisterEndpoint(func(e *echo.Echo) {
		g := e.Group("/api/mattermost")
		g.POST("/command", commandHandler)
		g.POST("/action", actionHandler)
	})
}
