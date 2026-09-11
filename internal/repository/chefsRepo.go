package repository

import (
	"moonwalk/pkg"

	"gorm.io/gorm"
)

func FindAvailableChef(tx *gorm.DB) (*pkg.Chefs, error) {
	var chef pkg.Chefs

	err := tx.Table("chefs").Order("cooking_completion_time ASC").Limit(1).Find(&chef).Error
	if err != nil {
		return nil, err
	}

	return &chef, nil
}

func UpdateChef(tx *gorm.DB, chef *pkg.Chefs) error {
	return tx.Save(&chef).Error
}
