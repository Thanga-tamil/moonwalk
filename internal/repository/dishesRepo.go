package repository

import (
	"database/sql"
	"moonwalk/internal/app"
	log "github.com/Thanga-tamil/logger_lib"
)

func GetAllDishes(page, size int) (*sql.Rows, error) {
	offset := (page - 1) * size

	q := `SELECT id, dish, is_available, price, available_upto, ` +
		 `created_at, concat(prep_time, " minutes") AS prep_time ` + 
		 `FROM dishes LIMIT $1 OFFSET $2;`

	log.Debugf("Query: %s ? LIMIT: %d OFFSET: %d", q, size, offset)

	rows, err := app.DB.Query(q, size, offset)

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return rows, nil
}

func TotalRecordsOfAvailableDishes() (int, error) {
	var totalRecords int

	err := app.DB.QueryRow(`SELECT COUNT(id) FROM dishes;`).Scan(&totalRecords)

	if err != nil {
		log.Error(err.Error())
		return -1, err
	}

	return totalRecords, nil
}

func IsDishExists(dishId int) (int, error) {

	var exists int
	err := app.DB.QueryRow(`SELECT EXISTS (SELECT id FROM dishes WHERE id = $1) AS IsDishExists;`, dishId).Scan(&exists)

	if err != nil {
		log.Error(err.Error())
		return -1, err
	}

	return exists, nil
}
