package service

import (
	"math"
	"sync"
	"time"

	"moonwalk/internal/app"
	ordersRepo "moonwalk/internal/repository"
	suppliersRepo "moonwalk/internal/repository"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_v2"
	"gorm.io/gorm"
)

var cronMu sync.Mutex

func StartCronService(cronInterval time.Duration) {
	log.Infox("Starting cron service with", "TimeInterval", cronInterval)

	ticker := time.NewTicker(cronInterval)
	go func() {
		defer ticker.Stop()

		for range ticker.C {
			if !cronMu.TryLock() {
				log.Debug("cron iteration skipped: previous iteration still running")
				continue
			}
			var wg sync.WaitGroup

			wg.Go(func() {
				handleResourceAwarePreparingOrders()
				handleResourceAwareReadyOrdersBySupplier()
			})

			wg.Wait()
			cronMu.Unlock()
		}
	}()
}

func handleResourceAwarePreparingOrders() {
	log.Info("handling resource aware preparing orders")
	err := app.DB.Transaction(func(tx *gorm.DB) error {

		alg := RES_AWARE
		currentStatus := PREPARING
		nextStatus := READY
		orders, err := ordersRepo.FindPreparingOrdersByEtaPlusOneMinute(tx, alg, currentStatus, nextStatus)
		if err != nil {
			return err
		}

		for _, o := range *orders {
			o.Status = READY
			o.UpdatedAt = time.Now()

			// audit order status 'READY' transition
			recordExecution(tx, &o)
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
			log.Info("No resource aware orders found to process 'READY' status")
			return nil
		}

		suppliers, err := suppliersRepo.FindResourceAwareHandlingSuppliers(tx)
		suppliersCount := len(*suppliers)
		if err != nil {
			return err
		}

		n := int(math.Min(float64(ordersCount), float64(suppliersCount)))

		orderSubList := make([]pkg.Order, 0, ordersCount)
		for i := 0; i < n; i++ {
			orderSubList = append(orderSubList, orders[i])
		}

		for _, o := range orderSubList {
			o.Status = SERVING
			o.UpdatedAt = time.Now()
			ordersRepo.UpdateOrder(tx, &o)

			// audit order status 'SERVING' transition
			recordExecution(tx, &o)
		}

		return nil
	})
	if err != nil {
		log.Errorx(err.Error())
	}
}
