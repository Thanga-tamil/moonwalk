package repository

import (
	"moonwalk/pkg"
	"moonwalk/internal/app"
	log "github.com/Thanga-tamil/logger_lib"
)

func GetAllDishes(page, size int) ([]pkg.Dish, error) {
	offset := (page - 1) * size

	var dishes []pkg.Dish

	err := app.DB.Table("dishes").
				  Select(`id, dish, price, (prep_time || ' minutes') AS prep_time, ` +
				  		 `is_available, created_at`).
				  Limit(size).Offset(offset).
				  Scan(&dishes).
				  Error

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return dishes, nil
}

func TotalRecordsOfDishes() (int64, error) {
	var totalRecords int64

	err := app.DB.Table("dishes").Count(&totalRecords).Error

	if err != nil {
		log.Error(err.Error())
		return -1, err
	}

	return totalRecords, nil
}

func GetDish(dishID int) (*pkg.Dish, error) {
	var dish pkg.Dish

	err := app.DB.Raw("SELECT * FROM dishes WHERE id = ?", dishID).Scan(&dish).Error

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &dish, nil
}

