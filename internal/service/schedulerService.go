package service

import (
	"moonwalk/internal/repository"
	"moonwalk/internal/utils"
	"moonwalk/pkg"
	"time"

	"gorm.io/gorm"
)

// recordExecution persists an audit entry for a single order status transition
// every time an order changes state (PENDING -> PREPARING -> SERVED). This is
// the audit trail required by the problem statement.
func recordExecution(tx *gorm.DB, o *pkg.Order, resourceType string) {
	timeEstimated := int64(o.Eta.Sub(o.CreatedAt).Seconds())
	if timeEstimated < 0 {
		timeEstimated = 0
	}
	timeElapsed := int64(time.Since(o.CreatedAt).Seconds())
	if timeElapsed < 0 {
		timeElapsed = 0
	}

	repository.RecordExecution(tx, &pkg.OrderExec{
		OrderId:       o.OrderId,
		Status:        o.Status,
		Algorithm:     o.Alg,
		TimeEstimated: int(timeEstimated),
		TimeElapsed:   int(timeElapsed),
		ResourceId:    o.ResourceId,
		CreatedAt:     time.Now(),
		ResourceType:  resourceType,
	})
}

func recordExecutions(tx *gorm.DB, orders *[]pkg.Order, resourceType string) {

	var execs []pkg.OrderExec

	for _, o := range *orders {
		timeEstimated := int64(o.Eta.Sub(o.CreatedAt).Seconds())
		if timeEstimated < 0 {
			timeEstimated = 0
		}
		timeElapsed := int64(time.Since(o.CreatedAt).Seconds())
		if timeElapsed < 0 {
			timeElapsed = 0
		}
		audit := &pkg.OrderExec{
			OrderId:       o.OrderId,
			Status:        o.Status,
			Algorithm:     o.Alg,
			TimeEstimated: int(timeEstimated),
			TimeElapsed:   int(timeElapsed),
			ResourceId:    o.ResourceId,
			CreatedAt:     time.Now(),
			ResourceType:  resourceType,
		}
		execs = append(execs, *audit)

	}

	if len(execs) > 0 {
		repository.RecordExecutions(tx, &execs)
	}

}

const (
	FIFO_ETA_MINUTES = 1

	IDLE = "IDLE"
	BUSY = "BUSY"

	FIFO      = "FIFO"
	RES_AWARE = "RESOURCE AWARE"

	SUPPLIER = "SUPPLIER"
	CHEF     = "CHEF"

	PROCESSING = "PROCESSING"
	PREPARING  = "PREPARING"
	PENDING    = "PENDING"
	READY      = "READY"
	SERVING    = "SERVING"
	SERVED     = "SERVED"
)

// scheduler dispatches
//   - FIFO (pre-cooked) dishes are served by a server
//   - resource-aware dishes are cooked by a chef
func resourceAwareEtaScheduler(dish *pkg.Dish, chef *pkg.Chefs) pkg.Order {
	var eta time.Time
	if time.Now().After(chef.CookingCompletionTime) {
		eta = time.Now().Add(time.Duration(dish.PrepTime+FIFO_ETA_MINUTES) * time.Minute)
	} else {
		eta = chef.CookingCompletionTime.Add(time.Duration(dish.PrepTime+FIFO_ETA_MINUTES) * time.Minute)
	}

	order := buildOrder(RES_AWARE, utils.GetRandomUUID(), CHEF, chef.Id, dish.Id, eta)

	// track each CHEF's cooking completion time
	// so to find chef who would complete the cooking process
	// and be idle
	chef.UpdatedAt = time.Now()
	chef.CookingCompletionTime = eta

	return order
}

func buildOrder(alg, orderId, resourceType string, resourceId, dishId int, eta time.Time) pkg.Order {
	return pkg.Order{
		Eta:          eta,
		ResourceId:   resourceId,
		ResourceType: resourceType,
		DishId:       dishId,
		Alg:          alg,
		Status:       "PENDING",
		CreatedAt:    time.Now(),
		OrderId:      orderId,
	}
}

func fifoEtaScheduler(dish *pkg.Dish, supplier *pkg.Suppliers) pkg.Order {
	var eta time.Time
	if time.Now().After(supplier.OrderCompletionTime) {
		eta = time.Now().Add(time.Duration(FIFO_ETA_MINUTES) * time.Minute)
	} else {
		eta = supplier.OrderCompletionTime.Add(time.Duration(FIFO_ETA_MINUTES) * time.Minute)
	}

	order := buildOrder(FIFO, utils.GetRandomUUID(), SUPPLIER, supplier.Id, dish.Id, eta)

	supplier.UpdatedAt = time.Now()
	supplier.OrderCompletionTime = eta

	return order
}
