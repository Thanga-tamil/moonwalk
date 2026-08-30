package app

import (
	"moonwalk/internal/config"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_lib"
)

func Start(conf *pkg.ServiceConfig) error {
	log.Info("Connecting to required external i/o services")
	if err := config.NewSqlite(conf.SqlDriverName, conf.SqlDataSourceName); err != nil {
		return err
	}
	return nil
}
