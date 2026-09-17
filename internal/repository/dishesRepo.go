package repository

import (
	"moonwalk/internal/app"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_v2"
)

func GetPaginatedDishes(page, size int) ([]*pkg.Dish, error) {
	offset := (page - 1) * size

	var dishes []*pkg.Dish

	err := app.DB.Table("dishes").
		Select(`id, dish, price, prep_time, ` +
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

func GetDish(dishID int) (pkg.Dish, error) {
	var dish pkg.Dish

	err := app.DB.Raw("SELECT * FROM dishes WHERE id = ? AND is_available = ?", dishID, true).Scan(&dish).Error

	if err != nil {
		log.Error(err.Error())
		return dish, err
	}

	return dish, nil
}

func InsertDish(d *pkg.Dish) error {
	return app.DB.Table("dishes").Create(d).Error
}

func InsertDishes(d []*pkg.Dish) error {
	return app.DB.Table("dishes").Create(d).Error
}

// a dish have unique constraint in schema level
// if a client wants to delete a dish: do hard delete
// so this way if client wants to add the same dish again
// /add dish API can be used to persist the data in schema
// without unique constraint error
func DeleteDishes(ids []int) (int64, error) {
	result := app.DB.Table("dishes").Where("id IN ?", ids).Delete("")

	if result.Error != nil {
		return -1, result.Error
	}

	return result.RowsAffected, nil
}
