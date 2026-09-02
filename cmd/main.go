package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"moonwalk/internal/app"
	"moonwalk/internal/utils"
	"moonwalk/internal/config"
	"moonwalk/internal/api/rest"
	"moonwalk/internal/service"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_lib"
)

func main() {
	// load the service config from config.json file from the server
	// and init all the required services using the loaded config
	conf := config.LoadConfig(utils.ConfigFile)

	fmt.Println("Initializing custom zap logger")

	log.NewLogger(utils.LogFile, conf.LogLevel)

	fmt.Println("Custom zap logger initialized successfully")

	// apply the server-wide scheduling strategy from config before any
	// request can be served
	service.SetSchedulerStrategy(conf.SchedulerStrategy)

	if err := app.Start(conf); err != nil {
		// try one recovery for the collective good,
		// upon more than one failure startup, stop
		// the service and start debugging
		panic(err)
	}

	// start the background cron service handling order
	// completion and pending order re-scheduling
	service.StartCronService()

	serveAsync(utils.ServerAddr, conf.ServerMode, conf)
}

// serveAsync starts the HTTP server in the background and blocks until either
// the server fails or an OS shutdown signal (SIGINT/SIGTERM) is received, in
// which case the server and database are shut down gracefully.
func serveAsync(addr, serverMode string, conf *pkg.ServiceConfig) {
	server, errChan := rest.Serve(addr, serverMode)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error:", err.Error())
		}
	case sig := <-quit:
		log.Infox("Received shutdown signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error("Error during server shutdown:", err.Error())
		}
	}

	// release the database connection pool last so no work is dropped mid-flight
	app.Close()
	log.Info("Application shutdown complete")
}
