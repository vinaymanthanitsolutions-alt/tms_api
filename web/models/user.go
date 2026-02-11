package models

type UserGet struct {
	EmpCode    string `json:"emp_code" binding:"required"`
	EmpName    string `json:"emp_name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Phone      string `json:"phone"`
	Password   string `json:"password" binding:"required"`
	Department string `json:"department"`
	Role       string `json:"role" binding:"required"`
	AdminID    int64  `json:"admin_id"`
}

