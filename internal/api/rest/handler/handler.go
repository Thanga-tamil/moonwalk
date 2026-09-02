package handler

import (
	"moonwalk/internal/service"
	"moonwalk/internal/utils"

	"github.com/gin-gonic/gin"
	log "github.com/Thanga-tamil/logger_lib"
)

// GetAvailableDishes function returns a list of available dishes
// by retrieving statistics from the db. Assume unavailability
// of dishes will be updated by the respective restaurants.
func GetAllDishes(ctx *gin.Context) {
	page, size, err := utils.Pagination(ctx)

	log.Debug("^GetAllDishes input param page:", page)
	log.Debug("^GetAllDishes input param size:", page)

	if err != nil {
		log.Error("Error while parsing integer from string:", err.Error())
		service.WriteErr(ctx, err.Error()); return
	}

	service.GetAllDishes(ctx, page, size)
}

func PlaceOrder(ctx *gin.Context) {
	data, err := service.ValidatePlaceOrderInput(ctx)

	if err != nil {
		log.Error("Error while parsing place order input:", err.Error())
		service.WriteErr(ctx, err.Error()); return
	}

	service.PlaceOrder(ctx, data)
}

func GetOrderTimer(ctx *gin.Context) {
	orderId := ctx.Param("id")

	if orderId == "" {
		service.WriteErr(ctx, "order id must not be empty")
		return
	}

	service.GetOrderTimer(ctx, orderId)
}
