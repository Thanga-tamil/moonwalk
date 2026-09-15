package route

import (
	"moonwalk/internal/api/rest/handler"

	"github.com/gin-gonic/gin"
)

func Router(serve *gin.Engine) {

	dish := serve.Group("/api/v1/dish")
	{
		dish.POST("", handler.AddDish)
		dish.POST("/dishes", handler.AddDishes)
		// dish.PUT("/update", handler.GetAllDishes)
		// dish.DELETE("/", handler.GetAllDishes)
		dish.GET("/dishes", handler.GetAllDishes)
	}

	order := serve.Group("/api/v1/order")
	{
		order.POST("", handler.PlaceOrder)
		order.GET("/:id", handler.GetOrderTimer)
	}

}
