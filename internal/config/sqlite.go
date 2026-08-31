package config

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	log "github.com/Thanga-tamil/logger_lib"
)

func NewSqlite(driverName, dataSourceName string) (*gorm.DB, error) {
	log.Infof("Initialize sqlite db")

	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Infox("Sqlite connection established and loaded in service in-memory successfully")
	return db, nil
}
