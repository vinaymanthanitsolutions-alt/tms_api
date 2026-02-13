package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func Router(server *gin.Engine) {
	server.GET("/emp/:emp_id", controllers.ShowEmployees)
	server.DELETE("/emp/:emp_id",controllers.DeleteUser)
	server.PATCH("/empRestore/:emp_id",controllers.RestoreUser)
}
