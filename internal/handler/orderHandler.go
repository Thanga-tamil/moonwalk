package handler

import (
	"moonwalk/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAvailableDishes function returns a list of available dishes
// by retrieving statistics from the db. Assume unavailability
// of dishes will be updated by the respective restaurants.
func GetAvailableDishes(ctx *gin.Context) {

	dishes := repository.GetAvailableDishes()
	totalRecords := len(repository.GetAvailableDishes())

	if totalRecords == 0 {
		ctx.JSON(http.StatusNoContent, ""); return 
	}

	success := map[string]any{
						"message": "Available dishes retrieved successfully",
						"data": dishes,
						"totalRecords": len(dishes),
					}

	ctx.JSON(http.StatusOK, success)
}

func PlaceOrder(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, map[string]string{"message": "Order placed successfully", "orderId": "?"})

}
