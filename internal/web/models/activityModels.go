package models

import "time"

type ActivityLog struct {
	LogID        int       `json:"log_id"`
	EmployeeID   string    `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	ActionType   string    `json:"action_type"`
	EntityType   string    `json:"entity_type"`
	EntityID     string    `json:"entity_id"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}