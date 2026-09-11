package service

import (
	"math"
	"sync"
	"time"

	"moonwalk/internal/app"
	chefsRepo "moonwalk/internal/repository"
	ordersRepo "moonwalk/internal/repository"
	suppliersRepo "moonwalk/internal/repository"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_v2"
	"gorm.io/gorm"
)

var cronMu sync.Mutex

func StartCronService(cronInterval time.Duration, pendingOrdersBatchSize int) {
	log.Infox("Starting cron service with", "TimeInterval", cronInterval)

	ticker := time.NewTicker(cronInterval)
	go func() {
		defer ticker.Stop()

		for range ticker.C {
			if !cronMu.TryLock() {
				log.Debug("cron iteration skipped: previous iteration still running")
				continue
			}
			func() {
				defer cronMu.Unlock()

				var wg sync.WaitGroup

				wg.Go(handleResourceAwarePreparingOrders)
				wg.Go(handleResourceAwareReadyOrdersBySupplier)
				wg.Go(handleResourceAwareServingOrders)
				wg.Go(func() {
					handlePendingOrders(pendingOrdersBatchSize)
				})
				wg.Go(handlFifoProcessingOrders)

				wg.Wait()
			}()
		}
	}()
}

func handleResourceAwarePreparingOrders() {
	log.Info("handling resource aware preparing orders")
	err := app.DB.Transaction(func(tx *gorm.DB) error {

		alg := RES_AWARE
		currentStatus := PREPARING
		nextStatus := READY
		eta := time.Now().Add(time.Minute)
		orders, orderIds, err := ordersRepo.UpdateResourceAwareOrdersStatusByETA(tx, alg, currentStatus, nextStatus, eta)
		if err != nil {
			return err
		}

		status := IDLE
		updatedChefCount, err := chefsRepo.UpdateChefStatus(tx, status, *orderIds)
		if err != nil {
			return err
		}
		log.Debugf("%d chef status updated to 'IDLE'", updatedChefCount)

		// audit order status 'READY' transition
		resourceType := CHEF
		for _, o := range *orders {
			recordExecution(tx, &o, resourceType)
		}

		return nil
	})

	if err != nil {
		log.Errorx(err.Error())
	}
}

func handleResourceAwareReadyOrdersBySupplier() {
	log.Info("handling resource aware ready orders")
	err := app.DB.Transaction(func(tx *gorm.DB) error {

		status := READY
		orders, err := ordersRepo.FetchResourceAwareOrdersByStatus(tx, status)
		ordersCount := len(orders)
		if err != nil {
			return err
		}
		if ordersCount == 0 {
			log.Debug("No resource aware 'preparing' order found to process 'READY' state")
			return nil
		}

		suppliers, err := suppliersRepo.FindResourceAwareHandlingSuppliers(tx)
		suppliersCount := len(suppliers)
		if err != nil {
			return err
		}
		log.Debugx("suppliers: ", suppliers)

		n := int(math.Min(float64(ordersCount), float64(suppliersCount)))

		orderSubList := make([]pkg.Order, 0, ordersCount)
		for i := 0; i < n; i++ {
			orderSubList = append(orderSubList, orders[i])
		}

		for i, o := range orderSubList {
			supplier := suppliers[i]
			if supplier.Status == IDLE {
				o.Status = SERVING
				o.UpdatedAt = time.Now()

				ordersRepo.UpdateOrder(tx, &o)

				// audit order status 'SERVING' transition
				resourceType := SUPPLIER
				o.ResourceId = supplier.Id
				recordExecution(tx, &o, resourceType)

				supplier.Status = BUSY
				supplier.UpdatedAt = time.Now()
				supplier.CurrentOrderID = o.OrderId
				if err := suppliersRepo.UpdateSupplier(tx, &supplier); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Errorx(err.Error())
	}
}

func handleResourceAwareServingOrders() {
	log.Info("handling resource aware serving orders")
	err := app.DB.Transaction(func(tx *gorm.DB) error {

		alg := RES_AWARE
		currentStatus := SERVING
		nextStatus := SERVED
		eta := time.Now()

		orders, orderIds, err := ordersRepo.UpdateResourceAwareOrdersStatusByETA(tx, alg, currentStatus, nextStatus, eta)
		if err != nil {
			return err
		} else if len(*orders) == 0 {
			return nil
		}
		status := IDLE
		chefsRepo.UpdateSupplierStatus(tx, orderIds, status)

		resourceType := CHEF
		// audit order status 'SERVED' transition
		for _, o := range *orders {
			recordExecution(tx, &o, resourceType)
		}

		return nil
	})

	if err != nil {
		log.Errorx(err.Error())
	}
}

func handlePendingOrders(pendingOrdersBatchSize int) {
	log.Info("handling resource aware pending orders")
	err := app.DB.Transaction(func(tx *gorm.DB) error {

		status := PENDING
		orders, err := ordersRepo.FetchPendingResourceAwareOrders(tx, status, pendingOrdersBatchSize)
		ordersCount := len(orders)
		if err != nil {
			return err
		} else if ordersCount == 0 {
			log.Info("No pending state order found")
			return nil
		}

		for _, o := range orders {
			if o.ResourceType == SUPPLIER {
				o.Status = PROCESSING
				supplier, err := suppliersRepo.FindSupplierByResourceId(tx, o.ResourceId)
				if err != nil {
					return err
				}
				if supplier.Status == IDLE {
					o.UpdatedAt = time.Now()
					if err = ordersRepo.UpdateOrder(tx, &o); err != nil {
						return err
					}

					// audit order status 'PROCESSING' transition
					resourceType := SUPPLIER
					recordExecution(tx, &o, resourceType)

					supplier.Status = BUSY
					supplier.UpdatedAt = time.Now()
					supplier.CurrentOrderID = o.OrderId
					if err = suppliersRepo.UpdateSupplier(tx, supplier); err != nil {
						return err
					}
				}
			} else {
				o.Status = PREPARING
				chef, err := chefsRepo.FindChefByResourceId(tx, o.ResourceId)
				if err != nil {
					return err
				}
				if chef.Status == IDLE {
					o.UpdatedAt = time.Now()
					if err = ordersRepo.UpdateOrder(tx, &o); err != nil {
						return err
					}

					// audit order status 'PREPARING' transition
					resourceType := CHEF
					recordExecution(tx, &o, resourceType)

					chef.Status = BUSY
					chef.UpdatedAt = time.Now()
					chef.CurrentOrderID = o.OrderId
					if err = chefsRepo.UpdateChef(tx, chef); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Errorx(err.Error())
	}
}

func handlFifoProcessingOrders() {
	log.Info("handling fifo processing orders")
	err := app.DB.Transaction(func(tx *gorm.DB) error {

		alg := FIFO
		eta := time.Now()
		currentStatus := PROCESSING
		nextStatus := SERVED

		orders, orderIds, err := ordersRepo.UpdateFifoProcessingOrders(tx, alg, currentStatus, nextStatus, eta)
		if err != nil {
			return err
		}

		for _, o := range *orders {
			o.Status = nextStatus
			o.UpdatedAt = time.Now()
			resourceType := SUPPLIER
			recordExecution(tx, &o, resourceType)
		}

		status := IDLE
		updatedSuppliers, err := suppliersRepo.UpdateSupplierStatus(tx, orderIds, status)
		if err != nil {
			return err
		}
		log.Debugf("%d supplier status updated to 'IDLE'", updatedSuppliers)

		return nil
	})

	if err != nil {
		log.Errorx(err.Error())
	}
}
