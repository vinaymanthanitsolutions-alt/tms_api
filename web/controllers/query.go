package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/web/models"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)
//ALL CHECK
func CreateQuery(c *gin.Context) {
	var q models.Query

	if err := c.ShouldBindJSON(&q); err != nil {
		utils.Failed(c, http.StatusBadRequest, err.Error())
		return
	}

	query := `
		INSERT INTO query_master
		(project_id, task_id, raised_by_employee_id, assigned_to_employee_id,
		query_title, query_description, query_priority, query_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := config.DB.Exec(
		query,
		q.ProjectID,
		q.TaskID,
		q.RaisedBy,
		q.AssignedTo,
		q.Title,
		q.Description,
		q.Priority,
		q.Status,
	)

	if err != nil {
		log.Println("CreateQuery error:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to create query")
		return
	}

	utils.Success(c, "Query created successfully")
}

func GetAllQueries(c *gin.Context) {
	rows, err := config.DB.Query(`
	SELECT 
		query_id,
		project_id,
		task_id,
		raised_by_employee_id,
		assigned_to_employee_id,
		query_title,
		query_description,
		query_priority,
		query_status,
		created_at,
		updated_at
	FROM query_master
	WHERE deleted_at IS NULL
	`)
	if err != nil {
		utils.Failed(c, 500, "Failed to fetch queries")
		return
	}
	defer rows.Close()

	var queries []map[string]interface{}

	for rows.Next() {
		var id int
		var projectID, assignedTo sql.NullString
		var taskID sql.NullInt64
		var raisedBy, title, description, priority, status string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id, &projectID, &taskID, &raisedBy,
			&assignedTo, &title, &description,
			&priority, &status, &createdAt, &updatedAt,
		)
		if err != nil {
			utils.Failed(c, 500, "Scan error")
			return
		}

		q := map[string]interface{}{
			"id":          id,
			"raised_by":   raisedBy,
			"title":       title,
			"description": description,
			"priority":    priority,
			"status":      status,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		}

		if projectID.Valid {
			q["project_id"] = projectID.String
		}
		if taskID.Valid {
			q["task_id"] = taskID.Int64
		}
		if assignedTo.Valid {
			q["assigned_to"] = assignedTo.String
		}

		queries = append(queries, q)
	}

	utils.Success(c, queries)
}

func GetQueriesByProject(c *gin.Context) {
	projectID := c.Param("project_id")

	rows, err := config.DB.Query(`
	SELECT 
		query_id,
		query_title,
		query_status,
		query_priority
	FROM query_master
	WHERE project_id = ? AND deleted_at IS NULL
	`, projectID)

	if err != nil {
		utils.Failed(c, 500, "Failed to fetch queries")
		return
	}
	defer rows.Close()

	var list []map[string]interface{}

	for rows.Next() {
		var id int
		var title, status, priority string

		if err := rows.Scan(&id, &title, &status, &priority); err != nil {
			utils.Failed(c, 500, "Scan error")
			return
		}

		list = append(list, map[string]interface{}{
			"id":       id,
			"title":    title,
			"status":   status,
			"priority": priority,
		})
	}

	utils.Success(c, list)
}

func UpdateQuery(c *gin.Context) {
	id := c.Param("id")

	var payload struct {
		Status     *string `json:"status"`
		AssignedTo *string `json:"assigned_to"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.Failed(c, 400, err.Error())
		return
	}

	_, err := config.DB.Exec(`
	UPDATE query_master
	SET query_status = COALESCE(?, query_status),
	    assigned_to_employee_id = COALESCE(?, assigned_to_employee_id)
	WHERE query_id = ? AND deleted_at IS NULL
	`, payload.Status, payload.AssignedTo, id)

	if err != nil {
		utils.Failed(c, 500, "Update failed")
		return
	}

	utils.Success(c, "Query updated successfully")
}
