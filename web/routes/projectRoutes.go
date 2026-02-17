package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func ProjectRoutes(r *gin.Engine) {
	project := r.Group("/project")

	project.POST("/", controllers.CreateProject) //by sarthak
	project.PUT("/:project_id", controllers.UpdateProject) //used by sarthak 
	project.DELETE("/:project_id", controllers.DeleteProject)// using by sarthak

	project.GET("/byPM", controllers.GetProjectsByPM) // used vansh
	project.GET("/", controllers.GetAllProjects)

	project.PUT("/assign/:project_id", controllers.AssignProjectManager) //used by sarthak
	project.GET("/admin", controllers.GetProjectsByAdmin) //using by sarthak we have to add progress section 
	project.GET("/details", controllers.GetProjectTeamDetails) // used vansh
	project.GET("/reportAsPM",controllers.GetProjectsGroupedByManager)
}