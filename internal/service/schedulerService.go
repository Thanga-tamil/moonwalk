package service

import (
	"time"
	"moonwalk/pkg"
	"moonwalk/internal/utils"
	log "github.com/Thanga-tamil/logger_lib"
)

const (
	IDLE        = "IDLE"
	BUSY        = "BUSY"
	FIFO        = "FIFO"
	RES_AWARE   = "RESOURCE AWARE"
	FIFO_ETA_MINUTES = 5
)

func scheduler(dish pkg.Dish, resources *[]pkg.Resources) pkg.Order {
	if dish.PreCooked {
		return fifoSchedule(dish, resources)
	}
	return resourceAwareSchedule(dish, resources)
}

func fifoSchedule(dish pkg.Dish, resources *[]pkg.Resources) pkg.Order {
	_, servers := utils.Filter(*resources)
	eta := time.Now().Add(time.Duration(FIFO_ETA_MINUTES) * time.Minute)

	for _, s := range servers {
		if s.Status == IDLE {
			log.Infox("FIFO schedule: assigning order to server", s.Id, dish.Id)
			return buildOrder(FIFO, utils.GetRandomUUID(), s.Id, dish.Id, eta)
		}
	}

	log.Infox("FIFO schedule: no idle server available, order queued")
	return buildOrder(FIFO, utils.GetRandomUUID(), 0, dish.Id, eta)
}

func resourceAwareSchedule(dish pkg.Dish, resources *[]pkg.Resources) pkg.Order {
	chefs, _ := utils.Filter(*resources)
	eta := time.Now().Add(time.Duration(dish.PrepTime) * time.Minute)

	for _, c := range chefs {
		if c.Status == IDLE {
			log.Infox("RESOURCE AWARE schedule: assigning order to chef", c.Id, dish.Id)
			return buildOrder(RES_AWARE, utils.GetRandomUUID(), c.Id, dish.Id, eta)
		}
	}

	log.Infox("RESOURCE AWARE schedule: no idle chef available, order queued")
	return buildOrder(RES_AWARE, utils.GetRandomUUID(), 0, dish.Id, eta)
}

func buildOrder(alg, orderId string, resId, dishId int, eta time.Time) pkg.Order {
	return pkg.Order{
		Eta:        eta,
		ResourceId: resId,
		DishId:     dishId,
		Alg:        alg,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
		OrderId:    orderId,
	}
}

