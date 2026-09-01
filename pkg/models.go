package pkg

import "time"

type ServiceConfig struct {
	SqlDriverName 	 	string	`json:"sqlDriverName"`
	SqlDataSourceName 	string 	`json:"sqlDataSourceName"`
	LogLevel			string	`json:"logLevel"`
	ServerMode			string	`json:"serverMode"`
}

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
	Chef        string    `json:"chef"`
	Status      string    `json:"status"`
	Eta         time.Time `json:"eta"`
	Algorithm   string    `json:"alg"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Resources struct {
	Type           string    `gorm:"column:chef" json:"type"`
	CurrentOrderID string    `gorm:"column:current_order_id" json:"current_order_id"`
	ChefStatus     string    `gorm:"column:chef_status" json:"chef_status"`
	OrderStatus    string    `gorm:"column:order_status" json:"order_status"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}
