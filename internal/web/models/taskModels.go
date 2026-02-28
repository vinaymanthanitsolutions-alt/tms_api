package models

import "time"

type TaskDetailResponse struct {
	Title            string     `json:"task_title"`
	ProjectID        string     `json:"project_id"`
	TeamID           string     `json:"team_id"`
	Status           string     `json:"status"`
	Deadline         *time.Time `json:"deadline,omitempty"`

	TeamLeaderID     string `json:"team_leader_id"`
	TeamLeaderName   string `json:"team_leader_name"`

	CreatedByID      string `json:"created_by_id"`
	CreatedByName    string `json:"created_by_name"`
}

type CreateTaskRequest struct {
	ProjectID   string `json:"project_id" binding:"required"`
	TeamID      string `json:"team_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	AssignedTo  string `json:"assigned_to" binding:"required"`
	CreatedBy   string `json:"created_by" binding:"required"`
	Deadline    string `json:"deadline"`
}

type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	AssignedTo  *string `json:"assigned_to"`
	Deadline    *string `json:"deadline"`
}