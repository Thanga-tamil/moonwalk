package config

import (
	"os"
	"moonwalk/pkg"
	"encoding/json"
	log "github.com/Thanga-tamil/logger_lib"
)

func LoadConfig(path string) *pkg.ServiceConfig {

	file, err := os.Open(path)

	if err != nil {
		log.Fatalf("Error while opening %s file \n", err.Error())
	}

	decoder := json.NewDecoder(file)

	var c pkg.ServiceConfig

	if err = decoder.Decode(&c); err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	log.Infof("decoded config.json file: %#v\n", c)

	return &c
}

