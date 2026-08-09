package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"

	"github.com/fmotalleb/north_outage/database"
	"github.com/fmotalleb/north_outage/models"
)

const (
	// defaultEventLimit caps the number of events returned when no limit
	// query parameter is supplied, so a bare request cannot dump the whole
	// table.
	defaultEventLimit = 1000
	// maxEventLimit is the hard cap for an explicit limit parameter.
	maxEventLimit = 10000
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
		db = db.Where("address LIKE ? ESCAPE '\\'", "%"+escapeLike(search)+"%")
	}

	limit := cast.ToInt(c.QueryParam("limit"))
	if limit <= 0 {
		limit = defaultEventLimit
	}
	if limit > maxEventLimit {
		limit = maxEventLimit
	}
	db = db.Limit(limit)

	var events []models.Event
	result := db.Find(&events)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": result.Error.Error()})
	}
	return c.JSON(http.StatusOK, events)
}

// escapeLike escapes LIKE wildcards in user input so search terms are matched
// literally instead of acting as pattern characters.
func escapeLike(s string) string {
	if !strings.ContainsAny(s, `\%_`) {
		return s
	}
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
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
