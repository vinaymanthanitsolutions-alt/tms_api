package utils

import (
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    data,
	})
}

func Failed(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"message": message,
	})
}

func Abort(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"message": message,
	})
}

func LogError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	_, file, line, _ := runtime.Caller(1)
	parts := strings.Split(file, "/")
	shortFile := parts[len(parts)-1]

	println("ERROR:", err.Error(), "| File:", shortFile, "| Line:", line)

	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error":   "Internal server error",
	})
}
