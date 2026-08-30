package rest

import (
	"strings"
	"moonwalk/internal/middleware"
	"moonwalk/internal/api/rest/route"
	"github.com/gin-gonic/gin"
)

// start the http server with gin lib 
func Serve(ADDR string) {

	// set in release mode
	setGinMode("")

	serve := gin.New()

	// Attach middlewares to GIN 
	serve.Use(middleware.LoggerChain())

	v1Group := serve.Group("/api/v1")

	route.Router(v1Group)

	serve.Run(ADDR) 
}

func setGinMode(env string) {
	switch strings.ToLower(env) {
	case "dev", "development":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}
