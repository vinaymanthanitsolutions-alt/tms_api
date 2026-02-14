package controllers

import (
	"net/http"
	"backend/internal/config"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type Project struct {
	ProjectID  string `json:"project_id"`
	Name       string `json:"name"`
	Description string `json:"description"`
	CreatedBy  string `json:"created_by"`
	PMID       string `json:"pm_id"`
	Status     string `json:"status"`
	Deadline   string `json:"deadline"`
}


func CreateProject(c *gin.Context) {
	var p Project

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

	var p Project
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

func GetProjectByID(c *gin.Context) {
	id := c.Param("project_id")

	rows, err := config.DB.Query(
		"SELECT * FROM project WHERE project_id=?",
		id,
	)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Fetch failed")
		return
	}

	utils.Success(c, rows)
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