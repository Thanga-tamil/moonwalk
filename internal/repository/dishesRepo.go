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
				  Select(`id, dish, is_available, price, available_upto, created_at, (prep_time || ' minutes') AS prep_time`).
				  Limit(size).Offset(offset).
				  Scan(&dishes).
				  Error

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return dishes, nil
}

func TotalRecordsOfAvailableDishes() (int64, error) {
	var totalRecords int64

	err := app.DB.Table("dishes").Count(&totalRecords).Error

	if err != nil {
		log.Error(err.Error())
		return -1, err
	}

	return totalRecords, nil
}

func IsDishExists(dishID int) (bool, error) {
	var exists bool

	err := app.DB.Raw("SELECT EXISTS (SELECT 1 FROM dishes WHERE id = ?)", dishID).Scan(&exists).Error

	if err != nil {
		log.Error(err.Error())
		return false, err
	}

	return exists, nil
}
