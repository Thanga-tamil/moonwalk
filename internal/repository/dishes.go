package repository

import (
	"database/sql"
	"moonwalk/internal/config"

	log "github.com/Thanga-tamil/logger_lib"
)

func GetAvailableDishes() *sql.Rows {
	rows, err := config.DB.Query(`SELECT dish, available_upto, created_at FROM dishes;`)

	if err != nil {
		log.Error(err.Error()); return nil
	}

	return rows
}
