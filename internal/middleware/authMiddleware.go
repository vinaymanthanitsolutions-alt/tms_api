package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"backend/internal/utils"
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


func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		roleInterface, exists := c.Get("role")
		if !exists {
			utils.Abort(c, http.StatusUnauthorized, "Invalid token")
			return
		}

		userRole, ok := roleInterface.(string)
		if !ok {
			utils.Abort(c, http.StatusUnauthorized, "Invalid role type")
			return
		}

		for _, r := range allowedRoles {
			if userRole == r {
				c.Next()
				return
			}
		}

		utils.Abort(c, http.StatusForbidden, "Access denied: insufficient permissions")
	}
}