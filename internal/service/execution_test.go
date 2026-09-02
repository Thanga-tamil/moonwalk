package service

import (
	"testing"
	"time"
	"moonwalk/pkg"
	"moonwalk/internal/app"
)

func TestRecordExecutionPersistsAuditRow(t *testing.T) {
	setupTestDB(t)

	created := time.Now().Add(-3 * time.Minute)
	eta := created.Add(20 * time.Minute) // 20 min estimated

	o := &pkg.Order{
		OrderId:    "exec-1",
		DishId:     6,
		ResourceId: 1,
		Status:     "PREPARING",
		Eta:        eta,
		Alg:        "RESOURCE AWARE",
		CreatedAt:  created,
	}

	recordExecution(o)

	var count int64
	if err := app.DB.Table("order_executions").Where("order_id = ?", "exec-1").Count(&count).Error; err != nil {
		t.Fatalf("failed to count executions: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 execution row, got %d", count)
	}

	var e pkg.OrderExecution
	if err := app.DB.Table("order_executions").Where("order_id = ?", "exec-1").First(&e).Error; err != nil {
		t.Fatalf("failed to fetch execution: %v", err)
	}

	if e.OrderId != "exec-1" {
		t.Fatalf("unexpected order id %q", e.OrderId)
	}
	if e.Status != "PREPARING" {
		t.Fatalf("unexpected status %q", e.Status)
	}
	if e.Algorithm != "RESOURCE AWARE" {
		t.Fatalf("unexpected algorithm %q", e.Algorithm)
	}
	if e.ResourceId != 1 {
		t.Fatalf("unexpected resource id %d", e.ResourceId)
	}
	// estimated total is 20 min = 1200s
	if e.TimeEstimated != 1200 {
		t.Fatalf("expected 1200s estimated, got %d", e.TimeEstimated)
	}
	if e.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be set")
	}
}
