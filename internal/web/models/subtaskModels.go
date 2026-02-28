package models

import "time"

type SubTask struct {
	ID     int `json:"id"`
	TaskID int `json:"task_id"`

	Title       string `json:"title"`
	Description string `json:"description"`

	Status string `json:"status"`

	AssignedTo string `json:"assigned_to"`
	AssignedBy string `json:"assigned_by"`

	Priority string `json:"priority"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type GetSubTask struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
	TaskID   int    `json:"task_id"`
}

type Member struct {
	EmpID      string       `json:"emp_id"`
	EmpName    string       `json:"emp_name"`
	Role       string       `json:"role"`
	Department string       `json:"department"`
	SubTasks   []GetSubTask `json:"sub_tasks"`
}
