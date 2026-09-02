package repository

import (
	"time"
	"moonwalk/internal/app"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_lib"
)

func Save(o *pkg.Order) error {
	if err := app.DB.Table("orders").Create(o).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func GetPendingOrders() ([]pkg.Order, error) {
	var orders []pkg.Order

	err := app.DB.Table("orders").
		Where("status = ?", "PENDING").
		Order("created_at ASC").
		Find(&orders).Error

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return orders, nil
}

func GetPreparingOrdersPastETA() ([]pkg.Order, error) {
	var orders []pkg.Order

	err := app.DB.Table("orders").
		Where("status = ? AND eta <= ?", "PREPARING", time.Now()).
		Find(&orders).Error

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return orders, nil
}

func UpdateOrderStatus(orderId string, status string) error {
	err := app.DB.Table("orders").
		Where("order_id = ?", orderId).
		Update("status", status).Error

	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

