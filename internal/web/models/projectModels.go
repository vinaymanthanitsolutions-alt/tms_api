package models

import "time"

type Project struct {
	ProjectID   string
	Name        string
	Description string
	CreatedBy   string
	PMID        *string
	Status      string
	Deadline    *time.Time
}

type CreateProjectRequest struct {
	ProjectID   string  `json:"project_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	CreatedBy   string  `json:"created_by" binding:"required"`
	PMID        *string `json:"pm_id"`
	Deadline    string  `json:"deadline"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Deadline    *string `json:"deadline"`
}

type AssignPMRequest struct {
	PMID *string `json:"pm_id"`
}

type ProjectListResponse struct {
	ProjectID   string     `json:"project_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedBy   string     `json:"created_by"`
	PMID        string     `json:"pm_id"`
	PMName      string     `json:"pm_name"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	Deadline    *time.Time `json:"deadline,omitempty"`
}

type ProjectByPMResponse struct {
	ProjectID string     `json:"project_id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	Deadline  *time.Time `json:"deadline,omitempty"`
}

type WeeklyProgress struct {
	Week  int `json:"week"`
	Count int `json:"completed_subtasks"`
}

type ProjectWeekly struct {
	ID       string           `json:"project_id"`
	Title    string           `json:"project_title"`
	Deadline time.Time        `json:"project_deadline"`
	Progress []WeeklyProgress `json:"weekly_progress"`
}

type MonthlyStats struct {
	Month         string `json:"month"`
	TotalProjects int    `json:"total_projects"`
	TotalProgress int    `json:"total_progress"`
}
