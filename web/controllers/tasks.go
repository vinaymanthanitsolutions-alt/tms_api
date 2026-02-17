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

func CreateTask(c *gin.Context) {

	var input struct {
		ProjectID   string `json:"project_id" binding:"required"`
		TeamID      string `json:"team_id" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		AssignedTo  string `json:"assigned_to" binding:"required"`
		CreatedBy   string `json:"created_by" binding:"required"`
		Deadline    string `json:"deadline"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var deadline sql.NullString
	if input.Deadline != "" {

		t, err := time.Parse(time.RFC3339, input.Deadline)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid deadline format. Use 2026-01-20T00:00:00Z"})
			return
		}
		deadline = sql.NullString{String: t.Format("2006-01-02 15:04:05"), Valid: true}
	} else {
		deadline = sql.NullString{Valid: false}
	}

	query := `
		INSERT INTO tasks 
		(project_id, team_id, title, description, assigned_to, created_by, deadline)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := config.DB.Exec(
		query,
		input.ProjectID,
		input.TeamID,
		input.Title,
		input.Description,
		input.AssignedTo,
		input.CreatedBy,
		deadline,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Task created successfully"})
}

func GetTasksByProject(c *gin.Context) {

	projectID := c.Param("project_id")

	log.Println("project_id :", projectID)

	rows, err := config.DB.Query(`
		 SELECT t.id, p.name, team_id, t.title, t.status, t.assigned_to, t.deadline
		FROM tasks t
		join project p
		on t.project_id = p.project_id
		WHERE t.project_id = ?
	`, projectID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tasks []gin.H

	for rows.Next() {
		var id int
		var projectName, teamID, title, status, assignedTo string
		var deadline *string

		rows.Scan(&id, &projectName, &teamID, &title, &status, &assignedTo, &deadline)

		tasks = append(tasks, gin.H{
			"id":          id,
			"projectName": projectName,
			"teamID":      teamID,
			"title":       title,
			"status":      status,
			"assigned_to": assignedTo,
			"deadline":    deadline,
		})
	}

	c.JSON(200, tasks)
}

func GetTasksByUser(c *gin.Context) {

	empID := c.Param("emp_id")

	rows, err := config.DB.Query(`
		SELECT id, title, status, project_id
		FROM tasks
		WHERE assigned_to = ?
	`, empID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tasks []gin.H

	for rows.Next() {
		var id int
		var title, status, projectID string

		rows.Scan(&id, &title, &status, &projectID)

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

	taskID := c.Param("id")

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := config.DB.Exec(`
		UPDATE tasks
		SET status = ?
		WHERE id = ?
	`, input.Status, taskID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Task status updated"})
}

func UpdateTask(c *gin.Context) {

	taskID := c.Param("id")

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		AssignedTo  string `json:"assigned_to"`
		Deadline    string `json:"deadline"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := config.DB.Exec(`
		UPDATE tasks
		SET title = ?, description = ?, assigned_to = ?, deadline = ?, status = ?
		WHERE id = ?
	`,
		input.Title,
		input.Description,
		input.AssignedTo,
		input.Deadline,
		input.Status,
		taskID,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Task updated successfully"})
}

func DeleteTask(c *gin.Context) {

	taskID := c.Param("id")

	_, err := config.DB.Exec(`
		DELETE FROM tasks WHERE id = ?
	`, taskID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Task deleted successfully"})
}

func GetTasksWithDetails(c *gin.Context) {

	rows, err := config.DB.Query(`
		SELECT 
			t.title,
			t.project_id,
			t.team_id,
			t.status,
			t.deadline,

			tl.emp_id   AS team_leader_id,
			tl.emp_name AS team_leader_name,

			cr.emp_id   AS created_by_id,
			cr.emp_name AS created_by_name

		FROM tasks t
		JOIN team tm ON t.team_id = tm.team_id
		JOIN employee tl ON tm.team_leader_id = tl.emp_id
		JOIN employee cr ON t.created_by = cr.emp_id
	`)

	if err != nil {
		log.Println("Query error:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch tasks")
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
