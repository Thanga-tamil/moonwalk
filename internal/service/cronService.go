package service

import (
	"strings"
	"sync"
	"time"

	"moonwalk/internal/app"
	orderRepo "moonwalk/internal/repository"
	resourceRepo "moonwalk/internal/repository"
	"moonwalk/internal/utils"

	log "github.com/Thanga-tamil/logger_lib"
	"gorm.io/gorm"
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
	err := app.DB.Transaction(func(tx *gorm.DB) error {
		orders, err := orderRepo.UpdateResourceAwareOrdersToReady(READY, eta)
		if err != nil {
			log.Error("Cron: error updating resource aware orders to READY:", err.Error())
			return err
		}

		// audit the transition to READY
		for _, o := range orders {
			o.Status = READY
			recordExecution(tx, &o)
		}

		resources, err := resourceRepo.GetSuppliers()
		if err != nil {
			log.Error("Cron: error fetching suppliers:", err.Error())
			return err
		}

		order, err := orderRepo.GetResourceAwareOrders(READY)
		if err != nil {
			log.Error("Cron: error fetching resource aware orders:", err.Error())
			return err
		}

		if strings.TrimSpace(order.OrderId) != "" {
			for _, resource := range *resources {
				if resource.Status == BUSY {
					continue
				}
				log.Debugf("Cron: processing resource aware order: %s with resource: %s", order.OrderId, resource.Type)

				if err := resourceRepo.UpdateChefStatusToIdle(tx, IDLE, order.ResourceId); err != nil {
					log.Error("Cron: error updating supplier status:", err.Error())
					continue
				}
				if err := resourceRepo.UpdateSupplierStatusToBusy(tx, &resource, BUSY, order.OrderId); err != nil {
					log.Error("Cron: error updating supplier status:", err.Error())
					continue
				}
				if err := orderRepo.UpdateResourceAwareOrdersStatusToServing(tx, order.OrderId); err != nil {
					log.Error("Cron: error serving resource aware orders:", err.Error())
					continue
				}
				order.ResourceId = resource.Id
				order.Status = "SERVING"

				// audit the transition to SERVING
				recordExecution(tx, &order)
			}
		}

		return nil
	})
	if err != nil {
		log.Error("Cron: error processing resource aware orders:", err.Error())
	}
}

func processPendingOrders() {
	log.Info("scheduled call :: process pending orders")
	err := app.DB.Transaction(func(tx *gorm.DB) error {

		orders, err := orderRepo.GetPendingOrders()
		if err != nil {
			log.Error("Cron: error fetching pending orders:", err.Error())
			return err
		}
		if len(orders) == 0 {
			return err
		}

		utils.PrettyPrint("pending orders", orders)

		resources, err := resourceRepo.GetResources()
		if err != nil {
			log.Error("Cron: error fetching resources:", err.Error())
			return err
		}

		for _, o := range orders {
			log.Debug("Cron: processing pending order: ", o.OrderId)
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
				if err := orderRepo.UpdateOrderStatusAndResourceId(tx, o.OrderId, status, order.ResourceId); err != nil {
					log.Error("Cron: error updating order status:", err.Error())
					continue
				}
				if err := resourceRepo.UpdateResourceStatus(tx, order.ResourceId, BUSY, o.OrderId); err != nil {
					log.Error("Cron: error updating resource status:", err.Error())
					continue
				}
				// audit the transition to PREPARING
				o.Status = status
				o.ResourceId = order.ResourceId
				recordExecution(tx, &o)
			}
		}
		return nil
	})
	if err != nil {
		log.Error("Cron: error processing pending orders:", err.Error())
	}
}

func processCompletedOrders() {
	log.Info("scheduled call :: process completed orders")
	err := app.DB.Transaction(func(tx *gorm.DB) error {
		orders, err := orderRepo.GetPreparingOrdersPastETA(tx)
		if err != nil {
			log.Error("Cron: error fetching completed orders:", err.Error())
			return err
		}

		for _, o := range orders {
			log.Infof("cron: completing order id: %s alg: %s ", o.OrderId, o.Alg)
			if err := orderRepo.UpdateOrderStatus(tx, o.OrderId, "SERVED", time.Now()); err != nil {
				log.Error("Cron: error updating order status:", err.Error())
				continue
			}
			if o.ResourceId > 0 {
				// empty the resource's current order id since the order is now served
				currentOrderId := ""
				var resourceId int
				if o.Alg == RES_AWARE {
					resourceId = resourceRepo.FindResourceByOrderId(tx, o.OrderId)
				} else {
					resourceId = o.ResourceId
				}
				if err := resourceRepo.UpdateResourceStatus(tx, resourceId, IDLE, currentOrderId); err != nil {
					log.Error("Cron: error freeing resource:", err.Error())
				}
			} else {
				log.Warn("Cron: order ", o.OrderId, " has no assigned resource, cannot free resource")
			}

			// audit the transition to SERVED
			o.Status = "SERVED"
			recordExecution(tx, &o)
		}
		return nil
	})
	if err != nil {
		log.Error("Cron: error processing completed orders:", err.Error())
	}
}
