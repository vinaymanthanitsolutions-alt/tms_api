package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)
//all check
func TaskRoutes(router *gin.Engine) {

	taskGroup := router.Group("/tasks")
	{
		taskGroup.POST("/", controllers.CreateTask)

		taskGroup.GET("/project/", controllers.GetTasksByProject) // by vansh, by divya(taskReport).
		taskGroup.GET("/user/:emp_id", controllers.GetTasksByUser)

		taskGroup.PATCH("/:id/status", controllers.UpdateTaskStatus)
		taskGroup.PUT("/:id", controllers.UpdateTask)

		taskGroup.DELETE("/:id", controllers.DeleteTask)
		taskGroup.GET("/getAll",controllers.GetTasksWithDetails)
	}
}
