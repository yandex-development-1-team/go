package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/config"
)

// CORS returns middleware that sets CORS headers from cfg.
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.AllowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", cfg.AllowOrigin)
		}
		if cfg.AllowMethods != "" {
			c.Header("Access-Control-Allow-Methods", cfg.AllowMethods)
		}
		if cfg.AllowHeaders != "" {
			c.Header("Access-Control-Allow-Headers", cfg.AllowHeaders)
		}
		if cfg.ExposeHeaders != "" {
			c.Header("Access-Control-Expose-Headers", cfg.ExposeHeaders)
		}
		if cfg.AllowCredentials != "" {
			c.Header("Access-Control-Allow-Credentials", cfg.AllowCredentials)
		}

		if c.Request.Method == "OPTIONS" {
			if cfg.MaxAge != "" {
				c.Header("Access-Control-Max-Age", cfg.MaxAge)
			}
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
