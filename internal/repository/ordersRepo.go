package repository

import (
	"moonwalk/internal/app"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_lib"
)

func Save(o *pkg.Order) {
	if err := app.DB.Table("orders").Create(o).Error; err != nil {
		log.Error(err.Error())
	}
}

