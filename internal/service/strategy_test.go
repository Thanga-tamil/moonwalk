package service

import (
	"testing"
	"moonwalk/pkg"
)

// strategyForDish is the pure decision function behind the configurable
// strategy; these tests pin down both the default and the forced behaviors.

func TestStrategyAutoUsesDishPreCooked(t *testing.T) {
	fifoDish := pkg.Dish{PreCooked: true}
	awareDish := pkg.Dish{PreCooked: false}

	if !strategyForDish(fifoDish, pkg.StrategyAuto) {
		t.Fatal("expected auto strategy to route pre-cooked dish to FIFO")
	}
	if strategyForDish(awareDish, pkg.StrategyAuto) {
		t.Fatal("expected auto strategy to route non pre-cooked dish to resource-aware")
	}
}

func TestStrategyForcedFIFO(t *testing.T) {
	dish := pkg.Dish{PreCooked: false}
	if !strategyForDish(dish, pkg.StrategyFIFO) {
		t.Fatal("expected forced FIFO strategy to route even a cooked dish to FIFO")
	}
}

func TestStrategyForcedResourceAware(t *testing.T) {
	dish := pkg.Dish{PreCooked: true}
	if strategyForDish(dish, pkg.StrategyResourceAware) {
		t.Fatal("expected forced resource-aware strategy to route even a pre-cooked dish to chefs")
	}
}

func TestSetSchedulerStrategyDefaultsToAuto(t *testing.T) {
	SetSchedulerStrategy("")
	if schedulerStrategy != pkg.StrategyAuto {
		t.Fatalf("expected default strategy auto, got %q", schedulerStrategy)
	}
	SetSchedulerStrategy(pkg.StrategyResourceAware)
	if schedulerStrategy != pkg.StrategyResourceAware {
		t.Fatalf("expected strategy resource_aware, got %q", schedulerStrategy)
	}
	// restore default for the rest of the test run
	SetSchedulerStrategy(pkg.StrategyAuto)
}
