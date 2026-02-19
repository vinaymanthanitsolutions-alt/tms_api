package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/web/models"
	"database/sql"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateProject(c *gin.Context) {
	var p models.Project

	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	
	var deadline sql.NullString
	if p.Deadline != "" {
		
		t, err := time.Parse(time.RFC3339, p.Deadline)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid deadline format. Use 2026-01-20T00:00:00Z"})
			return
		}
		deadline = sql.NullString{String: t.Format("2006-01-02 15:04:05"), Valid: true}
	} else {
		deadline = sql.NullString{Valid: false}
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
		deadline, 
	)

	if err != nil {
		log.Println("CreateProject error:", err)
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
	pmID := c.Query("pm_id")

	log.Printf("PM ID = [%s]\n", pmID)

	rows, err := config.DB.Query(`
		SELECT project_id, name, status, deadline
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
		var deadline sql.NullTime

		if err := rows.Scan(&id, &name, &status, &deadline); err != nil {
			utils.Failed(c, 500, "Scan error")
			return
		}

		project := map[string]interface{}{
			"project_id": id,
			"name":       name,
			"status":     status,
			"deadline":   nil,
		}

		if deadline.Valid {
			project["deadline"] = deadline.Time.Format("2006-01-02 15:04:05")
		}

		projects = append(projects, project)
	}

	utils.Success(c, projects)
}

func GetAllProjects(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "5")
	search := strings.TrimSpace(c.DefaultQuery("search", ""))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 5
	}

	offset := (page - 1) * limit

	query := `
		SELECT 
			p.project_id,
			p.name,
			p.description,
			p.created_by,
			p.pm_id,
			pm.emp_name AS pm_name,
			pm.manager_id AS pm_manager_id,
			pm_mgr.emp_name AS pm_manager_name,
			p.status,
			p.progress,
			p.deadline
		FROM project p
		LEFT JOIN employee pm ON p.pm_id = pm.emp_id
		LEFT JOIN employee pm_mgr ON pm.manager_id = pm_mgr.emp_id
	`
	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE p.name LIKE ? OR p.description LIKE ? OR p.pm_id LIKE ? OR p.created_by LIKE ?`
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern,searchPattern)
	}

	if where != "" {
		query += " " + where
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Fetch failed")
		return
	}
	defer rows.Close()

	var projects []map[string]interface{}

	for rows.Next() {
		var projectID, name, description, createdBy, pmID, pmName, status string
		var pmManagerID, pmManagerName sql.NullString
		var progress int
		var deadline sql.NullTime

		if err := rows.Scan(&projectID, &name, &description, &createdBy, &pmID, &pmName, &pmManagerID, &pmManagerName, &status, &progress, &deadline); err != nil {
			utils.Failed(c, http.StatusInternalServerError, "Scan error")
			return
		}

		project := map[string]interface{}{
			"project_id":      projectID,
			"name":            name,
			"description":     description,
			"created_by":      createdBy,
			"pm_id":           pmID,
			"pm_name":         pmName,
			"pm_manager_id":   nil,
			"pm_manager_name": nil,
			"status":          status,
			"progress":        progress,
			"deadline":        nil,
		}

		if pmManagerID.Valid {
			project["pm_manager_id"] = pmManagerID.String
		}
		if pmManagerName.Valid {
			project["pm_manager_name"] = pmManagerName.String
		}
		if deadline.Valid {
			project["deadline"] = deadline.Time.Format("2006-01-02 15:04:05")
		}

		projects = append(projects, project)
	}

	countQuery := "SELECT COUNT(*) FROM project p"
	countArgs := []interface{}{}
	if search != "" {
		countQuery += " WHERE p.name LIKE ? OR p.description LIKE ? OR p.pm_id LIKE ?"
		searchPattern := "%" + search + "%"
		countArgs = append(countArgs, searchPattern, searchPattern, searchPattern)
	}

	var total int
	err = config.DB.QueryRow(countQuery, countArgs...).Scan(&total)
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
    projectID := c.Param("project_id")
    var data struct {
        PMID *string `json:"pm_id"`
    }
    if err := c.ShouldBindJSON(&data); err != nil {
        utils.Failed(c, http.StatusBadRequest, "Invalid data")
        return
    }

    var currentPM sql.NullString
    err := config.DB.QueryRow(
        "SELECT pm_id FROM project WHERE project_id = ?",
        projectID,
    ).Scan(&currentPM)
    if err != nil {
        if err == sql.ErrNoRows {
            utils.Failed(c, http.StatusNotFound, "Project not found")
        } else {
            utils.Failed(c, http.StatusInternalServerError, "DB error")
        }
        return
    }

    if currentPM.Valid && data.PMID != nil && currentPM.String == *data.PMID {
        utils.Failed(c, http.StatusBadRequest, " This PM already choosen")
        return
    }

    _, err = config.DB.Exec(
        "UPDATE project SET pm_id=? WHERE project_id=?",
        data.PMID,
        projectID,
    )
    if err != nil {
		log.Println("Assigning Problem : ",err)
        utils.Failed(c, http.StatusInternalServerError, "Assignment failed")
        return
    }

    utils.Success(c, "PM assigned successfully")
}

