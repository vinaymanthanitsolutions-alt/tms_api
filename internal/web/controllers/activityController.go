package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/internal/web/models"

	"github.com/gin-gonic/gin"
)

func GetActivityLogs(c *gin.Context) {

	rows, err := config.DB.Query(`
	SELECT
	al.log_id,
	al.employee_id,
	e.employee_name,
	al.action_type,
	al.entity_type,
	al.entity_id,
	al.description,
	al.created_at
	FROM activity_log_master al
	JOIN employee_master e
	ON al.employee_id = e.employee_id
	ORDER BY al.created_at DESC
	LIMIT 50
	`)

	if err != nil {
		c.Error(err)
		return
	}

	defer rows.Close()

	var logs []models.ActivityLog

	for rows.Next() {

		var log models.ActivityLog

		err := rows.Scan(
			&log.LogID,
			&log.EmployeeID,
			&log.EmployeeName,
			&log.ActionType,
			&log.EntityType,
			&log.EntityID,
			&log.Description,
			&log.CreatedAt,
		)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logs = append(logs, log)
	}

	utils.Success(c,gin.H{
		"message":"Activity created succesfully",
		"data": logs,
	})
}
