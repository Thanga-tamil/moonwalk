package service

import (
	"errors"
	"net/http"
	"github.com/gin-gonic/gin"
	log "github.com/Thanga-tamil/logger_lib"

	"moonwalk/pkg"
	"moonwalk/internal/repository"
)

func GetAllDishes(ctx *gin.Context, page, size int) {

	// since the pagination handled in query itself, we can't get the 
	// totalRecords from the retrieved dataset, so handle and return no 
	// records case before processing the data
	totalRecords, err := repository.TotalRecordsOfDishes()
	if err != nil {
		log.Error("Error while retriving data from schema:", err.Error())
		WriteErr(ctx, err.Error())
		return
	}

	log.Debug("^GetAllDishes totalRecords:", totalRecords)
	if totalRecords == 0 {
		ctx.JSON(http.StatusNoContent, "")
		return 
	}

	// Retrieve available dishes from db.
	// let the query take care of pagination using limit & offset 
	dishes, err := repository.GetAllDishes(page, size)
	if err != nil {
		log.Error("Error while retriving All Dishes from schema:", err.Error())
		WriteErr(ctx, err.Error())
		return
	}

	response := pkg.Success(200, "Data retrieved successfully", dishes, totalRecords, len(dishes))

	log.Debugf("^GetAllDishes response: %#v", response)

	ctx.JSON(http.StatusOK, response)
}

func WriteErr(ctx *gin.Context, err string) {
	response := pkg.Failure(400, err)
	ctx.JSON(http.StatusBadRequest, response)
}

func ValidatePlaceOrderInput(ctx *gin.Context) (*pkg.PlaceOrderDto, error) {
	var data pkg.PlaceOrderDto

	if err := ctx.ShouldBindBodyWithJSON(&data); err != nil {
		return nil, errors.New(err.Error())
	}

	if data.DishId < 1 {
		return nil, errors.New("dishId must not be empty and dishId must be greater than 0")
	}

	return &data, nil
}

func PlaceOrder(ctx *gin.Context, data *pkg.PlaceOrderDto) {
	dish, err := repository.GetDish(data.DishId)

	if err != nil {
		log.Error("Error while parsing place order input:", err.Error())
		WriteErr(ctx, err.Error()); return
	} else if dish.Dish == "" {
		WriteErr(ctx, "dish not found for the input dishId"); return
	}

	resources, err := repository.GetResources()

	if err != nil { 
		WriteErr(ctx, err.Error()); return
	}

	order := scheduler(dish, resources)

	repository.Save(&order)
}

