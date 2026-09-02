package service

import (
	"testing"
	"time"
	"moonwalk/pkg"
	"moonwalk/internal/app"

	log "github.com/Thanga-tamil/logger_lib"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite instance with the schema
// required by the repository layer and assigns it to the global app.DB.
func setupTestDB(t *testing.T) {
	t.Helper()

	// initialize the logger so scheduler/cron logging doesn't panic
	log.NewLogger("/tmp/moonwalk_test.log", "DEBUG")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	if err := db.Exec(`CREATE TABLE resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type varchar(199) NOT NULL DEFAULT '',
		current_order_id varchar(199) NOT NULL DEFAULT '',
		status VARCHAR(20) NOT NULL DEFAULT 'IDLE',
		order_status varchar(199) NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		name varchar(199) NOT NULL DEFAULT ''
	)`).Error; err != nil {
		t.Fatalf("failed to create resources table: %v", err)
	}

	if err := db.Exec(`CREATE TABLE dishes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		dish TEXT NOT NULL,
		price INTEGER,
		prep_time INT,
		is_available BOOLEAN NOT NULL DEFAULT FALSE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME,
		pre_cooked BOOLEAN NOT NULL DEFAULT FALSE
	)`).Error; err != nil {
		t.Fatalf("failed to create dishes table: %v", err)
	}

	if err := db.Exec(`CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id varchar(199) NOT NULL DEFAULT '',
		dish_id INT NOT NULL DEFAULT 0,
		chef varchar(199) NOT NULL DEFAULT '',
		status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
		eta DATETIME,
		alg varchar(199) NOT NULL DEFAULT 'FIFO',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		resource_id INT,
		UNIQUE(order_id)
	)`).Error; err != nil {
		t.Fatalf("failed to create orders table: %v", err)
	}

	if err := db.Exec(`CREATE TABLE order_executions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id varchar(199) NOT NULL DEFAULT '',
		status VARCHAR(20) NOT NULL DEFAULT '',
		algorithm varchar(199) NOT NULL DEFAULT '',
		time_estimated INTEGER NOT NULL DEFAULT 0,
		time_elapsed INTEGER NOT NULL DEFAULT 0,
		resource_id INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error; err != nil {
		t.Fatalf("failed to create order_executions table: %v", err)
	}

	app.DB = db
}

func seedResource(t *testing.T, id int, rType, status string) {
	t.Helper()
	if err := app.DB.Exec(
		"INSERT INTO resources (id, type, status, name) VALUES (?, ?, ?, ?)",
		id, rType, status, rType,
	).Error; err != nil {
		t.Fatalf("failed to seed resource: %v", err)
	}
}

func seedDish(t *testing.T, id int, prepTime int, preCooked bool) {
	t.Helper()
	if err := app.DB.Exec(
		"INSERT INTO dishes (id, dish, prep_time, pre_cooked) VALUES (?, ?, ?, ?)",
		id, "Dish", prepTime, preCooked,
	).Error; err != nil {
		t.Fatalf("failed to seed dish: %v", err)
	}
}

func seedOrder(t *testing.T, orderId string, dishId, resourceId int, status string, eta time.Time) {
	t.Helper()
	o := &pkg.Order{
		OrderId:    orderId,
		DishId:     dishId,
		ResourceId: resourceId,
		Status:     status,
		Eta:        eta,
		Alg:        FIFO,
		CreatedAt:  time.Now(),
	}
	if err := app.DB.Table("orders").Create(o).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}
}

func getResourceStatus(t *testing.T, id int) string {
	t.Helper()
	var status string
	if err := app.DB.Table("resources").Select("status").Where("id = ?", id).Scan(&status).Error; err != nil {
		t.Fatalf("failed to read resource status: %v", err)
	}
	return status
}

func getOrderStatus(t *testing.T, orderId string) string {
	t.Helper()
	var status string
	if err := app.DB.Table("orders").Select("status").Where("order_id = ?", orderId).Scan(&status).Error; err != nil {
		t.Fatalf("failed to read order status: %v", err)
	}
	return status
}

func TestProcessCompletedOrders(t *testing.T) {
	setupTestDB(t)
	seedResource(t, 1, "CHEF", BUSY)
	seedOrder(t, "o1", 1, 1, "PREPARING", time.Now().Add(-time.Minute))

	processCompletedOrders()

	if got := getOrderStatus(t, "o1"); got != "SERVED" {
		t.Fatalf("expected order SERVED after completing, got %q", got)
	}
	if got := getResourceStatus(t, 1); got != IDLE {
		t.Fatalf("expected resource back to IDLE after completing, got %q", got)
	}
}

func TestProcessCompletedOrdersNotPastETA(t *testing.T) {
	setupTestDB(t)
	seedResource(t, 1, "CHEF", BUSY)
	seedOrder(t, "o1", 1, 1, "PREPARING", time.Now().Add(30*time.Minute))

	processCompletedOrders()

	if got := getOrderStatus(t, "o1"); got != "PREPARING" {
		t.Fatalf("expected order still PREPARING (future ETA), got %q", got)
	}
	if got := getResourceStatus(t, 1); got != BUSY {
		t.Fatalf("expected resource still BUSY, got %q", got)
	}
}

func TestProcessPendingFifoOrder(t *testing.T) {
	setupTestDB(t)
	seedResource(t, 3, "SERVER", IDLE)
	seedDish(t, 2, 0, true)
	seedOrder(t, "o1", 2, 0, "PENDING", time.Now())

	processPendingOrders()

	if got := getOrderStatus(t, "o1"); got != "PREPARING" {
		t.Fatalf("expected pending FIFO order assigned to PREPARING, got %q", got)
	}
	if got := getResourceStatus(t, 3); got != BUSY {
		t.Fatalf("expected server marked BUSY, got %q", got)
	}
}

func TestProcessPendingResourceAwareOrder(t *testing.T) {
	setupTestDB(t)
	seedResource(t, 1, "CHEF", IDLE)
	seedDish(t, 3, 20, false)
	seedOrder(t, "o1", 3, 0, "PENDING", time.Now())

	processPendingOrders()

	if got := getOrderStatus(t, "o1"); got != "PREPARING" {
		t.Fatalf("expected pending RESOURCE AWARE order assigned to PREPARING, got %q", got)
	}
	if got := getResourceStatus(t, 1); got != BUSY {
		t.Fatalf("expected chef marked BUSY, got %q", got)
	}
}

func TestProcessPendingNoAvailableResource(t *testing.T) {
	setupTestDB(t)
	seedResource(t, 1, "CHEF", BUSY)
	seedResource(t, 3, "SERVER", BUSY)
	seedDish(t, 3, 20, false)
	seedOrder(t, "o1", 3, 0, "PENDING", time.Now())

	processPendingOrders()

	if got := getOrderStatus(t, "o1"); got != "PENDING" {
		t.Fatalf("expected order to remain PENDING (no idle resource), got %q", got)
	}
}

func TestProcessPendingMultipleOrdersUsesDistinctResources(t *testing.T) {
	setupTestDB(t)
	seedResource(t, 1, "CHEF", IDLE)
	seedResource(t, 2, "CHEF", IDLE)
	seedDish(t, 3, 20, false)
	seedOrder(t, "o1", 3, 0, "PENDING", time.Now())
	seedOrder(t, "o2", 3, 0, "PENDING", time.Now())

	processPendingOrders()

	// first order gets chef 1
	if got := getOrderStatus(t, "o1"); got != "PREPARING" {
		t.Fatalf("expected o1 assigned, got %q", got)
	}
	// second order must be assigned to the remaining idle chef (2)
	if got := getOrderStatus(t, "o2"); got != "PREPARING" {
		t.Fatalf("expected o2 assigned to remaining chef, got %q", got)
	}
	if got := getResourceStatus(t, 1); got != BUSY {
		t.Fatalf("expected chef 1 BUSY, got %q", got)
	}
	if got := getResourceStatus(t, 2); got != BUSY {
		t.Fatalf("expected chef 2 BUSY, got %q", got)
	}
}
