package main

import (
	"fmt"
	"moonwalk/internal/api/rest"
	"moonwalk/internal/app"

	log "github.com/Thanga-tamil/logger_lib"
)

func main() {

	fmt.Println("Initializing custom zap logger")

	log.NewLogger("moonwalk.log")

	fmt.Println("Custom zap logger initialized successfully")

	if err := app.Start(); err != nil {
		// try one recovery for the collective good,
		// upon more than one failure startup, stop 
		// the service and start debugging 
		panic(err)
	}

	rest.Serve("0.0.0.0:8080")
}

