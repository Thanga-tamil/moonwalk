package route

import (
	"moonwalk/internal/api/rest/handler"
	"github.com/gin-gonic/gin"
)

func Router(ctx *gin.RouterGroup) {

	ctx.GET("/dishes", handler.GetAllDishes)
	ctx.POST("/order", handler.PlaceOrder)

}
