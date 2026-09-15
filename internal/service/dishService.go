package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"

	"moonwalk/internal/app"
	dishesRepo "moonwalk/internal/repository"
	"moonwalk/pkg"
)

var backgroundCtx = context.Background()
var AddNewDishBatchSize = 0

// GetAvailableDishes function returns a list of available dishes
// by retrieving statistics from the db. Assume unavailability
// of dishes will be updated by the respective restaurants.
func GetAllDishes(ctx *gin.Context, page, size int) {

	// since the pagination handled in query itself, we can't get the
	// totalRecords from the retrieved dataset, so handle and return no
	// records case before processing the data
	totalRecords, err := dishesRepo.TotalRecordsOfDishes()
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
	dishes, err := dishesRepo.GetAllDishes(page, size)
	if err != nil {
		log.Error("Error while retriving All Dishes from schema:", err.Error())
		WriteErr(ctx, err.Error())
		return
	}

	totalPages := totalRecords / int64(size)
	if totalRecords%int64(size) > 0 {
		totalPages++
	}
	if int64(page) > totalPages {
		WriteErr(ctx, "Page limit exceeded, Total pages available: "+strconv.FormatInt(totalPages, 10))
		return
	}
	response := pkg.Success(200, "Data retrieved successfully", dishes, totalRecords, len(dishes), totalPages)

	log.Debugf("^GetAllDishes response: %#v", response)

	ctx.JSON(http.StatusOK, response)
}

func ValidateAddDishInputPayload(dish *pkg.AddDishDto) error {
	if strings.TrimSpace(dish.Dish) == "" {
		return errors.New("'dish' can not be empty or null")
	} else if dish.Price <= 0 {
		return errors.New("'price' must be greater than 0")
	} else if !dish.PreCooked && dish.PrepTime <= 0 {
		return errors.New("none precooked dishes 'prepTime' must be greater than 0")
	} else if dish.PreCooked && dish.PrepTime != 0 {
		return errors.New("precooked dishes 'prepTime' must be 0")
	} else {
		return nil
	}
}

func AddDish(ctx *gin.Context, dishPayload *pkg.AddDishDto) {

	dish := &pkg.Dish{
		Dish:        dishPayload.Dish,
		Price:       dishPayload.Price,
		IsAvailable: dishPayload.IsAvailable,
		PrepTime:    dishPayload.PrepTime,
		PreCooked:   dishPayload.PreCooked,
		CreatedAt:   time.Now(),
	}

	if err := dishesRepo.InsertDish(dish); err != nil {
		WriteErr(ctx, err.Error())
		return
	}

	mapValue, err := json.Marshal(&dish)
	if err != nil {
		WriteErr(ctx, err.Error())
		return
	}

	// store dish in redis cache to reduce external I/O while retrieving dishes
	mapName := "dishes"
	mapKey := dish.Id
	if err := app.Redis.HSet(ctx, mapName, mapKey, mapValue).Err(); err != nil {
		WriteErr(ctx, err.Error())
		return
	}

	response := map[string]interface{}{
		"statusCode": 200,
		"dishId":     mapKey,
		"message":    "Dish added successfully",
	}

	ctx.JSON(http.StatusOK, response)
}

func ValidateAddDishesInputPayload(dishes *[]pkg.AddDishDto) error {
	if len(*dishes) > AddNewDishBatchSize {
		return errors.New("Maximum " + strconv.Itoa(AddNewDishBatchSize) + " dishes can be added at a time")
	}
	var errMsg string = ""
	for _, dish := range *dishes {
		if strings.TrimSpace(dish.Dish) == "" {
			errMsg = "'dish' can not be empty or null"
			break
		} else if dish.Price <= 0 {
			errMsg = "'price' must be greater than 0 for dish: " + dish.Dish
			break
		} else if !dish.PreCooked && dish.PrepTime <= 0 {
			errMsg = "none precooked dishes 'prepTime' must be greater than 0 for dish: " + dish.Dish
			break
		} else if dish.PreCooked && dish.PrepTime != 0 {
			errMsg = "precooked dishes 'prepTime' must be 0 for dish: " + dish.Dish
			break
		}
	}
	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

func AddDishes(ctx *gin.Context, dishesPayload *[]pkg.AddDishDto) {

	dishes := []*pkg.Dish{}
	for _, dish := range *dishesPayload {

		dish := pkg.Dish{
			Dish:        dish.Dish,
			Price:       dish.Price,
			IsAvailable: dish.IsAvailable,
			PrepTime:    dish.PrepTime,
			PreCooked:   dish.PreCooked,
			CreatedAt:   time.Now(),
		}

		dishes = append(dishes, &dish)
	}

	if err := dishesRepo.InsertDishes(dishes); err != nil {
		WriteErr(ctx, err.Error())
		return
	}

	for _, dish := range dishes {
		mapValue, err := json.Marshal(&dish)
		if err != nil {
			WriteErr(ctx, err.Error())
			return
		}

		// store dishes in redis cache to reduce external I/O while retrieving dishes
		mapName := "dishes"
		mapKey := dish.Id
		if err := app.Redis.HSet(ctx, mapName, mapKey, mapValue).Err(); err != nil {
			WriteErr(ctx, err.Error())
			return
		}
	}

	response := map[string]interface{}{
		"statusCode": 200,
		"message":    "Dishes added successfully",
	}

	ctx.JSON(http.StatusOK, response)
}