func GetProjectsByAdmin(c *gin.Context) {
	adminID := strings.TrimSpace(c.Query("admin_id"))

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

	search := strings.TrimSpace(c.DefaultQuery("search", ""))

	query := `
		SELECT 
			p.project_id,
			p.name,
			p.description,
			p.pm_id,
			p.status,
			p.deadline,
			p.progress,
			e.emp_name AS admin_name
		FROM project p
		LEFT JOIN employee e ON p.created_by = e.emp_id
		WHERE p.created_by = ?
	`
	args := []interface{}{adminID}

	if search != "" {
		query += `
		 AND (
			p.name LIKE ? 
			OR p.description LIKE ? 
			OR p.pm_id = ?
		 )`
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, search)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Fetch failed")
		return
	}
	defer rows.Close()

	var projects []map[string]interface{}
	for rows.Next() {
		var projectID, name, description, pmID, status, adminName string
		var deadline sql.NullTime
		var progress int

		if err := rows.Scan(&projectID, &name, &description, &pmID, &status, &deadline,&progress, &adminName); err != nil {
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
			"progress":    progress,
			"admin_id":    adminID,
			"admin_name":  adminName,
		}

		if deadline.Valid {
			project["deadline"] = deadline.Time.Format("2006-01-02 15:04:05")
		}

		projects = append(projects, project)
	}

	var total int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM project WHERE created_by=?", adminID).Scan(&total)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Count failed")
		return
	}

	utils.Success(c, gin.H{
		"page":     page,
		"limit":    limit,
		"total":    total,
		"projects": projects,
	})
}

