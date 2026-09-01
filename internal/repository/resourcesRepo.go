package repository

import (
	"moonwalk/pkg"
	"moonwalk/internal/app"
)

func GetResources() (*[]pkg.Resources, error) {
	var chef []pkg.Resources

	if err := app.DB.Table("Resources").Find(&chef).Error; err != nil {
		return nil, err
	}

	return &chef, nil
}

