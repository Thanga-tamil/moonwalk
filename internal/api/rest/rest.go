package rest

import (
	"moonwalk/internal/middleware"
	"moonwalk/internal/api/rest/route"

	"github.com/gin-gonic/gin"
)

// start the http server with gin lib 
func Serve(ADDR string) {
	serve := gin.New()

	// Attach middlewares to GIN 
	serve.Use(middleware.LoggerChain())

	v1Group := serve.Group("/api/v1")

	route.Router(v1Group)

	serve.Run(ADDR) 
}
