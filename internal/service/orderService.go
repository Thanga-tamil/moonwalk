package service

import (
	"moonwalk/internal/repository"
	"strconv"
	"time"

	log "github.com/Thanga-tamil/logger_lib"
	"github.com/gin-gonic/gin"
)

type Dish struct {
	Dish          string
	AvailableUpto time.Time
	CreatedAt     time.Time
}

func GetAllDishes(ctx *gin.Context) ([]Dish, int, error) {

	param := ctx.Request.URL.Query()
	pageStr := param.Get("page")
	sizeStr := param.Get("size")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		log.Error("Error while parsing integer from string:", err.Error())
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		log.Error("Error while parsing integer from string:", err.Error())
	}

	// Retrieve available dishes from db.
	// let the query take of pagination using offset page + page 
	// which will exclude the previous dataset
	rows, err := repository.GetAllDishes(page + page, size)
	if err != nil {
		log.Error("Error while retriving All Dishes from schema:", err.Error())
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
	}
	log.Info("totalRecords:", totalRecords)
	// return list of dish and totol records of dishes
	return dishes, totalRecords
}
