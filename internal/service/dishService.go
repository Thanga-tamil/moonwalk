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
	"github.com/redis/go-redis/v9"

	"moonwalk/internal/app"
	dishesRepo "moonwalk/internal/repository"
	"moonwalk/pkg"
)

var backgroundCtx = context.Background()
var AddNewDishBatchSize = 0
var RunTimeCacheProcessingBatchSize = 100

const (
	DISHES = "dishes"
)

/*
GetAllDishes function retrieves paginated dishes from redis cache and return.
If dishes not found in redis, retrieve paginated dishes from database and upon
successfull retrieval from database, data will be cached in redis to reduce
external IO's on the next API call.
*/

func GetDishes(ctx *gin.Context, page, size int) {

	totalRecords, err := app.Redis.ZCard(ctx, DISHES).Result()
	if err != nil {
		WriteErr(ctx, err.Error())
		return
	}

	var dishes []*pkg.Dish

	if totalRecords == 0 {
		totalRecords, dishes, err = fetchDishesFromDB(page, size)
		if totalRecords == 0 {
			ctx.JSON(http.StatusNoContent, "")
			return
		} else if err != nil {
			WriteErr(ctx, err.Error())
			return
		}
	} else {
		dishes, err = fetchDishesFromRedis(dishes, page, size)
		if err != nil {
			WriteErr(ctx, err.Error())
			return
		}
	}

	totalPages := totalRecords / int64(size)
	if totalRecords%int64(size) > 0 {
		totalPages++
	}

	response := pkg.Success{
		StatusCode:   200,
		Data:         dishes,
		Message:      "Data retrieved successfully",
		Count:        int16(len(dishes)),
		TotalPages:   int16(totalPages),
		TotalRecords: int16(totalRecords),
	}

	ctx.JSON(http.StatusOK, response)
}

func fetchDishesFromDB(page, size int) (int64, []*pkg.Dish, error) {

	log.Info("No dishes found in redis :: check and fetch dishes from database")

	totalRecords, err := dishesRepo.TotalRecordsOfDishes()
	if err != nil {
		return -1, nil, errors.New("Error while retriving data from schema:" + err.Error())
	}

	if totalRecords == 0 {
		return 0, nil, nil
	}

	dishes, err := dishesRepo.GetPaginatedDishes(page, size)
	if err != nil {
		return -1, nil, errors.New("Error while retriving All Dishes from schema:" + err.Error())
	}

	go cacheDishesInRedis()

	return totalRecords, dishes, nil

}

// Retrieve and process in batch
func cacheDishesInRedis() {

	page := 1
	for {
		dishes, err := dishesRepo.GetPaginatedDishes(page, RunTimeCacheProcessingBatchSize)
		page = (page + 1)

		if err != nil {
			log.Error("Error while retrieving dishes from database for processing redis cache")
			log.Error("Err:", err.Error())
			return
		} else if len(dishes) < 1 {
			log.Warnx("No dishes found in database for processing redis cache")
			return
		}

		log.Infofx("Caching %d dishes in redis", len(dishes))

		for _, dish := range dishes {
			mapValue, err := json.Marshal(&dish)
			if err != nil {
				log.Error(err.Error())
				return
			}

			// store dishes in redis cache to reduce external I/O while retrieving dishes
			sortedSetMem := redis.Z{Score: float64(dish.Id), Member: string(mapValue)}
			if err := app.Redis.ZAdd(backgroundCtx, DISHES, sortedSetMem).Err(); err != nil {
				log.Error(err.Error())
				return
			}
		}

		log.Info("Dishes cached in redis successfully")
	}

}

func fetchDishesFromRedis(dishes []*pkg.Dish, page, size int) ([]*pkg.Dish, error) {
	log.Info("Fetching dishes from redis")

	start := int64((page - 1) * size)
	stop := start + int64(size) - 1

	result, err := app.Redis.ZRange(backgroundCtx, DISHES, start, stop).Result()
	if err != nil {
		return nil, err
	}

	dishes = make([]*pkg.Dish, 0, len(result))

	for _, item := range result {
		var dish pkg.Dish

		if err := json.Unmarshal([]byte(item), &dish); err != nil {
			return nil, err
		}

		dishes = append(dishes, &dish)
	}

	return dishes, nil
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
		sortedSetMem := redis.Z{Score: float64(dish.Id), Member: string(mapValue)}
		if err := app.Redis.ZAdd(ctx, DISHES, sortedSetMem).Err(); err != nil {
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

func ValidateUpdateDishDtoPayload(updateDishDto pkg.UpdateDishDto) error {

	if updateDishDto.Id < 1 {
		return errors.New("Dish Id must be greater than 0")
	} else if updateDishDto.Price < 1 {
		return errors.New("price must be greater than 0")
	} else if updateDishDto.PreCooked && updateDishDto.PrepTime < 1 {
		return errors.New("Precooked dish prepTime must be greater than 0")
	}

	return nil

}

func UpdateDish(ctx *gin.Context, updateDishDto pkg.UpdateDishDto) {

}

func DeleteDishes(ctx *gin.Context, dishIds []int) {

	deletedRecords, err := dishesRepo.DeleteDishes(dishIds)
	if err != nil {
		WriteErr(ctx, err.Error())
		return
	}

	var msg string

	if deletedRecords == 0 {
		msg = "No dishes available for the input ids provided"
	} else {
		msg = "Dishes deleted successfully"
	}

	response := map[string]interface{}{
		"statusCode": 200,
		"message":    msg,
	}
	ctx.JSON(http.StatusOK, response)

}
