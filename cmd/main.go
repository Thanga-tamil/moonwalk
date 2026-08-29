package main

import (
	"fmt"
	"moonwalk/internal/app"
	log "github.com/Thanga-tamil/logger_lib"
)

func main() {

	fmt.Println("Initializing custom zap logger")

	log.NewLogger("moonwalk.log")

	fmt.Println("Custom zap logger initialized successfully")

	if err := app.App(); err != nil {
		// try one recovery for the collective good,
		// upon more than one failure startup, stop 
		// the service and start debugging 
		panic(err)
	}

	app.Run("0.0.0.0:8080")
}

