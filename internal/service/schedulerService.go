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
func recordExecution(tx *gorm.DB, o *pkg.Order) {
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
	})
}

const (
	FIFO_ETA_MINUTES = 1

	IDLE = "IDLE"
	BUSY = "BUSY"

	FIFO      = "FIFO"
	RES_AWARE = "RESOURCE AWARE"

	SUPPLIER = "SUPPLIER"
	CHEF     = "CHEF"

	PREPARING = "PREPARING"
	READY     = "READY"
	SERVING   = "SERVING"
	SERVED    = "SERVED"
)

// schedulerStrategy holds the server-wide strategy loaded from config. It is
// written once at startup to keep the scheduler safe under concurrent access.
var schedulerStrategy = pkg.StrategyAuto

// SetSchedulerStrategy configures the server-wide scheduling strategy from the
// application config. Called once during startup.
func SetSchedulerStrategy(strategy string) {
	if strategy == "" {
		strategy = pkg.StrategyAuto
	}
	schedulerStrategy = strategy
}

// strategyDish selects which strategy a dish uses given the optional server-wide
// strategy override. A "forced" strategy applies to every dish; "auto" defers
// to the dish's own pre-cooked flag.
func strategyForDish(dish *pkg.Dish, force string) bool {
	// returns true when FIFO (server) should be used
	switch force {
	case pkg.StrategyFIFO:
		return true
	case pkg.StrategyResourceAware:
		return false
	default:
		return dish.PreCooked
	}
}

// scheduler dispatches to the appropriate strategy:
//   - FIFO (pre-cooked / forced) dishes are served by a server
//   - resource-aware dishes are cooked by a chef
//
// The forced strategy comes from config so the same build can serve multiple
// restaurants with different performance strategies (multi-tenant adaptability).
//
// backlogMinutes is the estimated minutes of work already queued ahead of this
// order (from GetPendingBacklog). It is added to the ETA so the countdown
// reflects the current kitchen backlog, not just an empty kitchen.
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
