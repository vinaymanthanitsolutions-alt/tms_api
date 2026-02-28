package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/internal/web/models"
	"database/sql"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)
//ALL checked
func CreateProject(c *gin.Context) {
	var p models.CreateProjectRequest

	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var deadline sql.NullString
	if p.Deadline != "" {
		t, err := time.Parse(time.RFC3339, p.Deadline)
		if err != nil {
			utils.Failed(c, http.StatusBadRequest, "Invalid deadline format. Use 2026-01-20T00:00:00Z")
			return
		}
		deadline = sql.NullString{String: t.Format("2006-01-02 15:04:05"), Valid: true}
	} else {
		deadline = sql.NullString{Valid: false}
	}

	query := `
	INSERT INTO project_master
	(project_id, project_title, project_description, project_created_by, project_manager_id, project_deadline)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	if _, err := config.DB.Exec(
		query,
		p.ProjectID,
		p.Name,
		p.Description,
		p.CreatedBy,
		p.PMID,
		deadline,
	); err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, "Project created successfully")
}

func UpdateProject(c *gin.Context) {
	id := c.Param("project_id")
	var p models.UpdateProjectRequest

	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	query := `
	UPDATE project_master
	SET project_title=?, project_description=?, project_status=?, project_deadline=?
	WHERE project_id=?
	`

	if _, err := config.DB.Exec(query, p.Name, p.Description, p.Status, p.Deadline, id); err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, "Project updated successfully")
}

func DeleteProject(c *gin.Context) {

	id := c.Param("project_id")
	if id == "" {
		utils.Failed(c, 400, "project_id is required")
		return
	}

	result, err := config.DB.Exec(`
		UPDATE project_master
		SET deleted_at = NOW()
		WHERE project_id = ? AND deleted_at IS NULL
	`, id)

	if err != nil {
		c.Error(err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.Failed(c, 404, "Project not found or already deleted")
		return
	}

	utils.Success(c, "Project deleted successfully")
}

func GetProjectsByPM(c *gin.Context) {

	pmID := c.Query("pm_id")
	if pmID == "" {
		utils.Failed(c, 400, "pm_id is required")
		return
	}

	rows, err := config.DB.Query(`
		SELECT project_id, project_title, project_status, project_deadline
		FROM project_master
		WHERE project_manager_id = ?`, pmID)
	if err != nil {
		c.Error(err)
		return
	}
	defer rows.Close()

	var projects []models.ProjectByPMResponse

	for rows.Next() {

		var p models.ProjectByPMResponse
		var deadline sql.NullTime

		if err := rows.Scan(
			&p.ProjectID,
			&p.Name,
			&p.Status,
			&deadline,
		); err != nil {
			c.Error(err)
			return
		}

		if deadline.Valid {
			p.Deadline = &deadline.Time
		}

		projects = append(projects, p)
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
			p.project_title,
			p.project_description,
			p.project_created_by,
			p.project_manager_id,
			pm.employee_name AS pm_name,
			pm.manager_employee_id AS pm_manager_id,
			pm_mgr.employee_name AS pm_manager_name,
			p.project_status,
			p.project_progress,
			p.project_deadline
		FROM project_master p
		LEFT JOIN employee_master pm ON p.project_manager_id = pm.employee_id
		LEFT JOIN employee_master pm_mgr ON pm.manager_employee_id = pm_mgr.employee_id
	`

	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE p.project_title LIKE ? 
		          OR p.project_description LIKE ? 
		          OR p.project_manager_id LIKE ? 
		          OR p.project_created_by LIKE ?`
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if where != "" {
		query += " " + where
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		c.Error(err)
		return
	}
	defer rows.Close()

	var projects []map[string]interface{}

	for rows.Next() {
		var projectID, name, description, createdBy, pmID, pmName, status string
		var pmManagerID, pmManagerName sql.NullString
		var progress int
		var deadline sql.NullTime

		if err := rows.Scan(
			&projectID,
			&name,
			&description,
			&createdBy,
			&pmID,
			&pmName,
			&pmManagerID,
			&pmManagerName,
			&status,
			&progress,
			&deadline,
		); err != nil {
			c.Error(err) 
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

	if err := rows.Err(); err != nil {
		c.Error(err)
		return
	}

	countQuery := "SELECT COUNT(*) FROM project_master p"
	countArgs := []interface{}{}

	if search != "" {
		countQuery += ` WHERE p.project_title LIKE ? 
		                OR p.project_description LIKE ? 
		                OR p.project_manager_id LIKE ?`
		searchPattern := "%" + search + "%"
		countArgs = append(countArgs, searchPattern, searchPattern, searchPattern)
	}

	var total int
	err = config.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		c.Error(err) 
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
    var data models.AssignPMRequest
    if err := c.ShouldBindJSON(&data); err != nil {
        utils.Failed(c, http.StatusBadRequest, "Invalid data")
        return
    }

    var currentPM sql.NullString
    err := config.DB.QueryRow(
        "SELECT project_manager_id FROM project_master WHERE project_id = ?",
        projectID,
    ).Scan(&currentPM)
    if err != nil {
        if err == sql.ErrNoRows {
            utils.Failed(c, http.StatusNotFound, "Project not found")
        } else {
            c.Error(err)
        }
        return
    }

    if currentPM.Valid && data.PMID != nil && currentPM.String == *data.PMID {
        utils.Failed(c, http.StatusBadRequest, " This PM already choosen")
        return
    }

    _, err = config.DB.Exec(
        "UPDATE project_master SET project_manager_id=? WHERE project_id=?",
        data.PMID,
        projectID,
    )
    if err != nil {
        log.Println("Assigning Problem : ", err)
        c.Error(err)
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
        p.project_title,
        p.project_description,
        COALESCE(p.project_manager_id,''),
        p.project_status,
        p.project_deadline,
        p.project_progress,
        e.employee_name AS admin_name
    FROM project_master p
    LEFT JOIN employee_master e ON p.project_created_by = e.employee_id
    WHERE p.project_created_by = ?
`

	args := []interface{}{adminID}

	if search != "" {
		query += `
		 AND (
			p.project_title LIKE ? 
			OR p.project_description LIKE ? 
			OR p.project_manager_id LIKE ?
		 )`
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
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

		if err := rows.Scan(&projectID, &name, &description, &pmID, &status, &deadline, &progress, &adminName); err != nil {
			c.Error(err)
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
	err = config.DB.QueryRow(
		"SELECT COUNT(*) FROM project_master WHERE project_created_by=?",
		adminID,
	).Scan(&total)

	if err != nil {
		c.Error(err)
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
		FROM project_master p
		LEFT JOIN team_master t ON t.project_id = p.project_id
		LEFT JOIN employee_master e_tl ON e_tl.employee_id = t.team_leader_employee_id
		LEFT JOIN team_member_mapping tm ON tm.team_id = t.team_id
		LEFT JOIN employee_master e ON e.employee_id = tm.employee_id
		WHERE
			(? = '' OR 
			 p.project_id LIKE ? OR
			 p.project_title LIKE ? OR
			 e.employee_name LIKE ? OR
			 e_tl.employee_name LIKE ?
			)
	`

	var total int
	if err := config.DB.QueryRow(countQuery, search, searchLike, searchLike, searchLike, searchLike).Scan(&total); err != nil {
		c.Error(err)
		return
	}

	query := `
		SELECT 
			p.project_id,
			p.project_title AS project_name,
			t.team_id,
			e_tl.employee_name AS team_leader,
			e_tl.employee_email AS team_leader_email,
			e.employee_name AS employee_name,
			e.employee_email AS employee_email,
			e.employee_role,
			e.employee_department
		FROM project_master p
		LEFT JOIN team_master t ON t.project_id = p.project_id
		LEFT JOIN employee_master e_tl ON e_tl.employee_id = t.team_leader_employee_id
		LEFT JOIN team_member_mapping tm ON tm.team_id = t.team_id
		LEFT JOIN employee_master e ON e.employee_id = tm.employee_id
		WHERE
			(? = '' OR 
			 p.project_id LIKE ? OR
			 p.project_title LIKE ? OR
			 e.employee_name LIKE ? OR
			 e_tl.employee_name LIKE ?
			)
		ORDER BY p.project_id, t.team_id, e.employee_role
		LIMIT ? OFFSET ?
	`

	rows, err := config.DB.Query(
		query,
		search, searchLike, searchLike, searchLike, searchLike,
		limit, offset,
	)
	if err != nil {
		c.Error(err)
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
			c.Error(err)
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
			p.project_title,
			p.project_status,
			p.project_deadline,
			e.employee_id,
			e.employee_name
		FROM project_master p
		JOIN team_master t ON p.project_id = t.project_id
		JOIN employee_master e ON t.team_leader_employee_id = e.employee_id
		WHERE p.project_manager_id = ?
	`

	rows, err := config.DB.Query(query, managerID)
	if err != nil {
		c.Error(err)
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
			c.Error(err)
			return
		}


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
