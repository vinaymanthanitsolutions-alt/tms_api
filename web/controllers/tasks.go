package controllers

import (
	"backend/internal/config"

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
		input.Deadline,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Task created successfully"})
}

func GetTasksByProject(c *gin.Context) {

	projectID := c.Param("project_id")

	rows, err := config.DB.Query(`
		SELECT id, title, status, assigned_to, deadline
		FROM tasks
		WHERE project_id = ?
	`, projectID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tasks []gin.H

	for rows.Next() {
		var id int
		var title, status, assignedTo string
		var deadline *string

		rows.Scan(&id, &title, &status, &assignedTo, &deadline)

		tasks = append(tasks, gin.H{
			"id": id,
			"title": title,
			"status": status,
			"assigned_to": assignedTo,
			"deadline": deadline,
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
			"id": id,
			"title": title,
			"status": status,
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
