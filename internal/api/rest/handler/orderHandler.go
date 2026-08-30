package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"moonwalk/internal/service"
)

// GetAvailableDishes function returns a list of available dishes
// by retrieving statistics from the db. Assume unavailability
// of dishes will be updated by the respective restaurants.
func GetAvailableDishes(ctx *gin.Context) {

	service.GetAllDishes(ctx)

}


func PlaceOrder(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, map[string]string{"message": "Order placed successfully", "orderId": "?"})
}
