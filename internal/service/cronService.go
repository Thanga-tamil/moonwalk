package service

import (
	"math"
	"sync"
	"time"

	"moonwalk/internal/app"
	orderRepo "moonwalk/internal/repository"
	resourceRepo "moonwalk/internal/repository"
	"moonwalk/internal/utils"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_lib"
	"gorm.io/gorm"
)

const (
	READY = "READY"
)

var cronMu sync.Mutex

func StartCronService(cronInterval time.Duration) {
	log.Infox("Starting cron service with interval: ", cronInterval)

	ticker := time.NewTicker(cronInterval)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			if !cronMu.TryLock() {
				log.Debug("Cron iteration skipped: previous iteration still running")
				continue
			}
			var wg sync.WaitGroup

			wg.Go(func() {
				processResourceAwareOrders()
			})
			wg.Go(func() {
				processCompletedOrders()
			})
			wg.Go(func() {
				processPendingOrders()
			})

			wg.Wait()
			cronMu.Unlock()
		}
	}()
}

// make sure to use transaction for data consistency

func processResourceAwareOrders() {
	log.Info("scheduled call :: process resource aware orders")

	// each resource aware orders has to be served to customers by a supplier
	// ETA was calculated with the consideration of supplier time taken, so fetch and
	// update the resource aware orders by minusing supplier time taken from ETA and
	// update the status to 'READY'
	eta := time.Now().Add(time.Minute)
	err := app.DB.Transaction(func(tx *gorm.DB) error {

		updatedOrders, err := orderRepo.UpdateResourceAwareOrdersToReady(tx, READY, eta)
		log.Infof("Cron: %d resource aware orders updated to READY", len(updatedOrders))
		if err != nil {
			log.Error("Cron: error updating resource aware orders to READY:", err.Error())
			return err
		}
		// audit the transition to READY
		for _, o := range updatedOrders {
			o.Status = READY
			recordExecution(tx, &o)
		}

		orders, err := orderRepo.FetchOrdersByStatusReady(tx, READY, eta)
		if err != nil {
			log.Error("Cron: error fetching resource aware orders:", err.Error())
			return err
		} else if len(orders) == 0 {
			log.Info("Cron: no resource aware orders to process")
			return nil
		}

		utils.SortOrdersByCreatedAt(&orders)

		suppliers, err := resourceRepo.GetSuppliers()
		supplierCount := len(suppliers)
		if err != nil {
			log.Error("Cron: error fetching suppliers:", err.Error())
			return err
		} else if supplierCount == 0 {
			log.Debug("Cron: no suppliers available to process orders")
			return nil
		}

		n := int(math.Min(float64(len(orders)), float64(supplierCount)))

		orderSubList := make([]pkg.Order, 0, len(orders))
		for i := 0; i < n; i++ {
			orderSubList = append(orderSubList, orders[i])
		}

		for i, order := range orderSubList {
			supply := suppliers[i]
			if supply.Status == IDLE {
				log.Debugf("Cron: processing resource aware order: %s with resource: %s", order.OrderId, supply.Type)

				if err := resourceRepo.UpdateChefStatusToIdle(tx, IDLE, order.ResourceId); err != nil {
					log.Error("Cron: error updating supplier status:", err.Error())
					return err
				}
				if err := resourceRepo.UpdateSupplierStatusToBusy(tx, &supply, BUSY, order.OrderId); err != nil {
					log.Error("Cron: error updating supplier status:", err.Error())
					return err
				}
				if err := orderRepo.UpdateResourceAwareOrdersStatusToServing(tx, order.OrderId); err != nil {
					log.Error("Cron: error serving resource aware orders:", err.Error())
					return err
				}
				order.ResourceId = supply.Id
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
				currentOrderId := ""

				var resourceId int
				if o.Alg == RES_AWARE {
					resourceId = resourceRepo.FindResourceByOrderId(tx, o.OrderId)
				} else {
					resourceId = o.ResourceId
				}

				err := resourceRepo.UpdateResourceStatus(tx, resourceId, IDLE, currentOrderId)
				if err != nil {
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
