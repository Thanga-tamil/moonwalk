package service

import (
	"fmt"
	"time"
	"errors"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"

	"moonwalk/pkg"
	"moonwalk/internal/repository"
	log "github.com/Thanga-tamil/logger_lib"
)

type Dish struct {
	Dish          string
	AvailableUpto time.Time
	CreatedAt     time.Time
}

func GetAllDishes(ctx *gin.Context) {

	page, size, err := parseAndValidateGetAllDishesInput(ctx)
	log.Info("page:", page)
	log.Info("size:", size)

	if err != nil {
		log.Error("Error while parsing integer from string:", err.Error())
		writeErr(ctx, err.Error())
		return
	}

	// Retrieve available dishes from db.
	// let the query take of pagination using offset page + page 
	// which will exclude the previous dataset
	rows, err := repository.GetAllDishes(page, size)
	if err != nil {
		log.Error("Error while retriving All Dishes from schema:", err.Error())
		writeErr(ctx, err.Error())
		return
	}

	var dishes []Dish
	defer rows.Close()

	for rows.Next() {
		var dish Dish 
		rows.Scan(&dish.Dish, &dish.AvailableUpto, &dish.CreatedAt)

		dishes = append(dishes, dish)
	}

	totalRecords, err := repository.TotalRecordsOfAvailableDishes()

	if err != nil {
		log.Error("Error while retriving data from schema:", err.Error())
		writeErr(ctx, err.Error())
		return
	}

	log.Info("totalRecords:", totalRecords)

	if totalRecords == 0 {
		ctx.JSON(http.StatusNoContent, "")
		return 
	}

	response := pkg.Success(200, "Data retrieved successfully", dishes, totalRecords, len(dishes))

	ctx.JSON(http.StatusOK, response)
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

func writeErr(ctx *gin.Context, err string) {
	response := pkg.Failure(http.StatusBadRequest, err)
	ctx.JSON(http.StatusBadRequest, response)
}
