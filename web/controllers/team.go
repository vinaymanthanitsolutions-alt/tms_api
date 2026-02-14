package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTeam(c *gin.Context) {
	var input struct {
		TeamID       string `json:"team_id"`
		ProjectID    string `json:"project_id"`
		TeamLeaderID string `json:"team_leader_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	var exists string
	err := config.DB.QueryRow("SELECT project_id FROM project WHERE project_id = ?", input.ProjectID).Scan(&exists)
	if err != nil {
		utils.Failed(c, http.StatusBadRequest, "Project not found")
		return
	}

	_, err = config.DB.Exec(`
		INSERT INTO team (team_id, project_id, team_leader_id)
		VALUES (?, ?, ?)`,
		input.TeamID, input.ProjectID, input.TeamLeaderID)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "Team created successfully")
}

func GetTeamByID(c *gin.Context) {
	id := c.Param("id")

	var team struct {
		TeamID       string `json:"team_id"`
		ProjectID    string `json:"project_id"`
		TeamLeaderID string `json:"team_leader_id"`
		CreatedAt    string `json:"created_at"`
	}

	err := config.DB.QueryRow(`
		SELECT team_id, project_id, team_leader_id, created_at
		FROM team WHERE team_id = ?`, id).
		Scan(&team.TeamID, &team.ProjectID, &team.TeamLeaderID, &team.CreatedAt)

	if err != nil {
		utils.Failed(c, http.StatusNotFound, "Team not found")
		return
	}

	utils.Success(c, team)
}

func GetTeamsByProject(c *gin.Context) {
	projectID := c.Param("project_id")

	rows, err := config.DB.Query(`
		SELECT team_id, team_leader_id, created_at
		FROM team WHERE project_id = ?`, projectID)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var teams []map[string]interface{}

	for rows.Next() {
		var teamID, leaderID, createdAt string
		rows.Scan(&teamID, &leaderID, &createdAt)

		teams = append(teams, map[string]interface{}{
			"team_id":        teamID,
			"team_leader_id": leaderID,
			"created_at":     createdAt,
		})
	}

	utils.Success(c, teams)
}

func UpdateTeamLeader(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		TeamLeaderID string `json:"team_leader_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	_, err := config.DB.Exec(`
		UPDATE team SET team_leader_id = ?
		WHERE team_id = ?`,
		input.TeamLeaderID, id)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "Team leader updated")
}

func DeleteTeam(c *gin.Context) {
	id := c.Param("id")

	_, err := config.DB.Exec("DELETE FROM team WHERE team_id = ?", id)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "Team deleted successfully")
}

func AddTeamMember(c *gin.Context) {
	teamID := c.Param("id")

	var input struct {
		EmployeeID string `json:"employee_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	_, err := config.DB.Exec(`
		INSERT INTO team_members (team_id, employee_id)
		VALUES (?, ?)`,
		teamID, input.EmployeeID)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "Member added successfully")
}

func RemoveTeamMember(c *gin.Context) {
	teamID := c.Param("id")
	empID := c.Param("emp_id")

	_, err := config.DB.Exec(`
		DELETE FROM team_members
		WHERE team_id = ? AND employee_id = ?`,
		teamID, empID)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, "Member removed successfully")
}

func GetTeamMembers(c *gin.Context) {
	teamID := c.Param("id")

	rows, err := config.DB.Query(`
		SELECT e.emp_id, e.emp_name, e.email
		FROM team_members tm
		JOIN employee e ON tm.employee_id = e.emp_id
		WHERE tm.team_id = ?`, teamID)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var members []map[string]interface{}

	for rows.Next() {
		var id, name, email string
		rows.Scan(&id, &name, &email)

		members = append(members, map[string]interface{}{
			"emp_id": id,
			"name":   name,
			"email":  email,
		})
	}

	utils.Success(c, members)
}
