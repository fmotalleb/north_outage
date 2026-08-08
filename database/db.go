package database

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/fmotalleb/north_outage/database/driver"
)

var (
	rootDB *gorm.DB
	dbLog  = zap.NewExample().Named("database")
)

func Connect(connection string) (*gorm.DB, error) {
	if rootDB != nil {
		dbLog.Debug("returning cached connection")
		return rootDB, nil
	}
	dbLog.Debug("establishing connection")
	var conn gorm.Dialector
	var db *gorm.DB
	var err error
	if conn, err = driver.MakeConnection(connection); err != nil {
		dbLog.Error("failed to create dialect", zap.Error(err))
		return nil, err
	}
	if db, err = gorm.Open(conn, &gorm.Config{}); err != nil {
		dbLog.Error("failed to open connection", zap.Error(err))
		return nil, err
	}
	if err = db.Use(tracingPlugin{}); err != nil {
		dbLog.Error("failed to install tracing plugin", zap.Error(err))
		return nil, err
	}
	dbLog.Debug("connection established")
	rootDB = db
	return db, nil
}

// Get requires db to be [Connect]ed first.
// if called before a successful [Connect] will panic.
func Get() *gorm.DB {
	if rootDB == nil {
		panic("database is not initialized")
	}
	return rootDB
}
