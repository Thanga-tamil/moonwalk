package config

import (
	"encoding/json"
	"log"
	"moonwalk/pkg"
	"os"
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

	applyDefaults(&c)

	return &c
}

func applyDefaults(c *pkg.ServiceConfig) {
	if c.DbMaxOpenConns == 0 {
		c.DbMaxOpenConns = 10
	}
	if c.DbConnMaxLifetime == 0 {
		c.DbConnMaxLifetime = 1800 // 30 minutes
	}
}
