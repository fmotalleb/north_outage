package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func init() {
	RegisterEndpoint(
		func(api *echo.Echo) {
			api.GET("/api/up", up)
		},
	)
}

func up(c echo.Context) error {
	return c.String(http.StatusOK, "1")
}
