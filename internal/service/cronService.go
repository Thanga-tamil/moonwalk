package service

import (
	"strings"
	"sync"
	"time"

	orderRepo "moonwalk/internal/repository"
	resourceRepo "moonwalk/internal/repository"

	log "github.com/Thanga-tamil/logger_lib"
)

const (
	CRON_INTERVAL = 5 * time.Second
	READY         = "READY"
)

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

			processResourceAwareOrders()
			processCompletedOrders()
			processPendingOrders()
			cronMu.Unlock()
		}
	}()
}

// make sure to add transaction for data consistency

func processResourceAwareOrders() {
	log.Info("scheduled call :: process resource aware orders")

	// each resource aware orders has to be served to customers by a supplier
	// ETA was calculated with the consideration of supplier time taken, so fetch and
	// update the resource aware orders by minusing supplier time taken from ETA and
	// update the status to 'READY'
	eta := time.Now().Add(time.Minute)

	orders, err := orderRepo.UpdateResourceAwareOrdersToReady(READY, eta)
	if err != nil {
		log.Error("Cron: error updating resource aware orders to READY:", err.Error())
	}

	// audit the transition to READY
	for _, o := range orders {
		o.Status = READY
		recordExecution(&o)
	}

	resources, err := resourceRepo.GetSuppliers()
	if err != nil {
		log.Error("Cron: error fetching suppliers:", err.Error())
		return
	}

	order, err := orderRepo.GetResourceAwareOrders(READY)

	if err != nil {
		log.Error("Cron: error fetching resource aware orders:", err.Error())
	} else {
		if strings.TrimSpace(order.OrderId) != "" {
			for _, resource := range *resources {
				if resource.Status == BUSY {
					continue
				}

				log.Debug("Cron: processing resource aware order: %s with resource: %s", order.OrderId, resource.Type)

				if err := resourceRepo.UpdateChefStatusToIdle(IDLE, order.ResourceId); err != nil {
					log.Error("Cron: error updating supplier status:", err.Error())
					continue
				}
				if err := resourceRepo.UpdateSupplierStatusToBusy(&resource, BUSY, order.OrderId); err != nil {
					log.Error("Cron: error updating supplier status:", err.Error())
					continue
				}
				if err := orderRepo.UpdateResourceAwareOrdersStatusToServing(&resource, order.OrderId); err != nil {
					log.Error("Cron: error serving resource aware orders:", err.Error())
					continue
				}
				order.ResourceId = resource.Id
				order.Status = "SERVING"

				// audit the transition to SERVING
				recordExecution(order)
			}
		}
	}

}

func processPendingOrders() {
	log.Info("scheduled call :: process pending orders")
	orders, err := orderRepo.GetPendingOrders()
	log.Infof("pending orders: %#v", &orders)
	if err != nil {
		log.Error("Cron: error fetching pending orders:", err.Error())
		return
	}
	if len(orders) == 0 {
		return
	}

	resources, err := resourceRepo.GetResources()
	log.Infof("available resources: %#v", resources)
	if err != nil {
		log.Error("Cron: error fetching resources:", err.Error())
		return
	}

	for _, o := range orders {
		log.Debug("Cron: processing pending order: %s", o.OrderId)
		dish, err := orderRepo.GetDish(o.DishId)
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
			if err := orderRepo.UpdateOrderStatusAndResourceId(o.OrderId, status, order.ResourceId); err != nil {
				log.Error("Cron: error updating order status:", err.Error())
				continue
			}
			if err := resourceRepo.UpdateResourceStatus(order.ResourceId, BUSY, o.OrderId); err != nil {
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
	log.Info("scheduled call :: process completed orders")
	orders, err := orderRepo.GetPreparingOrdersPastETA()
	if err != nil {
		log.Error("Cron: error fetching completed orders:", err.Error())
		return
	}

	for _, o := range orders {
		log.Infof("cron: completing order id: %s alg: %s ", o.OrderId, o.Alg)
		if err := orderRepo.UpdateOrderStatus(o.OrderId, "SERVED", time.Now()); err != nil {
			log.Error("Cron: error updating order status:", err.Error())
			continue
		}
		if o.ResourceId > 0 {
			// empty the resource's current order id since the order is now served
			currentOrderId := ""
			var resourceId int
			if o.Alg == RES_AWARE {
				resourceId = resourceRepo.FindResourceByOrderId(o.OrderId)
			} else {
				resourceId = o.ResourceId
			}
			if err := resourceRepo.UpdateResourceStatus(resourceId, IDLE, currentOrderId); err != nil {
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
