package main

import (
	"context"
	"fmt"
	"moonwalk/internal/api/rest"
	"moonwalk/internal/app"
	"moonwalk/internal/config"
	"moonwalk/internal/service"
	"moonwalk/internal/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Thanga-tamil/logger_v2"
)

func main() {
	// load the service config from config.json file from the server
	// and init all the required services using the loaded config
	conf := config.LoadConfig(utils.ConfigFile)

	fmt.Println("Initializing custom Charmbracelet logger with logLevel:", conf.LogLevel)

	setLogger(conf.LogLevel, conf.LogWriteToFile, conf.LogFile)

	fmt.Println("Custom Charmbracelet logger initialized successfully")

	if err := app.Start(conf); err != nil {
		// try one recovery for the collective good,
		// upon more than one failure startup, stop
		// the service and start debugging
		panic(err)
	}

	// start the background cron service handling order
	// completion and pending order re-scheduling
	cronInterval := time.Duration(conf.CronInterval) * time.Second
	service.StartCronService(cronInterval, conf.PendingOrdersBatchSize)

	serveAsync(conf.ServerHost+":"+fmt.Sprint(conf.ServerPort), conf.ServerMode)
}

/* Log levels 0 = ERROR | 1 = INFO | 2 = DEBUG | 3 = WARN */
func setLogger(logLevel int, writeToFile bool, logFile string) {

	var level log.Level

	switch logLevel {
	case 0:
		level = log.LevelError
	case 1:
		level = log.LevelInfo
	case 2:
		level = log.LevelDebug
	case 3:
		level = log.LevelWarn
	default:
		level = log.LevelInfo
	}

	log.Configure(log.Options{
		Level:           level,
		Formatter:       log.TextFormatter,
		ReportTimestamp: true,
		WriteToFile:     writeToFile,
		FilePath:        logFile,
	})

}

// serveAsync starts the HTTP server in the background and blocks until either
// the server fails or an OS shutdown signal (SIGINT/SIGTERM) is received, in
// which case the server and database are shut down gracefully.
func serveAsync(addr, serverMode string) {
	server, errChan := rest.Serve(addr, serverMode)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error:", err.Error())
		}
	case sig := <-quit:
		log.Infox("", "Received shutdown signal", sig.String())

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
