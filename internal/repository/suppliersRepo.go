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
