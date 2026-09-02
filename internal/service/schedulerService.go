package service

import (
	"time"
	"moonwalk/pkg"
	"moonwalk/internal/utils"
	"moonwalk/internal/repository"
	log "github.com/Thanga-tamil/logger_lib"
)

// recordExecution persists an audit entry for a single order status transition
// every time an order changes state (PENDING -> PREPARING -> SERVED). This is
// the audit trail required by the problem statement.
func recordExecution(o *pkg.Order) {
	timeEstimated := int64(o.Eta.Sub(o.CreatedAt).Seconds())
	if timeEstimated < 0 {
		timeEstimated = 0
	}
	timeElapsed := int64(time.Since(o.CreatedAt).Seconds())
	if timeElapsed < 0 {
		timeElapsed = 0
	}

	repository.RecordExecution(&pkg.OrderExecution{
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
	IDLE        = "IDLE"
	BUSY        = "BUSY"
	FIFO        = "FIFO"
	RES_AWARE   = "RESOURCE AWARE"
	FIFO_ETA_MINUTES = 5
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
func scheduler(dish pkg.Dish, resources *[]pkg.Resources, backlogMinutes int) pkg.Order {
	if strategyForDish(dish, schedulerStrategy) {
		return fifoSchedule(dish, resources, backlogMinutes)
	}
	return resourceAwareSchedule(dish, resources, backlogMinutes)
}

func fifoSchedule(dish pkg.Dish, resources *[]pkg.Resources, backlogMinutes int) pkg.Order {
	_, servers := utils.Filter(*resources)
	eta := time.Now().Add(time.Duration(backlogMinutes+FIFO_ETA_MINUTES) * time.Minute)

	for _, s := range servers {
		if s.Status == IDLE {
			log.Infox("FIFO schedule: assigning order to server", s.Id, dish.Id)
			return buildOrder(FIFO, utils.GetRandomUUID(), s.Id, dish.Id, eta)
		}
	}

	log.Infox("FIFO schedule: no idle server available, order queued")
	return buildOrder(FIFO, utils.GetRandomUUID(), 0, dish.Id, eta)
}

func resourceAwareSchedule(dish pkg.Dish, resources *[]pkg.Resources, backlogMinutes int) pkg.Order {
	chefs, _ := utils.Filter(*resources)
	eta := time.Now().Add(time.Duration(backlogMinutes+dish.PrepTime) * time.Minute)

	for _, c := range chefs {
		if c.Status == IDLE {
			log.Infox("RESOURCE AWARE schedule: assigning order to chef", c.Id, dish.Id)
			return buildOrder(RES_AWARE, utils.GetRandomUUID(), c.Id, dish.Id, eta)
		}
	}

	log.Infox("RESOURCE AWARE schedule: no idle chef available, order queued")
	return buildOrder(RES_AWARE, utils.GetRandomUUID(), 0, dish.Id, eta)
}

func buildOrder(alg, orderId string, resId, dishId int, eta time.Time) pkg.Order {
	return pkg.Order{
		Eta:        eta,
		ResourceId: resId,
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

