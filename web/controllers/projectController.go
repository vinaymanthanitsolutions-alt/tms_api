package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/web/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateProject(c *gin.Context) {
	var p models.Project

	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request")
		return
	}

	query := `
	INSERT INTO project
	(project_id, name, description, created_by, pm_id, deadline)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := config.DB.Exec(
		query,
		p.ProjectID,
		p.Name,
		p.Description,
		p.CreatedBy,
		p.PMID,
		p.Deadline,
	)

	if err != nil {
		utils.Failed(c, http.StatusConflict, "Project creation failed")
		return
	}

	utils.Success(c, "Project created successfully")
}

func UpdateProject(c *gin.Context) {
	id := c.Param("project_id")

	var p models.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid data")
		return
	}

	query := `
	UPDATE project
	SET name=?, description=?, status=?, deadline=?
	WHERE project_id=?
	`

	_, err := config.DB.Exec(
		query,
		p.Name,
		p.Description,
		p.Status,
		p.Deadline,
		id,
	)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Update failed")
		return
	}

	utils.Success(c, "Project updated")
}

func DeleteProject(c *gin.Context) {
	id := c.Param("project_id")

	_, err := config.DB.Exec(
		"DELETE FROM project WHERE project_id=?",
		id,
	)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Delete failed")
		return
	}

	utils.Success(c, "Project deleted")
}

func GetProjectsByPM(c *gin.Context) {
	pmID := c.Param("pm_id")

	log.Printf("PM ID = [%s]\n", pmID)

	rows, err := config.DB.Query(`
		SELECT project_id, name, status
		FROM project
		WHERE pm_id = ?
	`, pmID)

	if err != nil {
		utils.Failed(c, 500, "Failed to fetch projects")
		return
	}
	defer rows.Close()

	projects := []map[string]interface{}{}

	for rows.Next() {
		var id, name, status string

		if err := rows.Scan(&id, &name, &status); err != nil {
			utils.Failed(c, 500, "Scan error")
			return
		}

		projects = append(projects, gin.H{
			"project_id": id,
			"name":       name,
			"status":     status,
		})
	}

	utils.Success(c, projects)
}

func GetAllProjects(c *gin.Context) {
	rows, err := config.DB.Query(
		"SELECT * FROM project",
	)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Fetch failed")
		return
	}

	utils.Success(c, rows)
}

func AssignProjectManager(c *gin.Context) {
	id := c.Param("project_id")

	var data struct {
		PMID string `json:"pm_id"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid data")
		return
	}

	_, err := config.DB.Exec(
		"UPDATE project SET pm_id=? WHERE project_id=?",
		data.PMID,
		id,
	)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Assignment failed")
		return
	}

	utils.Success(c, "PM assigned")
}

func GetProjectsByAdmin(c *gin.Context) {
	adminID := c.Param("admin_id")

	rows, err := config.DB.Query(
		"SELECT * FROM project WHERE created_by=?",
		adminID,
	)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Fetch failed")
		return
	}

	utils.Success(c, rows)
}
