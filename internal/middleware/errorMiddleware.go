package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()

		if c.Writer.Written() {
			return
		}

		if len(c.Errors) > 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   c.Errors.String(),
			})
		}
	}
}
