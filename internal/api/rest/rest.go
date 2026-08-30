package rest

import (
	"strings"
	"github.com/gin-gonic/gin"
	"moonwalk/internal/middleware"
	"moonwalk/internal/api/rest/route"
	log "github.com/Thanga-tamil/logger_lib"
)

// start and serve http server with gin lib 
func Serve(ADDR, serverMode string) {

	setGinMode(serverMode)

	serve := gin.New()

	// Attach middlewares to GIN 
	serve.Use(middleware.LoggerChain())

	v1Group := serve.Group("/api/v1")

	route.Router(v1Group)

	log.Infox("Application started successfully. Serving HTTP request response @ '", ADDR + "'")

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
