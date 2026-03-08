package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yandex-development-1-team/go/internal/logger"
	"go.uber.org/zap"
)

// Logger logs incoming HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		userAgent := c.Request.UserAgent()
		contentLength := c.Writer.Size()

		fullPath := path
		if query != "" {
			fullPath = path + "?" + query
		}

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", fullPath),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user_agent", userAgent),
			zap.Int("size", contentLength),
		}

		switch {
		case statusCode >= 500:
			logger.Error("HTTP server error", fields...)
		case statusCode >= 400:
			logger.Warn("HTTP client error", fields...)
		default:
			logger.Info("HTTP request", fields...)
		}
	}
}
