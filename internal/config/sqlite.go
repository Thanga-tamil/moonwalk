package config

import (
	"time"

	log "github.com/Thanga-tamil/logger_v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewSqlite opens a connection to the SQLite database with a tuned connection
// pool so the server can run for a long time without leaking connections.
// maxOpenConns of 0 keeps the driver default; connMaxLifetime of 0 leaves it
// unlimited (SQLite has a single writer, but the pool still bounds open FDs).
func NewSqlite(driverName, dataSourceName string, maxIdleConns, maxOpenConns, connMaxLifetime int) (*gorm.DB, error) {
	log.Infof("Initialize sqlite db")

	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if maxOpenConns > 0 {
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
	}

	if connMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
	}

	log.Infox("Sqlite connection established and loaded in service in-memory successfully")
	return db, nil
}
