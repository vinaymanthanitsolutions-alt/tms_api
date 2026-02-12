package models

type UserGet struct {
	EmpID    string `json:"emp_id" binding:"required"`
	EmpName    string `json:"emp_name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Phone      string `json:"phone"`
	Password   string `json:"password" binding:"required"`
	Department string `json:"department"`
	Role       string `json:"role" binding:"required"`
	AdminID    string  `json:"admin_id"`
}

