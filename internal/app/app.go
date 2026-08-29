package app

import (
	"moonwalk/internal/config"
	"moonwalk/internal/route"
	"moonwalk/internal/middleware"

	"github.com/gin-gonic/gin"
	log "github.com/Thanga-tamil/logger_lib"
)

func App() error {
	log.Info("Connecting to required external i/o services")
	if err := config.NewSqlite(); err != nil {
		return err
	}
	return nil
}

// start the http server with gin lib 
func Run(ADDR string) {
	serve := gin.New()
	// Attach middlewares to GIN 
	serve.Use(middleware.LoggerChain())

	v1Group := serve.Group("/api/v1")
	route.Router(v1Group)
	serve.Run(ADDR) 
}
