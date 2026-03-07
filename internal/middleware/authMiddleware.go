package middleware

import (
	"net/http"
	"strings"

	"backend/internal/utils"
	"backend/internal/web/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.Abort(c, http.StatusUnauthorized, "Token required")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return utils.GetJwtSecret(), nil
		})

		if err != nil || !token.Valid {
			utils.Abort(c, http.StatusUnauthorized, "Invalid or Expired Token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.Abort(c, http.StatusUnauthorized, "Invalid Claims")
			return
		}

		empID, ok := claims["emp_id"].(string)
		if !ok {
			utils.Abort(c, http.StatusUnauthorized, "Invalid Id")
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			utils.Abort(c, http.StatusUnauthorized, "Invalid Role")
			return
		}

		c.Set("emp_id", empID)
		c.Set("role", role)

		c.Next()
	}
}

func RequireMinRole(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {

		roleInterface, exists := c.Get("role")
		if !exists {
			utils.Abort(c, http.StatusUnauthorized, "Invalid token")
			return
		}

		userRole := roleInterface.(string)

		userLevel, ok1 := models.RoleHierarchy[userRole]
		requiredLevel, ok2 := models.RoleHierarchy[minRole]

		if !ok1 || !ok2 {
			utils.Abort(c, http.StatusForbidden, "Invalid role")
			return
		}

		if userLevel > requiredLevel {
			utils.Abort(c, http.StatusForbidden, "Access denied")
			return
		}

		c.Next()
	}
}
