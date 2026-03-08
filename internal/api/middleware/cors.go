package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS returns middleware for processing CORS headers
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow requests from any sources(so that the frontend doesn't have any problems)
		c.Header("Access-Control-Allow-Origin", "*")

		// Allow all basic HTTP methods
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		// Allow standard headers
		c.Header("Access-Control-Allow-Headers", `Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, 
					Authorization, accept, origin, Cache-Control, X-Requested-With`)

		// Headers that the client can read in the response
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, Date, X-Total-Count")

		// Allow the transfer of cookies and authorization data
		c.Header("Access-Control-Allow-Credentials", "true")

		// Processing the preflight request (OPTIONS)
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(204) // No Content
			return
		}

		c.Next()
	}
}
