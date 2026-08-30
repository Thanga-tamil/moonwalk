package repository

import (
	"moonwalk/internal/config"
	"time"

	log "github.com/Thanga-tamil/logger_lib"
)

type Dishes struct {
	Dish          string
	AvailableUpto time.Time
	CreatedAt     time.Time
}

func GetAvailableDishes() []Dishes {
	rows, err := config.DB.Query(`SELECT dish, available_upto, created_at FROM dishes;`)

	if err != nil {
		log.Error(err.Error()); return nil
	}
	defer rows.Close()

	var dishes []Dishes

	for rows.Next() {
		var dish Dishes

		if err := rows.Scan(&dish.Dish, &dish.AvailableUpto, &dish.CreatedAt); err != nil {
			log.Error(err.Error()); return nil
		}
		dishes = append(dishes, dish)
	}

	if err := rows.Err(); err != nil {
		log.Error(err.Error()); return nil
	}

	return dishes
}
