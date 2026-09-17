package route

import (
	dishHandler "moonwalk/internal/api/rest/handler"
	orderHandler "moonwalk/internal/api/rest/handler"

	"github.com/gin-gonic/gin"
)

func Router(v1Group *gin.RouterGroup) {

	dish := v1Group.Group("/api/v1/dish")
	{
		dish.POST("", dishHandler.AddDishes)
		dish.GET("", dishHandler.GetDishes)
		dish.PUT("", dishHandler.UpdateDish)
		dish.DELETE("", dishHandler.DeleteDishes)
	}

	order := v1Group.Group("/api/v1/order")
	{
		order.POST("", orderHandler.PlaceOrder)
		order.GET("/:id", orderHandler.GetOrderTimer)
	}

}