func GetProjectTeamDetails(c *gin.Context) {
	search := c.Query("search")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	searchLike := "%" + search + "%"

	countQuery := `
		SELECT COUNT(*)
		FROM project p
		LEFT JOIN team t ON t.project_id = p.project_id
		LEFT JOIN employee e_tl ON e_tl.emp_id = t.team_leader_id
		LEFT JOIN team_members tm ON tm.team_id = t.team_id
		LEFT JOIN employee e ON e.emp_id = tm.employee_id
		WHERE
			(? = '' OR 
			 p.project_id LIKE ? OR
			 p.name LIKE ? OR
			 e.emp_name LIKE ? OR
			 e_tl.emp_name LIKE ?
			)
	`
	var total int
	if err := config.DB.QueryRow(countQuery, search, searchLike, searchLike, searchLike, searchLike).Scan(&total); err != nil {
		utils.Failed(c, 500, "Failed to count project details")
		return
	}

	query := `
		SELECT 
			p.project_id,
			p.name AS project_name,
			t.team_id,
			e_tl.emp_name AS team_leader,
			e_tl.email AS team_leader_email,
			e.emp_name AS employee_name,
			e.email AS employee_email,
			e.role,
			e.department
		FROM project p
		LEFT JOIN team t ON t.project_id = p.project_id
		LEFT JOIN employee e_tl ON e_tl.emp_id = t.team_leader_id
		LEFT JOIN team_members tm ON tm.team_id = t.team_id
		LEFT JOIN employee e ON e.emp_id = tm.employee_id
		WHERE
			(? = '' OR 
			 p.project_id LIKE ? OR
			 p.name LIKE ? OR
			 e.emp_name LIKE ? OR
			 e_tl.emp_name LIKE ?
			)
		ORDER BY p.project_id, t.team_id, e.role
		LIMIT ? OFFSET ?
	`

	rows, err := config.DB.Query(
		query,
		search, searchLike, searchLike, searchLike, searchLike,
		limit, offset,
	)
	if err != nil {
		utils.Failed(c, 500, "Failed to fetch project details")
		return
	}
	defer rows.Close()

	var results []gin.H

	for rows.Next() {
		var projectID, projectName string
		var teamID, teamLeader, tLEmail sql.NullString
		var empName, empEmail, role, dept sql.NullString

		if err := rows.Scan(
			&projectID,
			&projectName,
			&teamID,
			&teamLeader,
			&tLEmail,
			&empName,
			&empEmail,
			&role,
			&dept,
		); err != nil {
			utils.Failed(c, 500, "Error reading data")
			return
		}

		results = append(results, gin.H{
			"project_id":        projectID,
			"project_name":      projectName,
			"team_id":           teamID.String,
			"team_leader":       teamLeader.String,
			"team_leader_email": tLEmail.String,
			"employee_name":     empName.String,
			"employee_email":    empEmail.String,
			"role":              role.String,
			"department":        dept.String,
		})
	}

	utils.Success(c, gin.H{
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": int(math.Ceil(float64(total) / float64(limit))),
		"results":     results,
	})
}

func GetProjectsGroupedByManager(c *gin.Context) {

	managerID := c.Query("emp_id")
	if managerID == "" {
		utils.Failed(c, http.StatusBadRequest, "manager_id is required")
		return
	}
	log.Println(managerID)
	query := `
		SELECT
			p.project_id,
			p.name,
			p.status,
			p.deadline,
			e.emp_id,
			e.emp_name
		FROM project p
		JOIN team t ON p.project_id = t.project_id
		JOIN employee e ON t.team_leader_id = e.emp_id
		WHERE p.pm_id = ?
	`

	rows, err := config.DB.Query(query, managerID)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	projectMap := make(map[string]*models.ProjectWithTL)

	for rows.Next() {
		var (
			projectID   string
			projectName string
			status      string
			deadline    *string
			tlID        string
			tlName      string
		)

		if err := rows.Scan(
			&projectID,
			&projectName,
			&status,
			&deadline,
			&tlID,
			&tlName,
		); err != nil {
			utils.Failed(c, http.StatusInternalServerError, err.Error())
			return
		}
		log.Println("project id", projectID)
		// IF PROJECT NOT EXISTS , CREATE IT
		if _, exists := projectMap[projectID]; !exists {
			projectMap[projectID] = &models.ProjectWithTL{
				ProjectID:   projectID,
				ProjectName: projectName,
				Status:      status,
				Deadline:    deadline,
				TeamLeaders: []models.TeamLeader{},
			}
		}

		//APPEND TEAM LEADER
		projectMap[projectID].TeamLeaders = append(
			projectMap[projectID].TeamLeaders,
			models.TeamLeader{
				TeamLeaderID:   tlID,
				TeamLeaderName: tlName,
			},
		)
	}

	// CONVERT MAP INTO SLICE
	var result []models.ProjectWithTL
	for _, project := range projectMap {
		result = append(result, *project)
	}

	utils.Success(c, result)
}
