package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)
//ALL CHECK
func QueryRoutes(r *gin.Engine) {
	query := r.Group("/query")

	query.POST("/queries", controllers.CreateQuery)
	query.GET("/queries", controllers.GetAllQueries)
	query.GET("/queries/project/:project_id", controllers.GetQueriesByProject)
	query.PUT("/queries/:id", controllers.UpdateQuery)

}
