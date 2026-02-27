package utils

import (
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

func LogError(err error) {
	if err == nil {
		return
	}

	_, file, line, _ := runtime.Caller(3)
	parts := strings.Split(file, "/")
	shortFile := parts[len(parts)-1]

	println("ERROR:", err.Error(), "| File:", shortFile, "| Line:", line)
}
