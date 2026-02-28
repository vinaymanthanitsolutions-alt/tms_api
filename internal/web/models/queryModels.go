package models

type Query struct {
	ID          int     `json:"id"`
	ProjectID   *string `json:"project_id"`
	TaskID      *int    `json:"task_id"`
	RaisedBy    string  `json:"raised_by"`
	AssignedTo  *string `json:"assigned_to"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	Status      string  `json:"status"`
}

type UpdateQuery struct {
	Status     *string `json:"status"`
	AssignedTo *string `json:"assigned_to"`
}
