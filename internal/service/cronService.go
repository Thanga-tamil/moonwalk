package service

import (
	"moonwalk/internal/repository"
	"sync"
	"time"

	log "github.com/Thanga-tamil/logger_lib"
)

const CRON_INTERVAL = 5 * time.Second

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

			processResourceAwareOrders("READY")
			processCompletedOrders()
			processPendingOrders()
			cronMu.Unlock()
		}
	}()
}

func processResourceAwareOrders(status string) {
	if err := repository.UpdateResourceAwareOrdersToReady(status); err != nil {
		log.Error("Cron: error updating resource aware orders to READY:", err.Error())
	}

	resources, err := repository.GetSuppliers()
	if err != nil {
		log.Error("Cron: error fetching suppliers:", err.Error())
		return
	}

	for _, resource := range *resources {
		if resource.Status == "BUSY" {
			continue
		}

		log.Debug("Cron: processing resource aware with resource: ", resource.Id)
		orderIds, err := repository.GetResourceAwareOrders(status)
		if err != nil {
			log.Error("Cron: error fetching resource aware orders:", err.Error())
			continue
		}
		repository.UpdateSupplierStatus(&resource, "BUSY", orderIds)
		repository.ServeResourceAwareOrders(&resource, orderIds)

	}

}

func processPendingOrders() {
	log.Info("process pending orders")
	orders, err := repository.GetPendingOrders()
	log.Infof("pending orders: %#v", &orders)
	if err != nil {
		log.Error("Cron: error fetching pending orders:", err.Error())
		return
	}
	if len(orders) == 0 {
		return
	}

	resources, err := repository.GetResources()
	log.Infof("available resources: %#v", resources)
	if err != nil {
		log.Error("Cron: error fetching resources:", err.Error())
		return
	}

	for _, o := range orders {
		log.Debug("Cron: processing pending order: ", o.OrderId)
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

		order := scheduler(dish, resources, backlogMinutes)
		if order.ResourceId > 0 {
			var status string
			if order.Alg == FIFO {
				status = "PROCESSING"
			} else {
				status = "PREPARING"
			}
			log.Infox("Cron: assigning pending order", o.OrderId, "to resource", order.ResourceId)
			if err := repository.UpdateOrderStatusAndResourceId(o.OrderId, status, order.ResourceId); err != nil {
				log.Error("Cron: error updating order status:", err.Error())
				continue
			}
			if err := repository.UpdateResourceStatus(order.ResourceId, BUSY, o.OrderId); err != nil {
				log.Error("Cron: error updating resource status:", err.Error())
				continue
			}
			// audit the transition to PREPARING
			o.Status = status
			o.ResourceId = order.ResourceId
			recordExecution(&o)
		}
	}
}

func processCompletedOrders() {
	orders, err := repository.GetPreparingOrdersPastETA()
	if err != nil {
		log.Error("Cron: error fetching completed orders:", err.Error())
		return
	}

	for _, o := range orders {
		log.Infof("cron: completing order id: %s alg: %s ", o.OrderId, o.Alg)
		if err := repository.UpdateOrderStatus(o.OrderId, "SERVED", time.Now()); err != nil {
			log.Error("Cron: error updating order status:", err.Error())
			continue
		}
		if o.ResourceId > 0 {
			currentOrderId := "" // empty the resource's current order id since the order is now served
			if err := repository.UpdateResourceStatus(o.ResourceId, IDLE, currentOrderId); err != nil {
				log.Error("Cron: error freeing resource:", err.Error())
			}
		} else {
			log.Warn("Cron: order ", o.OrderId, " has no assigned resource, cannot free resource")
		}

		// audit the transition to SERVED
		o.Status = "SERVED"
		recordExecution(&o)
	}
}
