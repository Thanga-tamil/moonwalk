package service

import (
	"testing"
	"time"
	"moonwalk/pkg"
)

func TestSchedulerFifoAssignment(t *testing.T) {
	dish := pkg.Dish{Id: 1, PreCooked: true}
	resources := []pkg.Resources{
		{Id: 1, Type: "CHEF", Status: IDLE},
		{Id: 2, Type: "CHEF", Status: IDLE},
		{Id: 3, Type: "SERVER", Status: IDLE},
		{Id: 4, Type: "SERVER", Status: IDLE},
	}

	order := scheduler(dish, &resources)

	if order.Alg != FIFO {
		t.Fatalf("expected FIFO algorithm, got %q", order.Alg)
	}
	if order.ResourceId != 3 {
		t.Fatalf("expected assignment to first idle server (id 3), got %d", order.ResourceId)
	}
	if order.DishId != 1 {
		t.Fatalf("expected dish id 1, got %d", order.DishId)
	}
	if order.OrderId == "" {
		t.Fatal("expected non-empty order id")
	}
}

func TestSchedulerFifoNoIdleServerQueues(t *testing.T) {
	dish := pkg.Dish{Id: 2, PreCooked: true}
	resources := []pkg.Resources{
		{Id: 1, Type: "CHEF", Status: IDLE},
		{Id: 2, Type: "CHEF", Status: BUSY},
		{Id: 3, Type: "SERVER", Status: BUSY},
		{Id: 4, Type: "SERVER", Status: BUSY},
	}

	order := scheduler(dish, &resources)

	if order.Alg != FIFO {
		t.Fatalf("expected FIFO algorithm, got %q", order.Alg)
	}
	if order.ResourceId != 0 {
		t.Fatalf("expected ResourceId 0 for queued order, got %d", order.ResourceId)
	}
	if order.Status != "PENDING" {
		t.Fatalf("expected PENDING status for queued order, got %q", order.Status)
	}
}

func TestSchedulerResourceAwareAssignment(t *testing.T) {
	dish := pkg.Dish{Id: 3, PreCooked: false, PrepTime: 20}
	resources := []pkg.Resources{
		{Id: 1, Type: "CHEF", Status: IDLE},
		{Id: 2, Type: "CHEF", Status: IDLE},
		{Id: 3, Type: "SERVER", Status: IDLE},
		{Id: 4, Type: "SERVER", Status: IDLE},
	}

	before := time.Now()
	order := scheduler(dish, &resources)
	after := time.Now()

	if order.Alg != RES_AWARE {
		t.Fatalf("expected RESOURCE AWARE algorithm, got %q", order.Alg)
	}
	if order.ResourceId != 1 {
		t.Fatalf("expected assignment to first idle chef (id 1), got %d", order.ResourceId)
	}
	expectedEta := order.CreatedAt.Add(20 * time.Minute)
	if order.Eta.Before(expectedEta.Add(-time.Minute)) || order.Eta.After(expectedEta.Add(time.Minute)) {
		t.Fatalf("expected ETA approx now+20min, got %v (before=%v after=%v)", order.Eta, before, after)
	}
}

func TestSchedulerResourceAwareNoIdleChefQueues(t *testing.T) {
	dish := pkg.Dish{Id: 4, PreCooked: false, PrepTime: 10}
	resources := []pkg.Resources{
		{Id: 1, Type: "CHEF", Status: BUSY},
		{Id: 2, Type: "CHEF", Status: BUSY},
		{Id: 3, Type: "SERVER", Status: IDLE},
		{Id: 4, Type: "SERVER", Status: IDLE},
	}

	order := scheduler(dish, &resources)

	if order.Alg != RES_AWARE {
		t.Fatalf("expected RESOURCE AWARE algorithm, got %q", order.Alg)
	}
	if order.ResourceId != 0 {
		t.Fatalf("expected ResourceId 0 for queued order, got %d", order.ResourceId)
	}
	if order.Status != "PENDING" {
		t.Fatalf("expected PENDING status for queued order, got %q", order.Status)
	}
}

func TestFifoScheduleEtaCalculation(t *testing.T) {
	dish := pkg.Dish{Id: 5, PreCooked: true}
	resources := []pkg.Resources{
		{Id: 3, Type: "SERVER", Status: IDLE},
	}

	order := fifoSchedule(dish, &resources)

	expectedEta := order.CreatedAt.Add(FIFO_ETA_MINUTES * time.Minute)
	if order.Eta.Before(expectedEta.Add(-time.Minute)) || order.Eta.After(expectedEta.Add(time.Minute)) {
		t.Fatalf("expected FIFO ETA ~now+5min, got %v", order.Eta)
	}
}

func TestBuildOrderFields(t *testing.T) {
	eta := time.Now().Add(10 * time.Minute)
	order := buildOrder(RES_AWARE, "order-123", 7, 9, eta)

	if order.OrderId != "order-123" {
		t.Fatalf("unexpected order id: %q", order.OrderId)
	}
	if order.ResourceId != 7 {
		t.Fatalf("unexpected resource id: %d", order.ResourceId)
	}
	if order.DishId != 9 {
		t.Fatalf("unexpected dish id: %d", order.DishId)
	}
	if order.Eta != eta {
		t.Fatalf("unexpected eta: %v", order.Eta)
	}
	if order.Alg != RES_AWARE {
		t.Fatalf("unexpected alg: %q", order.Alg)
	}
	if order.Status != "PENDING" {
		t.Fatalf("unexpected status: %q", order.Status)
	}
	if order.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}
