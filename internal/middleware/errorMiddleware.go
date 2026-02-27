package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()
		
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	}
}
