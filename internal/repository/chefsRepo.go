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

func FindChefByResourceId(tx *gorm.DB, resourceId int) (*pkg.Chefs, error) {
	var chef pkg.Chefs

	err := tx.Table("chefs").Where("id = ?", resourceId).Find(&chef).Error
	if err != nil {
		return nil, err
	}

	return &chef, nil
}

func UpdateChef(tx *gorm.DB, chef *pkg.Chefs) error {
	return tx.Save(&chef).Error
}

func UpdateChefStatus(tx *gorm.DB, status string, orderIds []string) (int64, error) {

	chefs := tx.Table("chefs").Where("current_order_id IN ?", orderIds).
		Updates(map[string]interface{}{
			"status": status,
		})

	if chefs.Error != nil {
		return 0, chefs.Error
	}

	return chefs.RowsAffected, nil
}
