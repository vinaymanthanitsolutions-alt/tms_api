package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func Router(server *gin.Engine) {
	server.GET("/emp", controllers.ShowEmployees) // used vansh, by sarthak
	server.GET("empAllUnderSameManager",controllers.GetEmployeesUnderSameManager) 
	server.DELETE("/emp/:emp_id",controllers.DeleteUser)  //by sarthak
	server.PUT("/emp/:emp_id",controllers.UpdateProfile)  //by sarthak
	server.PATCH("/empRestore/:emp_id",controllers.RestoreUser)
	server.GET("/empCounts",controllers.GetEmployeeCounts) 
	server.GET("/projectCounts",controllers.GetProjectCounts) // used vansh
	server.GET("/teamCounts",controllers.GetTeamCounts)  // used vansh
}
