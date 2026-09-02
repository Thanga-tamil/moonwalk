package repository

import (
	"moonwalk/internal/app"
	"moonwalk/pkg"

	log "github.com/Thanga-tamil/logger_lib"
)

// RecordExecution persists a single order status transition into the
// order_executions audit table. TimeEstimated and TimeElapsed are stored in
// seconds. Failures are logged and swallowed so an audit write never blocks
// the core order flow.
func RecordExecution(e *pkg.OrderExecution) {
	if err := app.DB.Table("order_executions").Create(e).Error; err != nil {
		log.Error("Error recording order execution audit:", err.Error())
	}
}
