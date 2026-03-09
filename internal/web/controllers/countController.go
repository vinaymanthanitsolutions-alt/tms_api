package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/internal/web/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// all checked
func GetEmployeeCounts(c *gin.Context) {

	var managerID string

	empID, exists := c.Get("emp_id")
	if exists {
		managerID = empID.(string)
	} else {
		managerID = c.Query("manager_id")
	}

	var response models.EmployeeCountsResponse
	var err error

	query := `
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(employee_status = 'ACTIVE'), 0),
			COALESCE(SUM(employee_status = 'INACTIVE'), 0),
			COALESCE(SUM(employee_status = 'SUSPENDED'), 0)
		FROM employee_master
	`

	if managerID != "" {
		query += " WHERE manager_employee_id = ?"
		err = config.DB.QueryRow(query, managerID).
			Scan(&response.TotalEmployees, &response.Active, &response.Inactive, &response.Suspended)
	} else {
		err = config.DB.QueryRow(query).
			Scan(&response.TotalEmployees, &response.Active, &response.Inactive, &response.Suspended)
	}

	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, response)
}

func GetProjectCounts(c *gin.Context) {

	pmID := c.Query("pm_id")
	adminID := c.Query("admin_id")

	// if pmID == "" && adminID == "" {
	// 	utils.Failed(c, http.StatusBadRequest, "pm_id or admin_id required")
	// 	return
	// }

	var response models.ProjectCountsResponse
	var err error

	query := `
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN project_status = 'PLANNING' THEN 1 ELSE 0 END), 0) AS planning,
			COALESCE(SUM(CASE WHEN project_status = 'ACTIVE' THEN 1 ELSE 0 END), 0) AS active,
			COALESCE(SUM(CASE WHEN project_status = 'COMPLETED' THEN 1 ELSE 0 END), 0) AS completed,
			COALESCE(SUM(CASE WHEN project_status != 'COMPLETED' AND project_deadline < NOW() THEN 1 ELSE 0 END), 0) AS overdue
		FROM project_master
	`

	if pmID != "" {
		query += " WHERE project_manager_id = ?"
		err = config.DB.QueryRow(query, pmID).
			Scan(&response.TotalProjects, &response.Planning, &response.Active, &response.Completed, &response.Overdue)
	} else if adminID != "" {
		query += " WHERE project_created_by = ?"
		err = config.DB.QueryRow(query, adminID).
			Scan(&response.TotalProjects, &response.Planning, &response.Active, &response.Completed, &response.Overdue)
	} else {
		err = config.DB.QueryRow(query).
			Scan(&response.TotalProjects, &response.Planning, &response.Active, &response.Completed, &response.Overdue)
	}

	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, response)
}

func GetTeamCounts(c *gin.Context) {

	projectID := c.Query("project_id")
	teamLeaderID := c.Query("team_leader_id")
	pmID := c.Query("pm_id")
	adminID := c.Query("admin_id")

	var response models.TeamCountsResponse
	var err error

	switch {
	case projectID != "":
		err = config.DB.QueryRow(
			`SELECT COUNT(*) FROM team_master WHERE project_id = ?`,
			projectID,
		).Scan(&response.TotalTeams)

	case teamLeaderID != "":
		err = config.DB.QueryRow(
			`SELECT COUNT(*) FROM team_master WHERE team_leader_employee_id = ?`,
			teamLeaderID,
		).Scan(&response.TotalTeams)

	case pmID != "":
		err = config.DB.QueryRow(`
			SELECT COUNT(*)
			FROM team_master t
			JOIN project_master p ON t.project_id = p.project_id
			WHERE p.project_manager_id = ?`,
			pmID,
		).Scan(&response.TotalTeams)

	case adminID != "":
		err = config.DB.QueryRow(`
			SELECT COUNT(*)
			FROM team_master t
			JOIN project_master p ON t.project_id = p.project_id
			WHERE p.project_created_by = ?`,
			adminID,
		).Scan(&response.TotalTeams)

	default:
		utils.Failed(c, http.StatusBadRequest,
			"project_id / team_leader_id / pm_id / admin_id required")
		return
	}

	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, response)
}

func GetTaskCounts(c *gin.Context) {

	role := c.Query("role")
	employeeID := c.Query("employee_id")

	if role == "" {
		utils.Failed(c, http.StatusBadRequest, "role is required")
		return
	}

	var response models.TaskCountsResponse
	var err error

	baseQuery := `
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(task_status = 'TODO'), 0),
			COALESCE(SUM(task_status = 'IN_PROGRESS'), 0),
			COALESCE(SUM(task_status = 'COMPLETED'), 0)
		FROM task_master t
		JOIN project_master p ON t.project_id = p.project_id
		JOIN team_master tm ON t.team_id = tm.team_id
		WHERE t.deleted_at IS NULL
	`

	switch role {

	case "SUPER_ADMIN":
		err = config.DB.QueryRow(baseQuery).
			Scan(&response.TotalTasks, &response.Todo,
				&response.InProgress, &response.Completed)

	case "ADMIN":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}
		query := baseQuery + " AND p.project_created_by = ?"
		err = config.DB.QueryRow(query, employeeID).
			Scan(&response.TotalTasks, &response.Todo,
				&response.InProgress, &response.Completed)

	case "PROJECT_MANAGER":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}
		query := baseQuery + " AND p.project_manager_id = ?"
		err = config.DB.QueryRow(query, employeeID).
			Scan(&response.TotalTasks, &response.Todo,
				&response.InProgress, &response.Completed)

	case "TEAM_LEADER":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}
		query := baseQuery + " AND tm.team_leader_employee_id = ?"
		err = config.DB.QueryRow(query, employeeID).
			Scan(&response.TotalTasks, &response.Todo,
				&response.InProgress, &response.Completed)

	case "DEVELOPER", "TESTER":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}
		query := baseQuery + " AND t.assigned_to_employee_id = ?"
		err = config.DB.QueryRow(query, employeeID).
			Scan(&response.TotalTasks, &response.Todo,
				&response.InProgress, &response.Completed)

	default:
		utils.Failed(c, http.StatusBadRequest, "invalid role")
		return
	}

	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, response)
}

