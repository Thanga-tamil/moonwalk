package service

import (
	"time"
	"moonwalk/pkg"
	"moonwalk/internal/utils"
	log "github.com/Thanga-tamil/logger_lib"
)

// Order struct {
// 	OrderId     string    `json:"orderId"`
// 	DishId      int       `json:"dishId"`
// 	Chef        string    `json:"chef"`
// 	Status      string    `json:"status"`
// 	Eta         time.Time `json:"eta"`
// 	Algorithm   string    `json:"alg"`
// 	CreatedAt   time.Time `json:"createdAt"`
// 	UpdatedAt   time.Time `json:"updatedAt"`
// }
func scheduler(dish pkg.Dish, chefs *[]pkg.Chefs) pkg.Order {

	var order pkg.Order 

	for _, chef := range *chefs {
		log.Info("hello cook:", chef)
		if chef.ChefStatus == "IDLE" {
			order = pkg.Order{	
								OrderId: utils.GetRandomUUID(),
								DishId: dish.Id,
								Chef: chef.Chef,
								Status: "PENDING",
								// Eta: nil,
								// Algorithm: ,
								CreatedAt: time.Now(),
								}
		}
	}

	return order
}

