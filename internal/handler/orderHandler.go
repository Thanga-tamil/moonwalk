package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func PlaceOrder(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, map[string]string{"message": "Order placed successfully", "orderId": "?"})

}
