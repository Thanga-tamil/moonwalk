package route

import (
	"moonwalk/internal/handler"

	"github.com/gin-gonic/gin"
)

func Router(ctx *gin.RouterGroup) {

	ctx.POST("/order", handler.PlaceOrder)

}
