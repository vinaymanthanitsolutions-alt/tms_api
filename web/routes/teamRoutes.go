package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func TeamRouter(r *gin.Engine) {
	r.POST("/teams", controllers.CreateTeam) // will used by vansh
	r.GET("/teams/:id", controllers.GetTeamByID)
	r.GET("/projects/:project_id/teams", controllers.GetTeamsByProject)
	r.PUT("/teams/:id", controllers.UpdateTeamLeader)
	r.DELETE("/teams/:id", controllers.DeleteTeam)

	r.POST("/teams/:id/members", controllers.AddTeamMember)
	r.DELETE("/teams/:id/members/:emp_id", controllers.RemoveTeamMember)
	r.GET("/teams/:id/members", controllers.GetTeamMembers)

}
