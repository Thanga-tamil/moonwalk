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

func GetOrder(orderId string) (pkg.Order, error) {
	var order pkg.Order

	err := app.DB.Table("orders").
		Where("order_id = ?", orderId).
		First(&order).Error

	if err != nil {
		log.Error(err.Error())
		return order, err
	}

	return order, nil
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

// GetPendingBacklog returns the estimated queued work for orders still
// awaiting a resource, split by resource type:
//   - fifoCount       : number of PENDING pre-cooked (FIFO) orders queued for servers
//   - resourceMinutes : total prep minutes of PENDING non pre-cooked (resource
//     aware) orders queued for chefs
//
// This lets the scheduler produce a backlog-aware ETA instead of assuming a
// free kitchen, per the "current backlog and resources" requirement.
func GetPendingBacklog() (fifoCount int, resourceMinutes int, err error) {
	var fifo int64
	err = app.DB.Table("orders AS o").
		Joins("JOIN dishes d ON o.dish_id = d.id").
		Where("o.status = ? AND d.pre_cooked = ?", "PENDING", true).
		Count(&fifo).Error
	if err != nil {
		log.Error(err.Error())
		return 0, 0, err
	}

	var minutes int64
	err = app.DB.Table("orders AS o").
		Joins("JOIN dishes d ON o.dish_id = d.id").
		Where("o.status = ? AND d.pre_cooked = ?", "PENDING", false).
		Select("COALESCE(SUM(d.prep_time), 0)").
		Scan(&minutes).Error
	if err != nil {
		log.Error(err.Error())
		return 0, 0, err
	}

	return int(fifo), int(minutes), nil
}

