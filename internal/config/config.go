package config

import (
	"os"
	"log"
	"moonwalk/pkg"
	"encoding/json"
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

// applyDefaults fills in sensible values for any config field that the user did
// not provide, and validates the scheduler strategy so a misconfiguration does
// not fail the server at request time.
func applyDefaults(c *pkg.ServiceConfig) {
	if c.SchedulerStrategy == "" {
		c.SchedulerStrategy = pkg.StrategyAuto
	}
	switch c.SchedulerStrategy {
	case pkg.StrategyAuto, pkg.StrategyFIFO, pkg.StrategyResourceAware:
	default:
		log.Fatalf("invalid schedulerStrategy %q; use one of auto, fifo, resource_aware", c.SchedulerStrategy)
	}

	if c.DbMaxOpenConns == 0 {
		c.DbMaxOpenConns = 10
	}
	if c.DbConnMaxLifetime == 0 {
		c.DbConnMaxLifetime = 1800 // 30 minutes
	}
}

