package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/internal/web/models"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
		query_status
	FROM query_master
	WHERE deleted_at IS NULL
	`)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch queries")
		return
	}
	defer rows.Close()

	var queries []models.Query

	for rows.Next() {
		var q models.Query
		var projectID, assignedTo sql.NullString
		var taskID sql.NullInt64

		err := rows.Scan(
			&q.ID, &projectID, &taskID,
			&q.RaisedBy, &assignedTo,
			&q.Title, &q.Description,
			&q.Priority, &q.Status,
		)
		if err != nil {
			utils.Failed(c, http.StatusInternalServerError, "Error scanning query row")
			return
		}

		if projectID.Valid {
			q.ProjectID = &projectID.String
		}
		if taskID.Valid {
			t := int(taskID.Int64)
			q.TaskID = &t
		}
		if assignedTo.Valid {
			q.AssignedTo = &assignedTo.String
		}

		queries = append(queries, q)
	}

	utils.Success(c, queries)
}

func GetQueriesByProject(c *gin.Context) {
	projectIDParam := c.Param("project_id")

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
		query_status
	FROM query_master
	WHERE project_id = ? AND deleted_at IS NULL
	`, projectIDParam)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch queries")
		return
	}
	defer rows.Close()

	var queries []models.Query

	for rows.Next() {
		var q models.Query
		var projectID, assignedTo, raisedBy sql.NullString
		var taskID sql.NullInt64
		var description sql.NullString 

		err := rows.Scan(
			&q.ID,         
			&projectID,   
			&taskID,      
			&raisedBy,    
			&assignedTo,   
			&q.Title,     
			&description,  
			&q.Priority,   
			&q.Status,     
		)
		if err != nil {
			utils.Failed(c, http.StatusInternalServerError, "Error scanning query row")
			return
		}

		if projectID.Valid {
			q.ProjectID = &projectID.String
		}
		if taskID.Valid {
			t := int(taskID.Int64)
			q.TaskID = &t
		}
		if assignedTo.Valid {
			q.AssignedTo = &assignedTo.String
		}
		if raisedBy.Valid {
			q.RaisedBy = raisedBy.String
		}
		if description.Valid {
			q.Description = description.String
		}

		queries = append(queries, q)
	}

	utils.Success(c, queries)
}

func UpdateQuery(c *gin.Context) {
	id := c.Param("id")

	
	var payload models.UpdateQuery

	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.Failed(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err := config.DB.Exec(`
	UPDATE query_master
	SET query_status = COALESCE(?, query_status),
	    assigned_to_employee_id = COALESCE(?, assigned_to_employee_id)
	WHERE query_id = ? AND deleted_at IS NULL
	`, payload.Status, payload.AssignedTo, id)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to update query")
		return
	}

	utils.Success(c, "Query updated successfully")
}
