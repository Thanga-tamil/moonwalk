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

	DB, err = config.NewSqlite(conf.SqlDriverName, conf.SqlDataSourceName,
		conf.DbMaxOpenConns, conf.DbConnMaxLifetime)
	if err != nil {
		return err
	}

	log.Infox("Connection established with all required external i/o services successfully")
	return nil
}

// Close gracefully releases the underlying database connection pool. Call this
// on shutdown so no connections are leaked while the server winds down.
func Close() {
	if DB == nil {
		return
	}
	sqlDB, err := DB.DB()
	if err != nil {
		log.Error("Error getting underlying sql.DB on close:", err.Error())
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Error("Error closing database:", err.Error())
		return
	}
	log.Info("Database connection pool closed successfully")
}
