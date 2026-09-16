package handler

import (
	dishService "moonwalk/internal/service"
	orderService "moonwalk/internal/service"
	timerService "moonwalk/internal/service"
	"moonwalk/internal/utils"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"
)

func AddDish(ctx *gin.Context) {
	var dish *pkg.AddDishDto

	if err := ctx.ShouldBindBodyWithJSON(&dish); err != nil {
		dishService.WriteErr(ctx, err.Error())
		return
	}

	if err := dishService.ValidateAddDishInputPayload(dish); err != nil {
		dishService.WriteErr(ctx, err.Error())
		return
	}

	log.Warnfx("dish :: %#v", dish)

	dishService.AddDish(ctx, dish)
}

func AddDishes(ctx *gin.Context) {
	var dishes *[]pkg.AddDishDto

	if err := ctx.ShouldBindBodyWithJSON(&dishes); err != nil {
		dishService.WriteErr(ctx, err.Error())
		return
	}

	if err := dishService.ValidateAddDishesInputPayload(dishes); err != nil {
		dishService.WriteErr(ctx, err.Error())
		return
	}

	log.Warnfx("dishes :: %#v", dishes)

	dishService.AddDishes(ctx, dishes)
}

func DeleteDishes(ctx *gin.Context) {
	log.Infox("hello")

	var deleteDishDto pkg.DeleteDishDto

	if err := ctx.ShouldBindBodyWithJSON(&deleteDishDto); err != nil {
		dishService.WriteErr(ctx, err.Error())
		return
	}

	dishService.DeleteDishes(ctx, deleteDishDto.Ids)
}

// GetAvailableDishes function returns a list of available dishes
// by retrieving statistics from the db. Assume unavailability
// of dishes will be updated by the respective restaurants.
func GetAllDishes(ctx *gin.Context) {
	page, size, err := utils.Pagination(ctx)

	log.Debug("^GetAllDishes input param page:", page)
	log.Debug("^GetAllDishes input param size:", page)

	if err != nil {
		log.Error("Error while parsing integer from string:", err.Error())
		dishService.WriteErr(ctx, err.Error())
		return
	}

	dishService.GetAllDishes(ctx, page, size)
}

func PlaceOrder(ctx *gin.Context) {
	data, err := orderService.ValidatePlaceOrderInput(ctx)

	if err != nil {
		log.Error("Error while parsing place order input:", err.Error())
		orderService.WriteErr(ctx, err.Error())
		return
	}

	orderService.PlaceOrder(ctx, data)
}

func GetOrderTimer(ctx *gin.Context) {
	orderId := ctx.Param("id")

	if orderId == "" {
		timerService.WriteErr(ctx, "order id must not be empty")
		return
	}

	timerService.GetOrderTimer(ctx, orderId)
}
