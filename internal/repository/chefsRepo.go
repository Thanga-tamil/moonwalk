package repository

import (
	"moonwalk/pkg"
	"moonwalk/internal/app"
)

func GetChefs() (*[]pkg.Chefs, error) {
	var chef []pkg.Chefs

	if err := app.DB.Table("chefs").Find(&chef).Error; err != nil {
		return nil, err
	}

	return &chef, nil
}

