package service

import (
	"moonwalk/internal/repository"
	orderRepo "moonwalk/internal/repository"
	"moonwalk/internal/utils"
	"moonwalk/pkg"
	"time"

	log "github.com/Thanga-tamil/logger_v2"

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

	repository.RecordExecution(tx, &pkg.OrderExecution{
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
	IDLE             = "IDLE"
	BUSY             = "BUSY"
	FIFO             = "FIFO"
	RES_AWARE        = "RESOURCE AWARE"
	FIFO_ETA_MINUTES = 1
	SUPPLIER         = "SUPPLIER"
	CHEF             = "CHEF"
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
func strategyForDish(dish pkg.Dish, force string) bool {
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
func scheduler(dish pkg.Dish, resource *pkg.Resources, backlogMinutes int) pkg.Order {
	if strategyForDish(dish, schedulerStrategy) {
		return fifoSchedule(dish, resource, backlogMinutes)
	}
	return resourceAwareSchedule(dish, resource, backlogMinutes)
}

func fifoSchedule(dish pkg.Dish, supplier *pkg.Resources, backlogMinutes int) pkg.Order {
	// servers := utils.Filter(*resource, SUPPLIER)
	eta := time.Now().Add(time.Duration(backlogMinutes+FIFO_ETA_MINUTES) * time.Minute)

	if supplier.Status == IDLE {
		log.Infox("FIFO schedule: assigning order to supplier", supplier.Id, dish.Id)
	} else {
		log.Infox("", "FIFO schedule: no idle supplier available, order queued to supplier", supplier.Id)
	}

	return buildOrder(FIFO, utils.GetRandomUUID(), supplier.Id, dish.Id, eta)
}

func resourceAwareSchedule(dish pkg.Dish, chef *pkg.Resources, backlogMinutes int) pkg.Order {
	// Each RESOURCE AWARE order should be served by a server after compilation of cooking by a chef.
	// So, the ETA should be calculated as the sum of backlogMinutes + dish.PrepTime + FIFO_ETA_MINUTES

	var eta time.Time

	if chef.Status == IDLE {
		eta = time.Now().Add(time.Duration(backlogMinutes+dish.PrepTime+FIFO_ETA_MINUTES) * time.Minute)
		log.Infof("RESOURCE AWARE schedule: assigning order to chef: %d, dishId: %d", chef.Id, dish.Id)
	} else {
		inProgressOrderByChef, err := orderRepo.FindChefInProgressOrder(chef.Id)
		if err != nil {
			log.Error("Error while fetching preparing orders for new order asigning:", err.Error())
			panic(err) // todo
		}
		// eta = time.Now().Add(time.Duration(backlogMinutes+dish.PrepTime+FIFO_ETA_MINUTES) * time.Minute)
		eta = inProgressOrderByChef.Eta.Add(time.Duration(backlogMinutes+dish.PrepTime+FIFO_ETA_MINUTES) * time.Minute)
		log.Info("", "RESOURCE AWARE schedule: no idle chef available, order queued to chef", chef.Id)
	}
	return buildOrder(RES_AWARE, utils.GetRandomUUID(), chef.Id, dish.Id, eta)
}

func buildOrder(alg, orderId string, resourceId, dishId int, eta time.Time) pkg.Order {
	return pkg.Order{
		Eta:        eta,
		ResourceId: resourceId,
		DishId:     dishId,
		Alg:        alg,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
		OrderId:    orderId,
	}
}

// backlogFor returns the estimated minutes of queued work that must finish
// before the given dish can be processed. FIFO (pre-cooked / forced) orders fill
// the server queue (each occupying the fixed serving time), while resource-aware
// orders fill the chef queue (each occupying its dish's prep time).
func backlogFor(dish pkg.Dish) (int, error) {
	fifoCount, resourceMinutes, err := repository.GetPendingBacklog()
	if err != nil {
		return 0, err
	}

	if strategyForDish(dish, schedulerStrategy) {
		return fifoCount * FIFO_ETA_MINUTES, nil
	}
	return resourceMinutes, nil
}
