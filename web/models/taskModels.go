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
