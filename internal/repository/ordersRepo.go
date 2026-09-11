package repository

import (
	"moonwalk/internal/app"
	"moonwalk/pkg"
	"time"

	log "github.com/Thanga-tamil/logger_v2"
	"gorm.io/gorm"
)

func Insert(tx *gorm.DB, o *pkg.Order) error {
	return tx.Table("orders").Create(o).Error
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

func UpdateOrder(tx *gorm.DB, order *pkg.Order) error {
	return tx.Model(&pkg.Order{}).
		Where("order_id = ?", order.OrderId).
		Updates(order).Error
}

// update order status and return the updated records for audit transition
func UpdateResourceAwareOrdersStatusByETA(tx *gorm.DB, alg, currentStatus, nextStatus string, eta time.Time) (*[]pkg.Order, error) {
	var orders []pkg.Order

	err := tx.Table("orders").
		Where("alg = ? and status = ? and eta <= ?", alg, currentStatus, eta).
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, o := range orders {
		ids = append(ids, o.OrderId)
	}

	result := tx.Table("orders").
		Where("order_id IN ?", ids).
		Updates(map[string]interface{}{
			"status":     nextStatus,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return nil, result.Error
	}

	log.Infof("%d resource aware orders updated to status: %s", result.RowsAffected, nextStatus)

	err = tx.Table("orders").Where("order_id IN ?", ids).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return &orders, nil
}

func FetchResourceAwareOrdersByStatus(tx *gorm.DB, status string) ([]pkg.Order, error) {
	var orders []pkg.Order

	err := tx.Table("orders").Where("status = ?", status).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func UpdateFifoProcessingOrders(tx *gorm.DB, alg, currentStatus, nextStatus string, eta time.Time) ([]string, error) {

	var orders []pkg.Order

	err := tx.Table("orders").
		Where("alg = ? and status = ? and eta <= ?", alg, currentStatus, eta).
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, o := range orders {
		ids = append(ids, o.OrderId)
	}

	updatedOrders := tx.Table("orders").
		Where("order_id in ?", ids).
		Updates(map[string]interface{}{
			"status":     nextStatus,
			"updated_at": time.Now(),
		})

	if updatedOrders.Error != nil {
		return nil, updatedOrders.Error
	}

	log.Infof("%d Fifo orders updated to status: %s", updatedOrders.RowsAffected, nextStatus)

	return ids, nil
}

func FetchPendingResourceAwareOrders(tx *gorm.DB, status string) ([]pkg.Order, error) {
	var orders []pkg.Order

	err := tx.Table("orders").Where("status = ?", status).Find(&orders).Order("ORDER BY eta ASC").Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}
