package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/web/models"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

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

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "5")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 5
	}

	offset := (page - 1) * limit

	rows, err := config.DB.Query(
		"SELECT project_id, name, description, created_by, pm_id, status, deadline FROM project LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Fetch failed")
		return
	}
	defer rows.Close()

	var projects []map[string]interface{}

	for rows.Next() {
		var projectID, name, description, createdBy, pmID, status string
		var deadline sql.NullTime

		if err := rows.Scan(&projectID, &name, &description, &createdBy, &pmID, &status, &deadline); err != nil {
			utils.Failed(c, http.StatusInternalServerError, "Scan error")
			return
		}

		project := map[string]interface{}{
			"project_id":  projectID,
			"name":        name,
			"description": description,
			"created_by":  createdBy,
			"pm_id":       pmID,
			"status":      status,
			"deadline":    nil,
		}

		if deadline.Valid {
			project["deadline"] = deadline.Time.Format("2006-01-02 15:04:05")
		}

		projects = append(projects, project)
	}

	var total int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM project").Scan(&total)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Count failed")
		return
	}

	response := map[string]interface{}{
		"page":     page,
		"limit":    limit,
		"total":    total,
		"projects": projects,
	}

	utils.Success(c, response)
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
	adminID := strings.TrimSpace(c.Param("admin_id"))

	rows, err := config.DB.Query(
		"SELECT project_id, name, description, pm_id, status, deadline FROM project WHERE created_by=?",
		adminID,
	)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Fetch failed")
		return
	}
	defer rows.Close()

	var projects []map[string]interface{}

	for rows.Next() {
		var projectID, name, description, pmID, status string
		var deadline sql.NullTime

		if err := rows.Scan(&projectID, &name, &description, &pmID, &status, &deadline); err != nil {
			utils.Failed(c, http.StatusInternalServerError, "Scan error")
			return
		}

		project := map[string]interface{}{
			"project_id":  projectID,
			"name":        name,
			"description": description,
			"pm_id":       pmID,
			"status":      status,
			"deadline":    nil,
		}

		if deadline.Valid {
			project["deadline"] = deadline.Time.Format("2006-01-02 15:04:05")
		}

		projects = append(projects, project)
	}

	utils.Success(c, projects)
}
