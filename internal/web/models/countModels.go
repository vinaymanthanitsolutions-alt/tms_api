package models


type EmployeeCountsResponse struct {
	TotalEmployees int `json:"total_employees"`
	Active         int `json:"active"`
	Inactive       int `json:"inactive"`
	Suspended      int `json:"suspended"`
}

type ProjectCountsResponse struct {
	TotalProjects int `json:"total_projects"`
	Planning      int `json:"planning"`
	Active        int `json:"active"`
	Completed     int `json:"completed"`
}

type TeamCountsResponse struct {
	TotalTeams int `json:"total_teams"`
}


type TaskCountsResponse struct {
	TotalTasks int `json:"total_tasks"`
	Todo       int `json:"todo"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
}

type QueryCountsResponse struct {
	TotalQueries int `json:"total_queries"`
	Open         int `json:"open"`
	InProgress   int `json:"in_progress"`
	Resolved     int `json:"resolved"`
	Closed       int `json:"closed"`
}


type EmployeeRoleCountsResponse struct {
	TotalEmployees int `json:"total_employees"`
	Admin          int `json:"admin"`
	ProjectManager int `json:"project_manager"`
	TeamLeader     int `json:"team_leader"`
	Developer      int `json:"developer"`
	Tester         int `json:"tester"`
}