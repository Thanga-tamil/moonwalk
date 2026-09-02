package pkg

import "time"

type ServiceConfig struct {
	SqlDriverName 	 	string	`json:"sqlDriverName"`
	SqlDataSourceName 	string 	`json:"sqlDataSourceName"`
	LogLevel			string	`json:"logLevel"`
	ServerMode			string	`json:"serverMode"`
	// SchedulerStrategy selects which scheduling algorithm is used server-wide.
	// Supported values: "auto", "fifo", "resource_aware". Defaults to "auto"
	// which picks per dish (FIFO for pre-cooked, resource-aware otherwise).
	SchedulerStrategy string `json:"schedulerStrategy"`
	// Database pool tuning for long-running server operation.
	DbMaxOpenConns    int `json:"dbMaxOpenConns"`
	DbConnMaxLifetime int `json:"dbConnMaxLifetime"` // seconds
}

// Scheduler strategy identifiers
const (
	StrategyAuto          = "auto"
	StrategyFIFO          = "fifo"
	StrategyResourceAware = "resource_aware"
)

type Dish struct {
	Id          int       `json:"id"`
	Dish        string    `json:"dish"`
	Price       int       `json:"price"`
	PrepTime    int       `json:"prep_time"` // in minutes
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PreCooked	bool	  `json:"pre_cooked"`
}

type PlaceOrderDto struct {
	DishId 		int	 `json:"dishId"`
}

type Order struct {
	OrderId     string    `json:"order_id"`
	DishId      int       `json:"dish_id"`
	ResourceId 	int    	  `json:"resource_id"`
	Status      string    `json:"status"`
	Eta         time.Time `json:"eta"`
	Alg		    string    `json:"alg"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Resources struct {
	Id             int       `gorm:"column:id" json:"id"`
	Type           string    `gorm:"column:type" json:"type"`
	CurrentOrderID string    `gorm:"column:current_order_id" json:"current_order_id"`
	Status     	   string    `gorm:"column:status" json:"chef_status"`
	OrderStatus    string    `gorm:"column:order_status" json:"order_status"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
	Name           string    `gorm:"column:name" json:"name"`
}

// OrderExecution records a single status transition for an order. It forms the
// audit trail required by the problem statement: every execution step is
// persisted with its timestamp, the time elapsed so far and the estimated
// total time, the resulting order status and the algorithm in use.
type OrderExecution struct {
	Id             int       `gorm:"column:id" json:"id"`
	OrderId        string    `gorm:"column:order_id" json:"order_id"`
	Status         string    `gorm:"column:status" json:"status"`
	Algorithm      string    `gorm:"column:algorithm" json:"algorithm"`
	TimeEstimated  int       `gorm:"column:time_estimated" json:"time_estimated"` // seconds
	TimeElapsed    int       `gorm:"column:time_elapsed" json:"time_elapsed"`     // seconds
	ResourceId     int       `gorm:"column:resource_id" json:"resource_id"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
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
