package middleware

import (
	"github.com/gin-gonic/gin"
)

// Auth возвращает middleware для проверки авторизации
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement auth logic

		// token := c.GetHeader("Authorization")
		// if token == "" {
		//     c.JSON(401, gin.H{"error": "no token"})
		//     c.Abort()
		//     return
		// }
		//
		// userID, err := validateToken(token)
		// if err != nil {
		//     c.JSON(401, gin.H{"error": "invalid token"})
		//     c.Abort()
		//     return
		// }
		//
		// c.Set("userID", userID)

		c.Next()
	}
}
