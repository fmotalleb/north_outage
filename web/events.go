package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"

	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/models"
)

func init() {
	RegisterEndpoint(
		func(api *echo.Echo) {
			api.GET("/api/events", events)
			api.GET("/api/updated_at", updatedAt)
		},
	)
}

func events(c echo.Context) error {
	db := database.
		Get().
		WithContext(c.Request().Context()).
		Table("events")

	city := c.QueryParam("city")
	if city != "" {
		db = db.Where("city = ?", city)
	}

	search := c.QueryParam("search")
	if search != "" {
		db = db.Where("address LIKE ?", "%"+search+"%")
	}

	limit := c.QueryParam("limit")
	lv := cast.ToInt(limit)
	if lv > 0 {
		db = db.Limit(lv)
	}

	var events []models.Event
	result := db.Find(&events)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
	}
	if len(events) == 0 {
		return c.JSON(http.StatusNotFound, events)
	}
	return c.JSON(http.StatusOK, events)
}

func updatedAt(c echo.Context) error {
	db := database.
		Get().
		WithContext(c.Request().Context()).
		Table("events")

	var createdAt time.Time

	err := db.
		Order("created_at DESC").
		Pluck("created_at", &createdAt).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"created_at": createdAt,
	})
}
