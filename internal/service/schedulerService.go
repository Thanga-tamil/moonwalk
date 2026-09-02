package service

import (
	"time"
	"moonwalk/pkg"
	"moonwalk/internal/utils"
	log "github.com/Thanga-tamil/logger_lib"
)

const (
	IDLE = "IDLE"
	RES_AWARE = "RESOURCE AWARE"
)

func scheduler(dish pkg.Dish, resources *[]pkg.Resources) pkg.Order {
	var order pkg.Order 

	_, servers := utils.Filter(*resources)

	log.Infox("chec")
	if dish.PreCooked {
		for _, s := range servers {
			if s.Status == IDLE {

				log.Infox("RES_AWARE, utils.GetRandomUUID(), s.Id, dish.Id, time.Now():", 
				RES_AWARE, utils.GetRandomUUID(), s.Id, dish.Id, time.Now())
				buildOrder(RES_AWARE, utils.GetRandomUUID(), s.Id, dish.Id, time.Now())
				break;
			}
		}
	}

	return order
}

func buildOrder(alg, orderId string, resId, dishId int, eta time.Time) pkg.Order {
	return pkg.Order{
		Eta: eta,
		ResourceId: resId,
		DishId: dishId,
		Alg: alg,
		Status: "PENDING",
		CreatedAt: time.Now(),
		OrderId: orderId,
	}
}

