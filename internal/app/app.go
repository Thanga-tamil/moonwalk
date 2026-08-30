package app

import (
	"moonwalk/internal/config"
	log "github.com/Thanga-tamil/logger_lib"
)

func Start() error {
	log.Info("Connecting to required external i/o services")
	if err := config.NewSqlite(); err != nil {
		return err
	}
	return nil
}
