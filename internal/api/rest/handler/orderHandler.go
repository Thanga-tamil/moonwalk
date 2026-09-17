package handler

import (
	orderService "moonwalk/internal/service"
	timerService "moonwalk/internal/service"

	log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"
)

func PlaceOrder(ctx *gin.Context) {
	data, err := orderService.ValidatePlaceOrderInput(ctx)

	if err != nil {
		log.Error("Error while parsing place order input:", err.Error())
		orderService.WriteErr(ctx, err.Error())
		return
	}

	orderService.PlaceOrder(ctx, data)
}

func GetOrderTimer(ctx *gin.Context) {
	orderId := ctx.Param("id")

	if orderId == "" {
		timerService.WriteErr(ctx, "order id must not be empty")
		return
	}

	timerService.GetOrderTimer(ctx, orderId)
}
