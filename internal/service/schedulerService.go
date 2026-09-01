package service

import (
	"moonwalk/internal/utils"
	"moonwalk/pkg"
	"time"

	log "github.com/Thanga-tamil/logger_lib"
)

func scheduler(dish pkg.Dish, resources *[]pkg.Resources) pkg.Order {
	var order pkg.Order 

	filtered := utils.Filter(resources, func(r pkg.Resources) bool {
		return r.Type == "something"
	})

	if dish.PreCooked {
		for _, res := range *resources {
			log.Info("hello cook:", res)
			if res.ChefStatus == "IDLE" {
				// buildOrder()
			}
		}
	}

	return order
}

func buildOrder(alg, chef string, dishId int, eta time.Time) pkg.Order {
	return pkg.Order{
		Eta: eta,
		Chef: chef,
		DishId: dishId,
		Algorithm: alg,
		Status: "PENDING",
		CreatedAt: time.Now(),
		OrderId: utils.GetRandomUUID(),
	}
}
