package config

import (
	"database/sql"
	log "github.com/Thanga-tamil/logger_lib"
	_ "github.com/glebarez/go-sqlite"
)

func NewSqlite(driverName, dataSourceName string) (*sql.DB, error) {
	log.Infof("Initialize sqlite db")

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	log.Infox("Sqlite connection established and loaded in service in-memory successfully")
	return db, nil
}
