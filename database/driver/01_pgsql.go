package driver

import (
	"cmp"
	"fmt"
	"net/url"
	"path/filepath"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//# go:build orm-sqlite

func init() {
	builders["postgres"] = posgreSQLBuilder
}

func posgreSQLBuilder(c *url.URL) (gorm.Dialector, error) {
	var pass string
	pass, _ = c.User.Password()
	query := c.Query()
	sslMode := cmp.Or(query.Get("sslmode"), "disable")
	tz := cmp.Or(query.Get("timeZone"), "UTC")
	db := filepath.Base(c.Path)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.Hostname(),
		c.User.Username(),
		pass,
		db,
		c.Port(),
		sslMode,
		tz,
	)
	return postgres.Open(dsn), nil
}
