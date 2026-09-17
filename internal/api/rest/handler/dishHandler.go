package handler

import (
	dishService "moonwalk/internal/service"
	"moonwalk/internal/utils"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"
)

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

// GetAvailableDishes function returns a list of available dishes
// by retrieving statistics from the db. Assume unavailability
// of dishes will be updated by the respective restaurants.
func GetDishes(ctx *gin.Context) {
	page, size, err := utils.Pagination(ctx)

	log.Debug("", "^GetDishes input param page:", page)
	log.Debug("", "^GetAlGetDisheslDishes input param size:", size)

	if err != nil {
		log.Error("Error while parsing integer from string:", err.Error())
		dishService.WriteErr(ctx, err.Error())
		return
	}

	dishService.GetDishes(ctx, page, size)
}

func UpdateDish(ctx *gin.Context) {}

func DeleteDishes(ctx *gin.Context) {
	log.Infox("hello")

	var deleteDishDto pkg.DeleteDishDto

	if err := ctx.ShouldBindBodyWithJSON(&deleteDishDto); err != nil {
		dishService.WriteErr(ctx, err.Error())
		return
	}

	dishService.DeleteDishes(ctx, deleteDishDto.Ids)
}
