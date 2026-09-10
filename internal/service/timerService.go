package service

import (
	"net/http"
	"time"

	log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"

	"moonwalk/internal/repository"
	"moonwalk/pkg"
)

// GetOrderTimer returns the countdown timer payload for an order. It computes
// the total estimated duration, the time already elapsed and the remaining time
// so a customer can render a live countdown.
func GetOrderTimer(ctx *gin.Context, orderId string) {
	order, err := repository.GetOrder(orderId)
	if err != nil {
		log.Error("Error fetching order for timer:", err.Error())
		WriteErr(ctx, "order not found for the given orderId")
		return
	}

	now := time.Now()

	// estimated total duration is the gap between creation and the ETA
	timeEstimated := int64(order.Eta.Sub(order.CreatedAt).Seconds())
	if timeEstimated < 0 {
		timeEstimated = 0
	}

	timeElapsed := int64(now.Sub(order.CreatedAt).Seconds())
	if timeElapsed < 0 {
		timeElapsed = 0
	}

	timeRemaining := timeEstimated - timeElapsed
	if timeRemaining < 0 {
		timeRemaining = 0
	}

	timer := pkg.TimerDto{
		OrderId:       order.OrderId,
		DishId:        order.DishId,
		Status:        order.Status,
		Algorithm:     order.Alg,
		Eta:           order.Eta,
		TimeEstimated: timeEstimated,
		TimeElapsed:   timeElapsed,
		TimeRemaining: timeRemaining,
	}

	response := map[string]interface{}{
		"statusCode": 200,
		"message":    "Order countdown retrieved successfully",
		"data":       timer,
	}

	ctx.JSON(http.StatusOK, response)
}
