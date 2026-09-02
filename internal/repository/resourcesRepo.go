package repository

import (
	"moonwalk/pkg"
	"moonwalk/internal/app"
)

func GetResources() (*[]pkg.Resources, error) {
	var resources []pkg.Resources

	if err := app.DB.Table("resources").Find(&resources).Error; err != nil {
		return nil, err
	}

	return &resources, nil
}

func UpdateResourceStatus(resourceId int, status string, currentOrderId string) error {
	err := app.DB.Table("resources").
		Where("id = ?", resourceId).
		Updates(map[string]interface{}{
			"status":           status,
			"current_order_id": currentOrderId,
		}).Error

	if err != nil {
		return err
	}
	return nil
}

