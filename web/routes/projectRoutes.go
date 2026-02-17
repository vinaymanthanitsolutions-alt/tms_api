package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func ProjectRoutes(r *gin.Engine) {
	project := r.Group("/project")

	project.POST("/", controllers.CreateProject) //by sarthak
	project.PUT("/:project_id", controllers.UpdateProject)
	project.DELETE("/:project_id", controllers.DeleteProject)

	project.GET("/byPM", controllers.GetProjectsByPM) // used vansh
	project.GET("/", controllers.GetAllProjects) // we have to add progress section also

	project.PUT("/assign/:project_id", controllers.AssignProjectManager)
	project.GET("/admin/:admin_id", controllers.GetProjectsByAdmin)
	project.GET("/details", controllers.GetProjectTeamDetails) // used vansh
	project.GET("/reportAsPM",controllers.GetProjectsGroupedByManager)
}