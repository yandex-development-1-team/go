package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/metrics"
)

// Metrics collects HTTP request metrics for Prometheus
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		statusCode := c.Writer.Status()
		method := c.Request.Method

		fullPath := c.FullPath()
		if fullPath == "" {
			fullPath = c.Request.URL.Path
		}

		statusStr := strconv.Itoa(statusCode)
		metrics.IncAPIRequests(method, fullPath, statusStr)
	}
}
