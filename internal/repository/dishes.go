package repository

import (
	"database/sql"
	"moonwalk/internal/config"
	log "github.com/Thanga-tamil/logger_lib"
)

func GetAllDishes(page, size int) (*sql.Rows, error) {
	offset := (page - 1) * size

	q := `SELECT dish, is_available, price, available_upto, created_at FROM dishes LIMIT $1 OFFSET $2;`

	rows, err := config.DB.Query(q, size, offset)

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return rows, nil
}

func TotalRecordsOfAvailableDishes() (int, error) {
	var totalRecords int

	err := config.DB.QueryRow(`SELECT COUNT(id) FROM dishes;`).Scan(&totalRecords)

	if err != nil {
		log.Error(err.Error())
		return -1, err
	}

	return totalRecords, nil
}
