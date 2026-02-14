package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func ProjectRoutes(r *gin.Engine) {
	project := r.Group("/project")

	project.POST("/", controllers.CreateProject)
	project.PUT("/:project_id", controllers.UpdateProject)
	project.DELETE("/:project_id", controllers.DeleteProject)

	project.GET("/:pm_id", controllers.GetProjectsByPM)
	project.GET("/", controllers.GetAllProjects)

	project.PUT("/assign/:project_id", controllers.AssignProjectManager)
	project.GET("/admin/:admin_id", controllers.GetProjectsByAdmin)
}