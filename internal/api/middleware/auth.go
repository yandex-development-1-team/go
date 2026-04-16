package middleware

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	service "github.com/yandex-development-1-team/go/internal/service/api"
)

type Middleware struct {
	client *sqlx.DB
}

func NewMiddlewareRepository(db *sqlx.DB) *Middleware {
	return &Middleware{client: db}
}

func (m *Middleware) Auth(jwtSecret []byte) gin.HandlerFunc {
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
				return jwtSecret, nil
			},
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": models.ErrValidation})
			return
		}

		claims, ok := token.Claims.(*service.AccessClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": models.ErrValidation})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.ErrUnauthorized})
			c.Abort()
			return
		}

		if role != service.RoleAdmin {
			logger.Error("role is not admin")
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireManagersOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.ErrUnauthorized})
			c.Abort()
			return
		}

		if role == service.RoleUser {
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}

		err := validateRoleFromRequest(role)
		if err != nil {
			logger.Error("failed to validate role", zap.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *Middleware) RoleVerification(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var permissionsForRole pq.StringArray
		const getRolePermissionsQuery = `SELECT permissions FROM role_permissions WHERE role = $1`

		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": models.ErrUnauthorized})
			c.Abort()
			return
		}

		err := validateRoleFromRequest(role)
		if err != nil {
			logger.Error("failed to validate role", zap.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}

		if role == service.RoleAdmin {
			c.Next()
			return
		}

		if role == service.RoleUser {
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}

		err = m.client.QueryRowContext(ctx, getRolePermissionsQuery, role).Scan(&permissionsForRole)
		if err != nil {
			logger.Error("failed to get permissions for role from db", zap.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
			c.Abort()
			return
		}

		for _, permissionForRole := range permissionsForRole {
			if permissionForRole == permission {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": models.ErrForbidden})
		c.Abort()
	}
}

func validateRoleFromRequest(roleReq any) error {
	exist := false

	for _, role := range service.Roles {
		if role == roleReq {
			exist = true
		}
	}

	if !exist {
		return fmt.Errorf("wrong role")
	}

	return nil
}
