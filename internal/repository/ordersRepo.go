package repository

import (
	"moonwalk/internal/app"
	"moonwalk/pkg"
	"time"

	log "github.com/Thanga-tamil/logger_lib"
	"gorm.io/gorm"
)

func Save(o *pkg.Order) error {
	return app.DB.Table("orders").Create(o).Error
}

func GetPendingOrders() ([]pkg.Order, error) {
	var orders []pkg.Order

	err := app.DB.Table("orders").Where("status = ?", "PENDING").
		Order("created_at ASC").Find(&orders).Error

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return orders, nil
}

func GetResourceAwareOrders(status string) (pkg.Order, error) {
	var order pkg.Order

	err := app.DB.Table("orders").
		Where("status = ? AND alg = ?", status, "RESOURCE AWARE").
		Order("created_at ASC limit 1").
		Find(&order).
		Error

	if err != nil {
		log.Error(err.Error())
		return pkg.Order{}, err
	}

	return order, nil
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

func GetPreparingOrdersPastETA(tx *gorm.DB) ([]pkg.Order, error) {
	var orders []pkg.Order

	err := tx.Table("orders").
		Where("status in (?, ?) AND eta <= ?", "PROCESSING", "SERVING", time.Now()).
		Find(&orders).Error

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return orders, nil
}

func UpdateOrderStatus(tx *gorm.DB, orderId, status string, servedAt time.Time) error {
	return tx.Table("orders").
		Where("order_id = ?", orderId).
		Updates(map[string]interface{}{
			"status":    status,
			"served_at": servedAt,
		}).Error
}

func UpdateOrderStatusAndResourceId(tx *gorm.DB, orderId, status string, resourceId int) error {
	return tx.Table("orders").
		Where("order_id = ?", orderId).
		Updates(map[string]interface{}{
			"status":      status,
			"resource_id": resourceId,
		}).Error
}

// GetPendingBacklog returns the estimated queued work for orders still
// awaiting a resource, split by resource type:
//   - fifoCount       : number of PENDING pre-cooked (FIFO) orders queued for servers
//   - resourceMinutes : total prep minutes of PENDING non pre-cooked (resource
//     aware) orders queued for chefs
//
// This lets the scheduler produce a backlog-aware ETA instead of assuming a
// free kitchen, per the "current backlog and resources" requirement.
func GetPendingBacklog() (fifoCount, resourceMinutes int, err error) {
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

	log.Debug("Pending backlog: fifoCount = ", fifo, ", resourceMinutes = ", minutes)
	return int(fifo), int(minutes), nil
}

// update orders to ready based on ETA and retrieve the updated records.
func UpdateResourceAwareOrdersToReady(status string, eta time.Time) ([]pkg.Order, error) {
	var orders []pkg.Order

	if err := app.DB.Table("orders").
		Where("status = ? AND eta <= ?", "PREPARING", eta).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return orders, nil
	}

	ids := make([]string, 0, len(orders))
	for _, order := range orders {
		ids = append(ids, order.OrderId)
	}

	if err := app.DB.Table("orders").
		Where("order_id IN ?", ids).
		Updates(map[string]interface{}{
			"status": status,
		}).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func UpdateResourceAwareOrdersStatusToServing(tx *gorm.DB, orderId string) error {
	return tx.Table("orders").
		Where("order_id = ?", orderId).
		Updates(map[string]interface{}{
			"status": "SERVING",
		}).Error
}
