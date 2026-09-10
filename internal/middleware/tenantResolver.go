package middleware

import (
	// log "github.com/Thanga-tamil/logger_v2"
	"github.com/gin-gonic/gin"
)

func TenantResolver() gin.HandlerFunc {
	return func(c *gin.Context) {
		// tenantId := c.Request.Header.Get("tenant-x")
		// path := c.Request.URL.Path
		// method := c.Request.Method

		// if tenantId == "" {
		// 	log.Warnf("missing tenant header. path: %s method: %s clientIP: %s", path, method, c.ClientIP())
		// 	c.AbortWithStatusJSON(400, gin.H{"error": "missing tenant-x header"})
		// 	return
		// }

		// db, ok := config.TenantDBs[tenantId]
		// if !ok {
		// 	log.Warnf("unknown tenant. path: %s method: %s clientIP: %s", path, method, c.ClientIP())
		// 	c.AbortWithStatusJSON(401, gin.H{"error": "invalid tenant"})
		// 	return
		// }

		// log.Warnf("tenant db resolved. tenant_id: %s path: %s method: %s", tenantId, path, method)

		// c.Set(tenantId, db)
		c.Next()
	}
}
