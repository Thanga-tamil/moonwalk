package rest

import (
	"moonwalk/internal/api/rest/route"
	"moonwalk/internal/middleware"
	"net/http"
	"strings"
	"time"

	log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"
)

// Serve starts and returns the HTTP server (and an error channel) so the caller
// can either block on server failures or gracefully shut it down on signals.
func Serve(ADDR, serverMode string) (*http.Server, <-chan error) {

	setGinMode(serverMode)

	serve := gin.New()

	// Attach middlewares to GIN
	serve.Use(middleware.LoggerChain())
	log.Debug("Attaching LoggerChain middleware to GIN")
	serve.Use(middleware.TenantResolver())
	log.Debug("Attaching TenantResolver middleware to GIN")
	serve.Use(gin.Recovery())
	log.Infox("Middlewares attached to GIN successfull")

	v1Group := serve.Group("/api/v1")

	route.Router(v1Group)

	log.Infox("Server started successfully.", "ServerMode", serverMode)
	log.Infox("Serving HTTP request response.", "@ADDRESS", ADDR)

	httpServer := &http.Server{
		Addr:         ADDR,
		Handler:      serve,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	return httpServer, errChan
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
