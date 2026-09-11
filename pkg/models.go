package pkg

import "time"

type ServiceConfig struct {
	ServerHost             string `json:"serverHost"`
	ServerPort             int    `json:"serverPort"`
	SqlDriverName          string `json:"sqlDriverName"`
	SqlDataSourceName      string `json:"sqlDataSourceName"`
	LogLevel               int    `json:"logLevel"` // 0 = ERROR, 1 = INFO, 2 = DEBUG, 3 = WARNING
	LogFile                string `json:"logFile"`
	LogWriteToFile         bool   `json:"logWriteToFile"`
	ServerMode             string `json:"serverMode"`
	CronInterval           int    `json:"cronInterval"` // seconds
	DbMaxIdleConns         int    `json:"dbMaxIdleConns"`
	DbMaxOpenConns         int    `json:"dbMaxOpenConns"`
	DbConnMaxLifetime      int    `json:"dbConnMaxLifetime"` // seconds
	PendingOrdersBatchSize int    `json:"batchSize"`
}

type Dish struct {
	Id          int       `json:"id"`
	Dish        string    `json:"dish"`
	Price       int       `json:"price"`
	PrepTime    int       `json:"prep_time"` // in minutes
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PreCooked   bool      `json:"pre_cooked"`
}

type PlaceOrderDto struct {
	DishId int `json:"dishId"`
}

type Order struct {
	OrderId      string    `json:"order_id"`
	DishId       int       `json:"dish_id"`
	ResourceId   int       `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	Status       string    `json:"status"`
	Eta          time.Time `json:"eta"`
	Alg          string    `json:"alg"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Suppliers struct {
	Id                  int       `gorm:"column:id" json:"id"`
	CurrentOrderID      string    `gorm:"column:current_order_id" json:"current_order_id"`
	Status              string    `gorm:"column:status" json:"chef_status"`
	OrderCompletionTime time.Time `gorm:"column:order_completion_time" json:"order_completion_time"`
	OrderHandlingType   bool      `json:"order_handling_type"`
	CreatedAt           time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type Chefs struct {
	Id                    int       `gorm:"column:id" json:"id"`
	CurrentOrderID        string    `gorm:"column:current_order_id" json:"current_order_id"`
	Status                string    `gorm:"column:status" json:"chef_status"`
	CookingCompletionTime time.Time `gorm:"column:cooking_completion_time" json:"cooking_completion_time"`
	CreatedAt             time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// OrderExecution records a single status transition for an order. It forms the
// audit trail required by the problem statement: every execution step is
// persisted with its timestamp, the time elapsed so far and the estimated
// total time, the resulting order status and the algorithm in use.
type OrderExec struct {
	Id            int       `gorm:"column:id" json:"id"`
	OrderId       string    `gorm:"column:order_id" json:"order_id"`
	Status        string    `gorm:"column:status" json:"status"`
	Algorithm     string    `gorm:"column:algorithm" json:"algorithm"`
	TimeEstimated int       `gorm:"column:time_estimated" json:"time_estimated"` // seconds
	TimeElapsed   int       `gorm:"column:time_elapsed" json:"time_elapsed"`     // seconds
	ResourceId    int       `gorm:"column:resource_id" json:"resource_id"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

// TimerDto is the payload returned by the countdown timer endpoint. It exposes
// the order's remaining time so a customer can render a live countdown.
type TimerDto struct {
	OrderId       string    `json:"order_id"`
	DishId        int       `json:"dish_id"`
	Status        string    `json:"status"`
	Algorithm     string    `json:"algorithm"`
	Eta           time.Time `json:"eta"`
	TimeEstimated int64     `json:"time_estimated_seconds"` // total estimated duration
	TimeElapsed   int64     `json:"time_elapsed_seconds"`   // elapsed since creation
	TimeRemaining int64     `json:"time_remaining_seconds"` // countdown value (>=0)
}
