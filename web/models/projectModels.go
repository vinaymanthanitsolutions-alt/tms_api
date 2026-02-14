package models

import "database/sql"



type Project struct {
	ProjectID  string `json:"project_id"`
	Name       string `json:"name"`
	Description string `json:"description"`
	CreatedBy  string `json:"created_by"`
	PMID       sql.NullString `json:"pm_id"`
	Status     string `json:"status"`
	Deadline   string `json:"deadline"`
}