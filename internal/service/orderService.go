package service

import (
	"fmt"
	"errors"
	"strconv"
	"strings"
	"net/http"
	"github.com/gin-gonic/gin"
	log "github.com/Thanga-tamil/logger_lib"

	"moonwalk/pkg"
	"moonwalk/internal/repository"
)

func GetAllDishes(ctx *gin.Context) {

	page, size, err := parseAndValidateGetAllDishesInput(ctx)

	log.Debug("^GetAllDishes input param page:", page)
	log.Debug("^GetAllDishes input param size:", page)

	if err != nil {
		log.Error("Error while parsing integer from string:", err.Error())
		writeErr(ctx, err.Error())
		return
	}

	// Retrieve available dishes from db.
	// let the query take of pagination using offset page + page 
	// which will exclude the previous dataset
	var dishes []pkg.Dish

	rows, err := repository.GetAllDishes(page, size)
	if err != nil {
		log.Error("Error while retriving All Dishes from schema:", err.Error())
		writeErr(ctx, err.Error())
		return
	}

	defer rows.Close()

	for rows.Next() {
		var dish pkg.Dish 

		err := rows.Scan(&dish.Id, &dish.Dish, &dish.IsAvailable, &dish.Price, 
						 &dish.AvailableUpto, &dish.CreatedAt, &dish.PrepTime)
		if err != nil {
			log.Debug("scan trace:", err.Error())
		}

		dishes = append(dishes, dish)
	}

	totalRecords, err := repository.TotalRecordsOfAvailableDishes()
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

	response := pkg.Success(200, "Data retrieved successfully", dishes, totalRecords, len(dishes))

	log.Debugf("^GetAllDishes response: %#v", response)

	ctx.JSON(http.StatusOK, response)
}

func writeErr(ctx *gin.Context, err string) {
	response := pkg.Failure(http.StatusBadRequest, err)
	ctx.JSON(http.StatusBadRequest, response)
}

func parseAndValidateGetAllDishesInput(ctx *gin.Context) (int, int, error) {
	const (
		defaultPage = 1
		defaultSize = 10
	)

	page := defaultPage
	size := defaultSize

	if pageStr := ctx.Query("page"); pageStr != "" {
		var err error

		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid page %q: %w", pageStr, err)
		}

		if page < 1 {
			return 0, 0, errors.New("page must be greater than 0")
		}
	}

	if sizeStr := ctx.Query("size"); sizeStr != "" {
		var err error

		size, err = strconv.Atoi(sizeStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid size %q: %w", sizeStr, err)
		}

		if size < 1 {
			return 0, 0, errors.New("size must be greater than 0")
		}
	}

	return page, size, nil
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
	} else if exists == 0 {
		writeErr(ctx, "dish not found for the input dishId")
		return
	}
	
	log.Infox(exists)
	log.Infofx("%#v\n", *data)
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
