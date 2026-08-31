package service

import (
	"errors"
	"net/http"
	"strings"
	"time"

	log "github.com/Thanga-tamil/logger_lib"
	"github.com/gin-gonic/gin"

	"moonwalk/internal/repository"
	"moonwalk/internal/utils"
	"moonwalk/pkg"
)

func GetAllDishes(ctx *gin.Context) {

	page, size, err := utils.Pagination(ctx)

	log.Debug("^GetAllDishes input param page:", page)
	log.Debug("^GetAllDishes input param size:", page)

	if err != nil {
		log.Error("Error while parsing integer from string:", err.Error())
		writeErr(ctx, err.Error())
		return
	}

	// since the pagination handled in query itself, we can't get the 
	// totalRecords from the retrieved dataset, so handle and return no 
	// records case before processing the data
	totalRecords, err := repository.TotalRecordsOfDishes()
	if err != nil {
		log.Error("Error while retriving data from schema:", err.Error())
		writeErr(ctx, err.Error())
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
		writeErr(ctx, err.Error())
		return
	}

	response := pkg.Success(200, "Data retrieved successfully", dishes, totalRecords, len(dishes))

	log.Debugf("^GetAllDishes response: %#v", response)

	ctx.JSON(http.StatusOK, response)
}

func writeErr(ctx *gin.Context, err string) {
	response := pkg.Failure(400, err)
	ctx.JSON(http.StatusBadRequest, response)
}

func PlaceOrder(ctx *gin.Context) {

	data, err := parseAndValidatePlaceOrderInput(ctx)
	if err != nil {
		log.Error("Error while parsing place order input:", err.Error())
		writeErr(ctx, err.Error())
		return
	}

	exists, err := repository.IsDishExists(data.DishId)
	if err != nil {
		log.Error("Error while parsing place order input:", err.Error())
		writeErr(ctx, err.Error())
		return
	} else if exists == false {
		writeErr(ctx, "dish not found for the input dishId")
		return
	}

	log.Infox(exists)
	log.Infofx("%#v\n", *data)

	order :=
	pkg.Order{
		OrderId: data.OrderId, 
		DishId: data.DishId, 
		PrepTime: "15", 
		CreatedAt: time.Now(),
	}

	repository.Save(&order)
}

func parseAndValidatePlaceOrderInput(ctx *gin.Context) (*pkg.PlaceOrderDto, error) {

	var dishId int 
	var orderId string
	var data pkg.PlaceOrderDto

	ctx.ShouldBindBodyWithJSON(&data)

	if dishId = data.DishId; dishId < 1 {
		return nil, errors.New("dishId must not be empty and dishId must be greater than 0")
	}

	if orderId = data.OrderId; strings.TrimSpace(orderId) == "" {
		return nil, errors.New("orderId must not be empty")
	}

	return &data, nil
}
