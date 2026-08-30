package service

import (
	"time"
	"moonwalk/internal/repository"
)

type Dish struct {
	Dish          string
	AvailableUpto time.Time
	CreatedAt     time.Time
}

func GetAvailableDishes() ([]Dish, int) {

	// Retrieve available dishes from db 
	rows := repository.GetAvailableDishes()

	var dishes []Dish

	defer rows.Close()

	for rows.Next() {
		
		var dish Dish 
		rows.Scan(&dish.Dish, &dish.AvailableUpto, &dish.CreatedAt)
		
		dishes = append(dishes, dish)
	}

	// return list of dish and totol records of dishes
	return dishes, len(dishes)
}
