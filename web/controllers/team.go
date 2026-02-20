package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"backend/web/models"
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
		log.Println("CreateTeam: invalid payload:", err)
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
			log.Println("CreateTeam: project not found:", input.ProjectID)
			utils.Failed(c, http.StatusNotFound, "Project does not exist")
			return
		}
		log.Println("CreateTeam: project verification failed:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to verify project")
		return
	}

	_, err = config.DB.Exec(`
		INSERT INTO team_master (team_id, project_id, team_leader_employee_id)
		VALUES (?, ?, ?)`,
		input.TeamID, input.ProjectID, input.TeamLeaderID)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			log.Println("CreateTeam: duplicate team id:", input.TeamID)
			utils.Failed(c, http.StatusConflict, "Team ID already exists")
			return
		}
		log.Println("CreateTeam: insert failed:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to create team")
		return
	}

	utils.Success(c, gin.H{"message": "Team created successfully"})
}

func GetTeamByID(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		log.Println("GetTeamByID: missing team id")
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
			log.Println("GetTeamByID: team not found:", id)
			utils.Failed(c, http.StatusNotFound, "Team not found")
			return
		}
		log.Println("GetTeamByID: query error:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch team")
		return
	}

	utils.Success(c, team)
}

func GetTeamsByProject(c *gin.Context) {

	projectID := c.Param("project_id")
	if projectID == "" {
		log.Println("GetTeamsByProject: missing project id")
		utils.Failed(c, http.StatusBadRequest, "Project ID is required")
		return
	}

	rows, err := config.DB.Query(`
		SELECT team_id, project_id, team_leader_employee_id, created_at
		FROM team_master WHERE project_id = ?`, projectID)

	if err != nil {
		log.Println("GetTeamsByProject: query error:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch teams")
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
			&team.CreatedAt,
		); err != nil {
			log.Println("GetTeamsByProject: scan error:", err)
			utils.Failed(c, http.StatusInternalServerError, "Failed to process team data")
			return
		}
		teams = append(teams, team)
	}

	utils.Success(c, teams)
}

func UpdateTeamLeader(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		log.Println("UpdateTeamLeader: missing team id")
		utils.Failed(c, http.StatusBadRequest, "Team ID is required")
		return
	}

	var input models.UpdateTeamLeaderRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Println("UpdateTeamLeader: invalid payload:", err)
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	result, err := config.DB.Exec(`
		UPDATE team_master SET team_leader_employee_id = ?
		WHERE team_id = ?`,
		input.TeamLeaderID, id)

	if err != nil {
		log.Println("UpdateTeamLeader: update failed:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to update team leader")
		return
	}

	_, err = result.RowsAffected()
	if err != nil {
		log.Println("UpdateTeamLeader: rowsAffected error:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to verify update result")
		return
	}

	
	utils.Success(c, gin.H{"message": "Team leader updated successfully"})
}

func DeleteTeam(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		log.Println("DeleteTeam: missing team id")
		utils.Failed(c, http.StatusBadRequest, "Team ID is required")
		return
	}

	result, err := config.DB.Exec("DELETE FROM team_master WHERE team_id = ?", id)
	if err != nil {
		log.Println("DeleteTeam: delete failed:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to delete team")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("DeleteTeam: rowsAffected error:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to verify deletion")
		return
	}

	if rowsAffected == 0 {
		log.Println("DeleteTeam: team not found:", id)
		utils.Failed(c, http.StatusNotFound, "Team not found")
		return
	}

	utils.Success(c, gin.H{"message": "Team deleted successfully"})
}

func AddTeamMember(c *gin.Context) {

	teamID := c.Param("id")
	if teamID == "" {
		log.Println("AddTeamMember: missing team id")
		utils.Failed(c, http.StatusBadRequest, "Team ID is required")
		return
	}

	var input models.AddTeamMemberRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Println("AddTeamMember: invalid payload:", err)
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	_, err := config.DB.Exec(`
		INSERT INTO team_member_mapping (team_id, employee_id)
		VALUES (?, ?)`,
		teamID, input.EmployeeID)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			log.Println("AddTeamMember: duplicate entry:", teamID, input.EmployeeID)
			utils.Failed(c, http.StatusConflict, "Member already exists in team")
			return
		}
		log.Println("AddTeamMember: insert failed:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to add team member")
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
		log.Println("RemoveTeamMember: delete failed:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to remove team member")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("RemoveTeamMember: rowsAffected error:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to verify removal")
		return
	}

	if rowsAffected == 0 {
		log.Println("RemoveTeamMember: member not found in team:", teamID, empID)
		utils.Failed(c, http.StatusNotFound, "Member not found in team")
		return
	}

	utils.Success(c, gin.H{"message": "Member removed successfully"})
}

func GetTeamMembers(c *gin.Context) {

	teamID := c.Param("id")
	if teamID == "" {
		log.Println("GetTeamMembers: missing team id")
		utils.Failed(c, http.StatusBadRequest, "Team ID is required")
		return
	}

	rows, err := config.DB.Query(`
		SELECT e.employee_id, e.employee_name, e.employee_email
		FROM team_member_mapping tm
		JOIN employee_master e ON tm.employee_id = e.employee_id
		WHERE tm.team_id = ?`, teamID)

	if err != nil {
		log.Println("GetTeamMembers: query failed:", err)
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch team members")
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
			log.Println("GetTeamMembers: scan error:", err)
			utils.Failed(c, http.StatusInternalServerError, "Failed to process member data")
			return
		}
		members = append(members, member)
	}

	utils.Success(c, members)
}
