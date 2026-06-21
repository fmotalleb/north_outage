package driver

import (
	"net/url"
	"os"
	"path"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//# go:build orm-sqlite

func init() {
	builders["sqlite"] = sqliteBuilder
}

func sqliteBuilder(c *url.URL) (gorm.Dialector, error) {
	fullPath := path.Join(c.Hostname(), c.Path)
	parent := path.Dir(fullPath)
	if _, err := os.ReadDir(parent); err != nil {
		return nil, err
	}
	return sqlite.Open(fullPath), nil
}
