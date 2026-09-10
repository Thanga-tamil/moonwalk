package middleware

import (
	"time"

	log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"
)

func LoggerChain() gin.HandlerFunc {

	return func(c *gin.Context) {
		start := time.Now()

		log.Infof("Request HTTP Method: '%s' Path -----> '%s' Params: '%v'",
			c.Request.Method, c.Request.URL, c.Request.URL.Query())

		// Pre-handler phase
		c.Next()

		// Post-handler phase
		latency := time.Since(start)
		log.Infof("Request took %v", latency)
	}

}
