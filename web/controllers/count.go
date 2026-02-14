package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetEmployeeCounts(c *gin.Context) {

	managerID := c.Query("manager_id")

	var (
		total, active, inactive, suspended int
		err error
	)

	if managerID != "" {
		query := `
			SELECT
				COUNT(*) AS total,
				COALESCE(SUM(status = 'ACTIVE'), 0) AS active,
				COALESCE(SUM(status = 'INACTIVE'), 0) AS inactive,
				COALESCE(SUM(status = 'SUSPENDED'), 0) AS suspended
			FROM employee
			WHERE manager_id = ?
		`

		err = config.DB.QueryRow(query, managerID).
			Scan(&total, &active, &inactive, &suspended)

	} else {
		query := `
			SELECT
				COUNT(*) AS total,
				COALESCE(SUM(status = 'ACTIVE'), 0) AS active,
				COALESCE(SUM(status = 'INACTIVE'), 0) AS inactive,
				COALESCE(SUM(status = 'SUSPENDED'), 0) AS suspended
			FROM employee
		`

		err = config.DB.QueryRow(query).
			Scan(&total, &active, &inactive, &suspended)
	}

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Count fetch failed")
		return
	}

	utils.Success(c, gin.H{
		"total_employees": total,
		"active":          active,
		"inactive":        inactive,
		"suspended":       suspended,
	})
}

func GetProjectCounts(c *gin.Context) {

	pmID := c.Query("pm_id")
	adminID := c.Query("admin_id")

	if pmID == "" && adminID == "" {
		utils.Failed(c, http.StatusBadRequest, "pm_id or admin_id required")
		return
	}

	var (
		total, planning, active, completed int
		err error
	)

	if pmID != "" {
		query := `
			SELECT
				COUNT(*) AS total,
				COALESCE(SUM(status = 'PLANNING'), 0) AS planning,
				COALESCE(SUM(status = 'ACTIVE'), 0) AS active,
				COALESCE(SUM(status = 'COMPLETED'), 0) AS completed
			FROM project
			WHERE pm_id = ?
		`

		err = config.DB.QueryRow(query, pmID).
			Scan(&total, &planning, &active, &completed)
	}

	if adminID != "" {
		query := `
			SELECT
				COUNT(*) AS total,
				COALESCE(SUM(status = 'PLANNING'), 0) AS planning,
				COALESCE(SUM(status = 'ACTIVE'), 0) AS active,
				COALESCE(SUM(status = 'COMPLETED'), 0) AS completed
			FROM project
			WHERE created_by = ?
		`

		err = config.DB.QueryRow(query, adminID).
			Scan(&total, &planning, &active, &completed)
	}

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch project counts")
		return
	}

	utils.Success(c, gin.H{
		"total_projects": total,
		"planning":       planning,
		"active":         active,
		"completed":      completed,
	})
}

func GetTeamCounts(c *gin.Context) {

	projectID := c.Query("project_id")
	teamLeaderID := c.Query("team_leader_id")
	pmID := c.Query("pm_id")
	adminID := c.Query("admin_id")

	var (
		total int
		err   error
	)

	switch {
	case projectID != "":
		query := `
			SELECT COUNT(*)
			FROM team
			WHERE project_id = ?
		`
		err = config.DB.QueryRow(query, projectID).Scan(&total)

	case teamLeaderID != "":
		query := `
			SELECT COUNT(*)
			FROM team
			WHERE team_leader_id = ?
		`
		err = config.DB.QueryRow(query, teamLeaderID).Scan(&total)

	case pmID != "":
		query := `
			SELECT COUNT(*)
			FROM team t
			JOIN project p ON t.project_id = p.project_id
			WHERE p.pm_id = ?
		`
		err = config.DB.QueryRow(query, pmID).Scan(&total)

	case adminID != "":
		query := `
			SELECT COUNT(*)
			FROM team t
			JOIN project p ON t.project_id = p.project_id
			WHERE p.created_by = ?
		`
		err = config.DB.QueryRow(query, adminID).Scan(&total)

	default:
		utils.Failed(c, http.StatusBadRequest,
			"project_id / team_leader_id / pm_id / admin_id required")
		return
	}

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch team count")
		return
	}

	utils.Success(c, gin.H{
		"total_teams": total,
	})
}
