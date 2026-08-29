package main

import (
	"moonwalk/internal/route"
	"moonwalk/internal/middleware"

	"github.com/gin-gonic/gin"
	log "github.com/Thanga-tamil/logger_lib"
)

func main() {

	log.NewLogger("moonwalk.log")

	serve := gin.New()

	// Attach middlewares to GIN 
	serve.Use(middleware.LoggerChain())

	v1Group := serve.Group("/api/v1")

	route.Router(v1Group)

	serve.Run("0.0.0.0:8080")

}

