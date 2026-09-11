package config

import (
	log "github.com/Thanga-tamil/logger_v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewMySQL opens a connection to the MySQL database with a tuned connection
// pool so the server can run for a long time without leaking connections.
func NewMySQL(driverName, dataSourceName string, maxIdleConns, maxOpenConns int) (*gorm.DB, error) {
	log.Infof("Initialize mysql db")

	db, err := gorm.Open(mysql.Open(dataSourceName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if maxOpenConns > 0 {
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
	}

	log.Infox("MySQL connection established successfully")
	return db, nil
}
