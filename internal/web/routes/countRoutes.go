package routes

import (
	"backend/internal/web/controllers"

	"github.com/gin-gonic/gin"
)

func CountRouter(r *gin.Engine) {
	r.GET("/empCounts", controllers.GetEmployeeCounts)    // used divya
	r.GET("/projectCounts", controllers.GetProjectCounts) // used vansh,by Divya
	r.GET("/teamCounts", controllers.GetTeamCounts)   // used vansh
	r.GET("/taskCounts", controllers.GetTaskCounts)       //used sarthak
	r.GET("/queryCount",controllers.GetAllQueries)  //using by sarthak
	r.GET("/empRoleCount",controllers.GetEmployeeCountsByRole)   //used by divya
}
