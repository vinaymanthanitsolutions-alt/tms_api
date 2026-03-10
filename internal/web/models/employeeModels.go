package models

type UserGet struct {
	EmpID      string `json:"emp_id" binding:"required"`
	EmpName    string `json:"emp_name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Phone      string `json:"phone"`
	Password   string `json:"password" binding:"required"`
	Department string `json:"department"`
	Role       string `json:"role" binding:"required"`
	ManagerID  string `json:"manager_id"`
}

type UserShow struct {
	EmpID       string  `json:"emp_id" binding:"required"`
	EmpName     string  `json:"emp_name" binding:"required"`
	Email       string  `json:"email" binding:"required"`
	Phone       string  `json:"phone"`
	Password    string  `json:"password" binding:"required"`
	Department  string  `json:"department"`
	Role        string  `json:"role" binding:"required"`
	Status      string  `json:"status"`
	ManagerID   *string `json:"manager_id,omitempty"` // OMITEMPTY SKIPS FEILD LIKE WHEN MANAGER IS NULL (SUPER_ADMIN)
	ManagerName *string `json:"manager_name,omitempty"`
}

type UserUpdate struct {
	EmpName    string `json:"emp_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	Department string `json:"department"`
	Role       string `json:"role"`
	ManagerID  string `json:"manager_id"`
	Status     string `json:"status"`
}

type TeamLeader struct {
	TeamLeaderID   string `json:"team_leader_id"`
	TeamLeaderName string `json:"team_leader_name"`
}

type ProjectWithTL struct {
	ProjectID   string       `json:"project_id"`
	ProjectName string       `json:"project_name"`
	Status      string       `json:"status"`
	Deadline    *string      `json:"deadline"`
	TeamLeaders []TeamLeader `json:"team_leaders"`
}

type EmployeeUnderManager struct {
	EmpID   string `json:"emp_id"`
	EmpName string `json:"emp_name"`
	Role    string `json:"role"`
}

type LatestEmployee struct {
	EmpID string `json:"emp_id"`
	EmpName string `json:"emp_name"`
	EmpRole string `json:"emp_role"`
	EmpDepartment string `json:"emp_department"`
	CreatedAt *string `json:"created_at"`
	EmpStatus string `json:"emp_status"`
}
