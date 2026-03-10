package controllers

import (
	"backend/internal/config"
	"backend/internal/services"
	"backend/internal/utils"
	"backend/internal/web/models"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// all check
func CreateTask(c *gin.Context) {

	var input models.CreateTaskRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var deadline sql.NullString
	if input.Deadline != "" {

		t, err := time.Parse(time.RFC3339, input.Deadline)
		if err != nil {
			utils.Failed(c, http.StatusBadRequest, "Invalid deadline format. Use 2026-01-20T00:00:00Z")
			return
		}
		deadline = sql.NullString{String: t.Format("2006-01-02 15:04:05"), Valid: true}
	} else {
		deadline = sql.NullString{Valid: false}
	}

	_, err := config.DB.Exec(`
		INSERT INTO task_master
		(project_id, team_id, task_title, task_description, assigned_to_employee_id, created_by_employee_id, task_deadline)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, input.ProjectID, input.TeamID, input.Title, input.Description, input.AssignedTo, input.CreatedBy, deadline)
	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, "Task created successfully")
}

func GetTasksByProject(c *gin.Context) {

	projectID := c.Query("project_id")

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	search := c.Query("search")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := `
		SELECT 
		t.task_id,
		t.project_id,
		p.project_title AS project_name,
		t.team_id,
		t.created_by_employee_id,
		t.task_title,
		t.task_status,
		t.assigned_to_employee_id,
		tl.employee_id AS tl_id,
		tl.employee_name AS tl_name,
		tl.employee_department,
		t.task_deadline
		FROM task_master t
		JOIN project_master p 
		ON t.project_id = p.project_id
		LEFT JOIN team_master tm 
		ON t.team_id = tm.team_id
		LEFT JOIN employee_master tl 
		ON tm.team_leader_employee_id = tl.employee_id
		WHERE t.project_id = ?
	`

	args := []interface{}{projectID}

	if search != "" {
		query += `
	AND (
		t.task_title LIKE ?
		OR t.task_status LIKE ?
		OR e.employee_name LIKE ?
	)`
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		c.Error(err)
		return
	}
	defer rows.Close()

	var tasks []gin.H

	for rows.Next() {
		var id string
		var project_id, projectName, teamID, createdBy, title, status, assignedTo, teamLeaderID, department string
		var teamLeaderName sql.NullString
		var deadline sql.NullString

		err := rows.Scan(
			&id,
			&project_id,
			&projectName,
			&teamID,
			&createdBy,
			&title,
			&status,
			&assignedTo,
			&teamLeaderID,
			&teamLeaderName,
			&department,
			&deadline,
		)
		if err != nil {
			c.Error(err)
			return
		}

		tasks = append(tasks, gin.H{
			"id":             id,
			"project_id":     project_id,
			"projectName":    projectName,
			"teamID":         teamID,
			"createdBy":      createdBy,
			"title":          title,
			"status":         status,
			"assigned_to":    assignedTo,
			"teamLeaderID":   teamLeaderID,
			"teamLeaderName": teamLeaderName.String,
			"department":     department,
			"deadline":       deadline.String,
		})
	}

	utils.Success(c, gin.H{
		"page":  page,
		"limit": limit,
		"data":  tasks,
	})
}

func GetTasksByUser(c *gin.Context) {

	empID := c.Param("emp_id")

	rows, err := config.DB.Query(`
	SELECT task_id, task_title, task_status, project_id
	FROM task_master
	WHERE assigned_to_employee_id = ?
	`, empID)

	if err != nil {
		c.Error(err)
		return
	}
	defer rows.Close()

	var tasks []gin.H

	for rows.Next() {
		var id int
		var title, status, projectID string

		err = rows.Scan(&id, &title, &status, &projectID)
		if err != nil {
			c.Error(err)
			return
		}

		tasks = append(tasks, gin.H{
			"id":         id,
			"title":      title,
			"status":     status,
			"project_id": projectID,
		})
	}

	c.JSON(200, tasks)
}

func UpdateTaskStatus(c *gin.Context) {

	taskIDStr := c.Param("id")

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var input models.UpdateTaskStatusRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, err.Error())
		return
	}

	// Update task status manually
	result, err := config.DB.Exec(`
		UPDATE task_master
		SET task_status = ?
		WHERE task_id = ?
	`, input.Status, taskID)
	row, err := result.RowsAffected()

	if row == 0 {
	}
	if err != nil {
		c.Error(err)
		return
	}

	// Refresh progress chain
	projectID, err := services.RefreshTaskProgress(taskID)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}

	err = services.RefreshProjectProgress(projectID)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "Task status updated successfully")
}

func UpdateTask(c *gin.Context) {

	taskID := c.Param("id")

	var input models.UpdateTaskRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, err.Error())
		return
	}

	query := "UPDATE task_master SET "
	args := []interface{}{}

	if input.Title != nil {
		query += "task_title = ?, "
		args = append(args, *input.Title)
	}

	if input.Description != nil && *input.Description != "" {
		query += "task_description = ?, "
		args = append(args, *input.Description)
	}

	if input.AssignedTo != nil {
		query += "assigned_to_employee_id = ?, "
		args = append(args, *input.AssignedTo)
	}

	if input.Deadline != nil {
		query += "task_deadline = ?, "
		args = append(args, *input.Deadline)
	}

	if len(args) == 0 {
		utils.Failed(c, http.StatusBadRequest, "No valid fields to update")
		return
	}
	query = query[:len(query)-2] + " WHERE task_id = ?"
	args = append(args, taskID)

	_, err := config.DB.Exec(query, args...)
	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, gin.H{"message": "Task updated successfully"})
}

func DeleteTask(c *gin.Context) {

	taskID := c.Param("id")

	_, err := config.DB.Exec(`
	DELETE FROM task_master WHERE task_id = ?
`, taskID)

	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, gin.H{"message": "Task deleted successfully"})
}

func GetTasksWithDetails(c *gin.Context) {

	rows, err := config.DB.Query(`
	SELECT 
		t.task_title,
		t.project_id,
		t.team_id,
		t.task_status,
		t.task_deadline,

		tl.employee_id   AS team_leader_id,
		tl.employee_name AS team_leader_name,

		cr.employee_id   AS created_by_id,
		cr.employee_name AS created_by_name

	FROM task_master t
	JOIN team_master tm ON t.team_id = tm.team_id
	JOIN employee_master tl ON tm.team_leader_employee_id = tl.employee_id
	JOIN employee_master cr ON t.created_by_employee_id = cr.employee_id
`)

	if err != nil {
		c.Error(err)
		return
	}
	defer rows.Close()

	var tasks []models.TaskDetailResponse

	for rows.Next() {

		var task models.TaskDetailResponse
		var deadline sql.NullTime

		err := rows.Scan(
			&task.Title,
			&task.ProjectID,
			&task.TeamID,
			&task.Status,
			&deadline,
			&task.TeamLeaderID,
			&task.TeamLeaderName,
			&task.CreatedByID,
			&task.CreatedByName,
		)

		if err != nil {
			log.Println("Scan error:", err)
			continue
		}

		if deadline.Valid {
			task.Deadline = &deadline.Time
		}

		tasks = append(tasks, task)
	}

	utils.Success(c, tasks)
}
