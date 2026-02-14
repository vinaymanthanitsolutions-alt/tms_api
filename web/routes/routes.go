package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func Router(server *gin.Engine) {
	server.GET("/emp", controllers.ShowEmployees)
	server.DELETE("/emp/:emp_id",controllers.DeleteUser)
	server.PUT("/emp/:emp_id",controllers.UpdateProfile)
	server.PATCH("/empRestore/:emp_id",controllers.RestoreUser)
	server.GET("/empCounts",controllers.GetEmployeeCounts)
	server.GET("/projectCounts",controllers.GetProjectCounts)
	server.GET("/teamCounts",controllers.GetTeamCounts)
}
