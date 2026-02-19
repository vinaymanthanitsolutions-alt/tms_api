package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func Router(server *gin.Engine) {
	server.GET("/emp", controllers.ShowEmployees) // used vansh, by sarthak,managerID ,managerName Added for divya(for users).
	server.GET("empAllUnderSameManager",controllers.GetEmployeesUnderSameManager) 
	server.DELETE("/emp/:emp_id",controllers.DeleteUser)  //by sarthak and divya
	server.PUT("/emp/:emp_id",controllers.UpdateProfile)  //by sarthak and divya
	server.PATCH("/empRestore/:emp_id",controllers.RestoreUser)
	server.GET("/empCounts",controllers.GetEmployeeCounts)   // used divya
	server.GET("/projectCounts",controllers.GetProjectCounts) // used vansh,by Divya
	server.GET("/teamCounts",controllers.GetTeamCounts)  // used vansh
}
