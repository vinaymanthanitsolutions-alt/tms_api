package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func SubTasksRoutes(r *gin.Engine) {
	r.POST("/subtasks", controllers.CreateSubTask)
	r.PUT("/subtasks/:id/status", controllers.UpdateSubTaskStatus)
	r.DELETE("/subtasks/:id", controllers.DeleteSubTask)
	r.GET("/tasks/:task_id/subtasks", controllers.GetSubTasksByTask)
	r.GET("/team/members-subtasks", controllers.GetTeamMembersWithSubTasks) //will used by divya
}