func GetQueryCounts(c *gin.Context) {

	role := c.Query("role")
	employeeID := c.Query("employee_id")

	if role == "" {
		utils.Failed(c, http.StatusBadRequest, "role is required")
		return
	}

	var response models.QueryCountsResponse
	var err error

	baseQuery := `
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(query_status = 'OPEN'), 0),
			COALESCE(SUM(query_status = 'IN_PROGRESS'), 0),
			COALESCE(SUM(query_status = 'RESOLVED'), 0),
			COALESCE(SUM(query_status = 'CLOSED'), 0)
		FROM query_master q
		LEFT JOIN project_master p ON q.project_id = p.project_id
		LEFT JOIN task_master t ON q.task_id = t.task_id
		LEFT JOIN sub_task_master st ON q.sub_task_id = st.sub_task_id
		WHERE q.deleted_at IS NULL
	`

	switch role {

	case "SUPER_ADMIN":
		err = config.DB.QueryRow(baseQuery).
			Scan(&response.TotalQueries, &response.Open,
				&response.InProgress, &response.Resolved, &response.Closed)

	case "ADMIN":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}
		query := baseQuery + " AND p.project_created_by = ?"
		err = config.DB.QueryRow(query, employeeID).
			Scan(&response.TotalQueries, &response.Open,
				&response.InProgress, &response.Resolved, &response.Closed)

	case "PROJECT_MANAGER":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}
		query := baseQuery + " AND p.project_manager_id = ?"
		err = config.DB.QueryRow(query, employeeID).
			Scan(&response.TotalQueries, &response.Open,
				&response.InProgress, &response.Resolved, &response.Closed)

	case "TEAM_LEADER":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}
		query := baseQuery + `
			AND (
				st.assigned_by_employee_id = ?
				OR t.created_by_employee_id = ?
			)`
		err = config.DB.QueryRow(query, employeeID, employeeID).
			Scan(&response.TotalQueries, &response.Open,
				&response.InProgress, &response.Resolved, &response.Closed)

	case "DEVELOPER", "TESTER":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}
		query := baseQuery + `
			AND (
				q.raised_by_employee_id = ?
				OR q.assigned_to_employee_id = ?
			)`
		err = config.DB.QueryRow(query, employeeID, employeeID).
			Scan(&response.TotalQueries, &response.Open,
				&response.InProgress, &response.Resolved, &response.Closed)

	default:
		utils.Failed(c, http.StatusBadRequest, "invalid role")
		return
	}

	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, response)
}

func GetEmployeeCountsByRole(c *gin.Context) {

	role := c.Query("role")
	employeeID := c.Query("employee_id")

	if role == "" {
		utils.Failed(c, http.StatusBadRequest, "role is required")
		return
	}

	var response models.EmployeeRoleCountsResponse
	var err error

	baseQuery := `
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(employee_role = 'ADMIN'), 0),
			COALESCE(SUM(employee_role = 'PROJECT_MANAGER'), 0),
			COALESCE(SUM(employee_role = 'TEAM_LEADER'), 0),
			COALESCE(SUM(employee_role = 'DEVELOPER'), 0),
			COALESCE(SUM(employee_role = 'TESTER'), 0)
		FROM employee_master
		WHERE deleted_at IS NULL
	`

	switch role {

	case "SUPER_ADMIN":
		err = config.DB.QueryRow(baseQuery).
			Scan(
				&response.TotalEmployees,
				&response.Admin,
				&response.ProjectManager,
				&response.TeamLeader,
				&response.Developer,
				&response.Tester,
			)

	case "ADMIN":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}

		query := baseQuery + " AND manager_employee_id = ?"

		err = config.DB.QueryRow(query, employeeID).
			Scan(
				&response.TotalEmployees,
				&response.Admin,
				&response.ProjectManager,
				&response.TeamLeader,
				&response.Developer,
				&response.Tester,
			)

	case "PROJECT_MANAGER":
		if employeeID == "" {
			utils.Failed(c, http.StatusBadRequest, "employee_id required")
			return
		}

		query := baseQuery + " AND pm_employee_id = ?"

		err = config.DB.QueryRow(query, employeeID).
			Scan(
				&response.TotalEmployees,
				&response.Admin,
				&response.ProjectManager,
				&response.TeamLeader,
				&response.Developer,
				&response.Tester,
			)

	default:
		utils.Failed(c, http.StatusBadRequest, "invalid role")
		return
	}

	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, response)
}
