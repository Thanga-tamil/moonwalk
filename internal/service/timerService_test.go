package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"moonwalk/pkg"
	"moonwalk/internal/app"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetOrderTimerCountdown(t *testing.T) {
	setupTestDB(t)

	created := time.Now().Add(-2 * time.Minute)
	eta := time.Now().Add(8 * time.Minute) // 10 min total estimated
	o := &pkg.Order{
		OrderId:    "timer-1",
		DishId:     5,
		ResourceId: 3,
		Status:     "PREPARING",
		Eta:        eta,
		Alg:        "RESOURCE AWARE",
		CreatedAt:  created,
	}
	if err := app.DB.Table("orders").Create(o).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/order/timer-1", nil)

	GetOrderTimer(ginCtx, "timer-1")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object in response, got %#v", body["data"])
	}

	// ~10 min total estimated (~600s)
	est := data["time_estimated_seconds"].(float64)
	if est < 590 || est > 610 {
		t.Fatalf("expected ~600s estimated, got %v", est)
	}

	// ~2 min elapsed (~120s)
	el := data["time_elapsed_seconds"].(float64)
	if el < 110 || el > 130 {
		t.Fatalf("expected ~120s elapsed, got %v", el)
	}

	// ~8 min remaining (~480s)
	rem := data["time_remaining_seconds"].(float64)
	if rem < 470 || rem > 490 {
		t.Fatalf("expected ~480s remaining, got %v", rem)
	}

	if data["status"].(string) != "PREPARING" {
		t.Fatalf("unexpected status %v", data["status"])
	}
}

func TestGetOrderTimerNotFound(t *testing.T) {
	setupTestDB(t)

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/order/nonexistent", nil)

	GetOrderTimer(ginCtx, "nonexistent")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing order, got %d", w.Code)
	}
}
