package middleware

import (
	"github.com/gin-gonic/gin"
)

// Auth returns middleware for authorization verification
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement auth logic
		c.Next()
	}
}
