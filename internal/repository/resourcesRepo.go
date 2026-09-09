package repository

import (
	"moonwalk/internal/app"
	"moonwalk/pkg"
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

func UpdateResourceStatus(resourceId int, status string, currentOrderId string) error {
	return app.DB.Table("resources").
		Where("id = ?", resourceId).
		Updates(map[string]interface{}{
			"status":           status,
			"current_order_id": currentOrderId,
		}).Error
}

func UpdateSupplierStatusToBusy(resource *pkg.Resources, status, orderId string) error {
	return app.DB.Table("resources").
		Where("id = ?", resource.Id).
		Updates(map[string]interface{}{
			"status":           status,
			"current_order_id": orderId,
		}).Error
}

func UpdateChefStatusToIdle(status string, resourceId int) error {
	return app.DB.Table("resources").
		Where("id = ?", resourceId).
		Updates(map[string]interface{}{
			"status":           status,
			"current_order_id": "",
		}).Error
}
