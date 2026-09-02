package service

import (
	"sync"
	"time"
	"moonwalk/internal/repository"
	log "github.com/Thanga-tamil/logger_lib"
)

const CRON_INTERVAL = 10 * time.Second

var cronMu sync.Mutex

func StartCronService() {
	log.Infox("Starting cron service with interval:", CRON_INTERVAL)
	ticker := time.NewTicker(CRON_INTERVAL)
	go func() {
		for range ticker.C {
			if !cronMu.TryLock() {
				log.Debug("Cron iteration skipped: previous iteration still running")
				continue
			}
			processCompletedOrders()
			processPendingOrders()
			cronMu.Unlock()
		}
	}()
}

func processCompletedOrders() {
	orders, err := repository.GetPreparingOrdersPastETA()
	if err != nil {
		log.Error("Cron: error fetching completed orders:", err.Error())
		return
	}

	for _, o := range orders {
		log.Infox("Cron: completing order", o.OrderId)
		if err := repository.UpdateOrderStatus(o.OrderId, "SERVED"); err != nil {
			log.Error("Cron: error updating order status:", err.Error())
			continue
		}
		if o.ResourceId > 0 {
			if err := repository.UpdateResourceStatus(o.ResourceId, IDLE, ""); err != nil {
				log.Error("Cron: error freeing resource:", err.Error())
			}
		}

		// audit the transition to SERVED
		o.Status = "SERVED"
		recordExecution(&o)
	}
}

func processPendingOrders() {
	orders, err := repository.GetPendingOrders()
	if err != nil {
		log.Error("Cron: error fetching pending orders:", err.Error())
		return
	}
	if len(orders) == 0 {
		return
	}

	resources, err := repository.GetResources()
	if err != nil {
		log.Error("Cron: error fetching resources:", err.Error())
		return
	}

	for _, o := range orders {
		dish, err := repository.GetDish(o.DishId)
		if err != nil {
			log.Error("Cron: error fetching dish:", err.Error())
			continue
		}

		backlogMinutes, err := backlogFor(dish)
		if err != nil {
			log.Error("Cron: error computing backlog:", err.Error())
			continue
		}

		assigned := scheduler(dish, resources, backlogMinutes)
		if assigned.ResourceId > 0 {
			log.Infox("Cron: assigning pending order", o.OrderId, "to resource", assigned.ResourceId)
			if err := repository.UpdateOrderStatus(o.OrderId, "PREPARING"); err != nil {
				log.Error("Cron: error updating order status:", err.Error())
				continue
			}
			if err := repository.UpdateResourceStatus(assigned.ResourceId, BUSY, o.OrderId); err != nil {
				log.Error("Cron: error updating resource status:", err.Error())
				continue
			}
			for i, r := range *resources {
				if r.Id == assigned.ResourceId {
					(*resources)[i].Status = BUSY
					break
				}
			}

			// audit the transition to PREPARING
			o.Status = "PREPARING"
			o.ResourceId = assigned.ResourceId
			recordExecution(&o)
		}
	}
}
