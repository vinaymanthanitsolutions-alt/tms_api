package services

import "github.com/gin-gonic/gin"

func LogCreate(c *gin.Context, entityType string, entityID string, newData interface{}) {

	empID := c.GetString("emp_id")

	LogActivity(
		empID,
		"CREATE",
		entityType,
		entityID,
		nil,
		newData,
		entityType+" created",
	)
}

func LogUpdate(c *gin.Context, entityType string, entityID string, oldData interface{}, newData interface{}) {

	empID := c.GetString("emp_id")

	LogActivity(
		empID,
		"UPDATE",
		entityType,
		entityID,
		oldData,
		newData,
		entityType+" updated",
	)
}

func LogDelete(c *gin.Context, entityType string, entityID string, oldData interface{}) {

	empID := c.GetString("emp_id")

	LogActivity(
		empID,
		"DELETE",
		entityType,
		entityID,
		oldData,
		nil,
		entityType+" deleted",
	)
}