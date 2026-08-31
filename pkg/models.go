package pkg

import "time"

type ServiceConfig struct {
	SqlDriverName 	 	string 		`json:"sqlDriverName"`
	SqlDataSourceName 	string 		`json:"sqlDataSourceName"`
	LogLevel			string		`json:"logLevel"`
	ServerMode			string		`json:"serverMode"`
}

type Dish struct {
	Id 			  int			`json:"id"`
	Dish          string 		`json:"dish"`
	IsAvailable   bool 			`json:"isAvailable"`
	Price 		  int			`json:"price"`
	AvailableUpto string	 	`json:"availableUpto"`
	CreatedAt     time.Time 	`json:"createdAt"`
	PrepTime      string 		`json:"prepTime"`
}

type PlaceOrderDto struct {
	OrderId 	string 	`json:"orderId"`
	DishId 		int		`json:"dishId"`
}

type Order struct {
	Id 			  int			`json:"id"`
	OrderId       string 		`json:"orderId"`
	DishId        int	 		`json:"dishId"`
	Price 		  int			`json:"price"`
	CreatedAt     time.Time 	`json:"createdAt"`
	PrepTime      string 		`json:"prepTime"`
}
