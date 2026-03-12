package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yandex-development-1-team/go/internal/models"
	service "github.com/yandex-development-1-team/go/internal/service/api"
	"net/http"
	"strings"
)

// Auth returns middleware for authorization verification
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": models.ErrUnauthorized,
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": models.ErrUnauthorized,
			})
			return
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(
			tokenString,
			&service.AccessClaims{},
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				//todo тут нужен правильный секретный ключ, который берется из конфигов
				//для тестов через postmen использовала секретный ключ "monday" и payload{
				//  "user_id": 12345,
				//  "role": "admin",
				//  "exp": 1773491431, - нужно заменить на новую дату, чтобы токен был валиден
				//  "iat": 1773405031
				//}
				return []byte("monday"), nil
			},
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": models.ErrValidation})
			return
		}

		claims, ok := token.Claims.(*service.AccessClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": models.ErrValidation})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireAdmin returns middleware for verification role admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.ErrUnauthorized})
			c.Abort()
			return
		}

		if role != service.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}

		c.Next()
	}
}
