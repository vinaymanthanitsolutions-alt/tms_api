package models

import "time"

type Team struct {
	TeamID       string    `json:"team_id"`
	ProjectID    string    `json:"project_id"`
	TeamLeaderID string    `json:"team_leader_id"`
	TLName       string    `json:"tl_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateTeamRequest struct {
	TeamID       string `json:"team_id" binding:"required"`
	ProjectID    string `json:"project_id" binding:"required"`
	TeamLeaderID string `json:"team_leader_id" binding:"required"`
}

type UpdateTeamLeaderRequest struct {
	TeamLeaderID string `json:"team_leader_id" binding:"required"`
}

type AddTeamMemberRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
}

type TeamMemberResponse struct {
	EmployeeID string `json:"emp_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
}
