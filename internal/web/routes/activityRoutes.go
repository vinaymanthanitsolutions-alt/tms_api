package routes

import (
	"backend/internal/web/controllers"

	"github.com/gin-gonic/gin"
)

func ActivityRoute(r *gin.Engine) {
	r.GET("/activity", controllers.GetActivityLogs)
}
