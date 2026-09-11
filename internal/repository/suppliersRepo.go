package repository

import (
	"moonwalk/pkg"

	"gorm.io/gorm"
)

func FindAvailableSupplier(tx *gorm.DB) (*pkg.Suppliers, error) {
	var supplier pkg.Suppliers

	err := tx.Table("suppliers").Order("order_completion_time ASC").Limit(1).Find(&supplier).Error
	if err != nil {
		return nil, err
	}

	return &supplier, nil
}

func UpdateSupplier(tx *gorm.DB, supplier *pkg.Suppliers) error {
	return tx.Save(&supplier).Error
}

func FindResourceAwareHandlingSuppliers(tx *gorm.DB) ([]pkg.Suppliers, error) {
	var suppliers []pkg.Suppliers

	err := tx.Table("suppliers").
		Where("order_handling_type = ?", true).
		Order("order_completion_time ASC").Find(&suppliers).Error

	if err != nil {
		return nil, err
	}

	return suppliers, nil
}

func UpdateSupplierStatus(tx *gorm.DB, orderIds *[]string, status string) (int64, error) {

	suppliers := tx.Table("suppliers").
		Where("current_order_id IN ?", *orderIds).
		Updates(map[string]interface{}{
			"status": status,
		})

	if suppliers.Error != nil {
		return 0, suppliers.Error
	}

	return suppliers.RowsAffected, nil
}

func FindSupplierByResourceId(tx *gorm.DB, resourceId int) (*pkg.Suppliers, error) {
	var supplier pkg.Suppliers

	err := tx.Table("suppliers").Where("id = ?", resourceId).Find(&supplier).Error

	if err != nil {
		return nil, err
	}

	return &supplier, nil
}
