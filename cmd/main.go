package main

import (
	"fmt"
	"moonwalk/internal/app"
	"moonwalk/internal/utils"
	"moonwalk/internal/config"
	"moonwalk/internal/api/rest"

	log "github.com/Thanga-tamil/logger_lib"
)

func main() {

	// load the service config from config.json file from the server 
	// and init all the required services using the loaded config
	conf := config.LoadConfig(utils.ConfigFile)

	fmt.Println("Initializing custom zap logger")

	log.NewLogger(utils.LogFile, conf.LogLevel)

	fmt.Println("Custom zap logger initialized successfully")

	if err := app.Start(conf); err != nil {
		// try one recovery for the collective good,
		// upon more than one failure startup, stop 
		// the service and start debugging 
		panic(err)
	}

	rest.Serve(utils.ServerAddr, conf.ServerMode)
}

