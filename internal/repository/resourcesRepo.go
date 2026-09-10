package repository

import (
	"moonwalk/internal/app"
	"moonwalk/pkg"

	"gorm.io/gorm"
)

func GetResources() (*[]pkg.Resources, error) {
	var resources []pkg.Resources

	if err := app.DB.Table("resources").Where("status = ?", "IDLE").Find(&resources).Order("updated_at ASC").Error; err != nil {
		return nil, err
	}

	return &resources, nil
}

// -- RESOURE AWARE: find the CHEF whom will complete the cooking first and asign a new order to this CHEF in PENDING status
// -- so called: backlog
// select * from orders order by eta asc limit 1;
func FindResource(tx *gorm.DB, isPreCooked bool) (*pkg.Resources, error) {
	var resource pkg.Resources
	var q string

	if isPreCooked {
		q = `select * from resources as r
			 where type = 'SUPPLIER' and r.status = 'IDLE' 
			 or (select resource_id from orders where status in ('PENDING', 'PROCESSING') 
			 order by eta asc limit 1) order by updated_at asc limit 1;`
	} else {
		q = `select * from resources as r
			 where type = 'CHEF' and 
			 ( r.status = 'IDLE' or 
			 (select resource_id from orders where status in ('PENDING', 'PREPARING') order by eta asc limit 1)
			 ) order by updated_at asc limit 1;`
	}
	if err := tx.Raw(q).Scan(&resource).Error; err != nil {
		return nil, err
	}

	return &resource, nil
}

func GetSuppliers() ([]pkg.Resources, error) {
	var resources []pkg.Resources

	if err := app.DB.Table("resources").
		Where("type = ? and order_handling_type = ?", "SUPPLIER", true).Find(&resources).
		Order("updated_at ASC").
		Error; err != nil {
		return nil, err
	}

	return resources, nil
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
