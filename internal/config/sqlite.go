package config

import (
	"database/sql"
	log "github.com/Thanga-tamil/logger_lib"
	_ "github.com/glebarez/go-sqlite"
)

var DB *sql.DB

func NewSqlite() error {
	log.Infof("Initialize sqlite db")

	db, err := sql.Open("sqlite", "/home/milrine/projects/MoonWalk/moonwalk.db")
	if err != nil {
		return err
	}

	DB = db

	log.Info("Sqlite connection established and loaded in service in-memory successfully")
	return nil
}
