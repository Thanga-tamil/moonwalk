package app

import (
	"moonwalk/pkg"
	"moonwalk/internal/config"

	"gorm.io/gorm"
	log "github.com/Thanga-tamil/logger_lib"
)

var DB *gorm.DB

func Start(conf *pkg.ServiceConfig) error {
	log.Info("Connecting to required external i/o services")

	var err error

	DB, err = config.NewSqlite(conf.SqlDriverName, conf.SqlDataSourceName);
	if  err != nil {
		return err
	}

	log.Infox("Connection established with all required external i/o services successfully")
	return nil
}
