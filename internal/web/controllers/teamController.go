package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/internal/web/models"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)
//ALL CHECKED
func CreateTeam(c *gin.Context) {

	var input models.CreateTeamRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var exists string
	err := config.DB.QueryRow(
		`SELECT project_id FROM project_master WHERE project_id = ?`,
		input.ProjectID,
	).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.Failed(c, http.StatusNotFound, "Project does not exist")
			return
		}
		c.Error(err)
		return
	}

	_, err = config.DB.Exec(`
		INSERT INTO team_master (team_id, project_id, team_leader_employee_id)
		VALUES (?, ?, ?)`,
		input.TeamID, input.ProjectID, input.TeamLeaderID)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			utils.Failed(c, http.StatusConflict, "Team ID already exists")
			return
		}
		c.Error(err)
		return
	}

	utils.Success(c, gin.H{"message": "Team created successfully"})
}

func GetTeamByID(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		utils.Failed(c, http.StatusBadRequest, "Team ID is required")
		return
	}

	var team models.Team

	err := config.DB.QueryRow(`
		SELECT team_id, project_id, team_leader_employee_id, created_at
		FROM team_master WHERE team_id = ?`, id).
		Scan(&team.TeamID, &team.ProjectID, &team.TeamLeaderID, &team.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.Failed(c, http.StatusNotFound, "Team not found")
			return
		}
		c.Error(err)
		return
	}

	utils.Success(c, team)
}

func GetTeamsByProject(c *gin.Context) {

	projectID := c.Param("project_id")
	if projectID == "" {
		utils.Failed(c, http.StatusBadRequest, "Project ID is required")
		return
	}

	rows, err := config.DB.Query(`
		SELECT 
			t.team_id,
			t.project_id,
			t.team_leader_employee_id,
			e.employee_name AS tl_name,
			t.created_at
		FROM team_master t
		JOIN employee_master e 
			ON t.team_leader_employee_id = e.employee_id
		WHERE t.project_id = ?
		AND t.deleted_at IS NULL
	`, projectID)

	if err != nil {
		c.Error(err)
		return
	}
	defer rows.Close()

	var teams []models.Team

	for rows.Next() {
		var team models.Team
		if err := rows.Scan(
			&team.TeamID,
			&team.ProjectID,
			&team.TeamLeaderID,
			&team.TLName,
			&team.CreatedAt,
		); err != nil {
			c.Error(err)
			return
		}
		teams = append(teams, team)
	}

	utils.Success(c, teams)
}

func UpdateTeamLeader(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		utils.Failed(c, http.StatusBadRequest, "Team ID is required")
		return
	}

	var input models.UpdateTeamLeaderRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	result, err := config.DB.Exec(`
		UPDATE team_master SET team_leader_employee_id = ?
		WHERE team_id = ?`,
		input.TeamLeaderID, id)

	if err != nil {
		c.Error(err)
		return
	}

	_, err = result.RowsAffected()
	if err != nil {
		c.Error(err)
		return
	}

	
	utils.Success(c, gin.H{"message": "Team leader updated successfully"})
}

func DeleteTeam(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		utils.Failed(c, http.StatusBadRequest, "Team ID is required")
		return
	}

	result, err := config.DB.Exec("DELETE FROM team_master WHERE team_id = ?", id)
	if err != nil {
		c.Error(err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.Error(err)
		return
	}

	if rowsAffected == 0 {
		utils.Failed(c, http.StatusNotFound, "Nothing to update")
		return
	}

	utils.Success(c, gin.H{"message": "Team deleted successfully"})
}

func AddTeamMember(c *gin.Context) {

	teamID := c.Param("id")
	if teamID == "" {
		utils.Failed(c, http.StatusBadRequest, "Team ID is required")
		return
	}

	var input models.AddTeamMemberRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	_, err := config.DB.Exec(`
		INSERT INTO team_member_mapping (team_id, employee_id)
		VALUES (?, ?)`,
		teamID, input.EmployeeID)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			utils.Failed(c, http.StatusConflict, "Member already exists in team")
			return
		}
		c.Error(err)
		return
	}

	utils.Success(c, gin.H{"message": "Member added successfully"})
}

func RemoveTeamMember(c *gin.Context) {

	teamID := c.Param("id")
	empID := c.Param("emp_id")

	if teamID == "" || empID == "" {
		log.Println("RemoveTeamMember: missing parameters")
		utils.Failed(c, http.StatusBadRequest, "Team ID and Employee ID are required")
		return
	}

	result, err := config.DB.Exec(`
		DELETE FROM team_member_mapping
		WHERE team_id = ? AND employee_id = ?`,
		teamID, empID)

	if err != nil {
		c.Error(err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.Error(err)
		return
	}

	if rowsAffected == 0 {
		utils.Failed(c, http.StatusNotFound, "Member not found in team")
		return
	}

	utils.Success(c, gin.H{"message": "Member removed successfully"})
}

func GetTeamMembers(c *gin.Context) {

	teamID := c.Param("id")
	if teamID == "" {
		utils.Failed(c, http.StatusBadRequest, "Team ID is required")
		return
	}

	rows, err := config.DB.Query(`
		SELECT e.employee_id, e.employee_name, e.employee_email
		FROM team_member_mapping tm
		JOIN employee_master e ON tm.employee_id = e.employee_id
		WHERE tm.team_id = ?`, teamID)

	if err != nil {
		c.Error(err)
		return
	}
	defer rows.Close()

	var members []models.TeamMemberResponse

	for rows.Next() {
		var member models.TeamMemberResponse
		if err := rows.Scan(
			&member.EmployeeID,
			&member.Name,
			&member.Email,
		); err != nil {
			c.Error(err)
			return
		}
		members = append(members, member)
	}

	utils.Success(c, members)
}
