package repository

import (
	"moonwalk/internal/app"
	"moonwalk/pkg"

	"gorm.io/gorm"
)

func GetResources() (*[]pkg.Resources, error) {
	var resources []pkg.Resources

	if err := app.DB.Table("resources").Find(&resources).Error; err != nil {
		return nil, err
	}

	return &resources, nil
}

func GetSuppliers() (*[]pkg.Resources, error) {
	var resources []pkg.Resources

	if err := app.DB.Table("resources").Where("order_handling_type = ?", true).Find(&resources).Error; err != nil {
		return nil, err
	}

	return &resources, nil
}

func UpdateResourceStatus(tx *gorm.DB, resourceId int, status string, currentOrderId string) error {
	return tx.Table("resources").
		Where("id = ?", resourceId).
		Updates(map[string]interface{}{
			"status":           status,
			"current_order_id": currentOrderId,
		}).Error
}

func UpdateSupplierStatusToBusy(tx *gorm.DB, resource *pkg.Resources, status, orderId string) error {
	return tx.Table("resources").
		Where("id = ?", resource.Id).
		Updates(map[string]interface{}{
			"status":           status,
			"current_order_id": orderId,
		}).Error
}

func UpdateChefStatusToIdle(tx *gorm.DB, status string, resourceId int) error {
	return tx.Table("resources").
		Where("id = ?", resourceId).
		Updates(map[string]interface{}{
			"status":           status,
			"current_order_id": "",
		}).Error
}

func FindResourceByOrderId(tx *gorm.DB, orderId string) int {
	var resource pkg.Resources

	if err := tx.Table("resources").Where("current_order_id = ?", orderId).First(&resource).Error; err != nil {
		return 0
	}

	return resource.Id
}
