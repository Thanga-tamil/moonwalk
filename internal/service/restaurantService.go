package service

import (
	"errors"
	"net/http"
	"strconv"

	log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"moonwalk/internal/app"
	chefsRepo "moonwalk/internal/repository"
	dishesRepo "moonwalk/internal/repository"
	ordersRepo "moonwalk/internal/repository"
	"moonwalk/pkg"
)

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
	err := app.DB.Transaction(func(tx *gorm.DB) error {
		dish, err := dishesRepo.GetDish(data.DishId)

		if err != nil {
			log.Error("Error while parsing place order input:", err.Error())
			WriteErr(ctx, err.Error())
			return err
		} else if dish.Dish == "" {
			WriteErr(ctx, "dish not available for the input dishId")
			return err
		}

		var resource any
		var order *pkg.Order

		if dish.PreCooked {
			// find supplier for fifo
			// supplier, ok := resource.(pkg.Suppliers)
			// 	if !ok {
			// 		errMsg := "Unable to convert resource to specified type 'chef'"
			// 		WriteErr(ctx, errMsg)
			// 		return errors.New(errMsg)
			// 	}
			// handlePreCookedOrder()

		} else {
			resource, err = chefsRepo.FindAvailableChef(tx)
			if err != nil {
				WriteErr(ctx, "Error whlie fetching chef")
				return err
			}

			chef, ok := resource.(*pkg.Chefs)
			if !ok {
				errMsg := "Unable to convert resource to specified type 'chef'"
				WriteErr(ctx, errMsg)
				return errors.New(errMsg)
			}

			order, err = handleResourceAwareOrder(tx, chef, &dish)
			if err != nil {
				WriteErr(ctx, err.Error())
				return err
			}
		}

		log.Infox("order: ", order)

		response := map[string]interface{}{
			"statusCode": 200,
			"message":    "Order placed successfully",
			"data":       order,
		}
		ctx.JSON(http.StatusOK, response)
		return nil
	})
	if err != nil {
		log.Error("Error while placing order:", err.Error())
		WriteErr(ctx, err.Error())
		return
	}
}

func handleResourceAwareOrder(tx *gorm.DB, chef *pkg.Chefs, dish *pkg.Dish) (*pkg.Order, error) {

	order := resourceAwareEtaScheduler(dish, chef)

	// audit pending status of the order
	recordExecution(tx, &order)
	if err := ordersRepo.Insert(tx, &order); err != nil {
		return nil, err
	}

	if chef.Status == IDLE {
		order.Status = "PREPARING"
		order.ResourceType = CHEF
		ordersRepo.UpdateOrder(tx, &order)

		// audit preparing status of the order
		recordExecution(tx, &order)
	}

	if chef.Status == IDLE {
		chef.CurrentOrderID = order.OrderId
		chef.Status = BUSY
	}
	if err := chefsRepo.UpdateChef(tx, chef); err != nil {
		return nil, err
	}

	return &order, nil
}
