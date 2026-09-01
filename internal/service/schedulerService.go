package service

import (
	"moonwalk/pkg"
	log "github.com/Thanga-tamil/logger_lib"
)

func scheduler(dish *pkg.Dish, chefs *[]pkg.Chefs) {

	for chef := range *chefs {
		log.Info("hello cook:", chef)
	}
}

