package repository

import (
	"database/sql"
	"moonwalk/internal/config"
	log "github.com/Thanga-tamil/logger_lib"
)

func GetAvailableDishes(page, size int) (*sql.Rows, error) {
	rows, err := config.DB.Query(`SELECT dish, available_upto, created_at FROM dishes LIMIT $1 OFFSET $2;`, size, page)

	if err != nil {
		log.Error(err.Error()); return nil, err
	}

	return rows, nil
}

func TotalRecordsOfAvailableDishes() (int, error) {
	rows, err := config.DB.Query(`SELECT count(id) FROM dishes;`)

	if err != nil {
		log.Error(err.Error()); return -1, err
	}

	var totalRecords int

	rows.Scan(&totalRecords)

	return totalRecords, nil
}
