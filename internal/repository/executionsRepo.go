package repository

import (
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_v2"
	"gorm.io/gorm"
)

// RecordExecution persists a single order status transition into the
// order_executions audit table. TimeEstimated and TimeElapsed are stored in
// seconds. Failures are logged and swallowed so an audit write never blocks
// the core order flow.
func RecordExecution(tx *gorm.DB, e *pkg.OrderExec) {
	if err := tx.Table("order_exec").Create(e).Error; err != nil {
		log.Error("Error recording order execution audit:", err.Error())
	}
}

func RecordExecutions(tx *gorm.DB, e *[]pkg.OrderExec) {
	if err := tx.Table("order_exec").Create(e).Error; err != nil {
		log.Error("Error recording order execution audit:", err.Error())
	}
}
