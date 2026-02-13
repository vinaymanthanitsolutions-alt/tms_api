package controllers

import (
	"net/http"

	"backend/internal/config"
	"backend/internal/utils"
	"backend/web/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Add(c *gin.Context) {
	var data models.UserGet

	// 1️⃣ Bind JSON
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// 2️⃣ Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(data.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Error generating password")
		return
	}

	// 3️⃣ Insert employee
	query := `
		INSERT INTO employee (
			emp_id,
			emp_name,
			email,
			phone,
			emp_password,
			department,
			role,
			manager_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = config.DB.Exec(
		query,
		data.EmpID,
		data.EmpName,
		data.Email,
		data.Phone,
		string(hashedPassword),
		data.Department,
		data.Role,
		data.ManagerID, // ✅ string
	)
	if err != nil {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
	})
	return
}

	c.JSON(http.StatusOK, gin.H{
		"message": "Employee created successfully",
	})
}



func ShowEmployees(c *gin.Context) {

	managerID := c.Param("emp_id")
	if managerID == "" {
	utils.Failed(c, http.StatusBadRequest, "emp_id is required")
	return
	}

	// managerID, exists := c.Get("emp_id")
	// if !exists {
	// 	utils.Failed(c, http.StatusUnauthorized, "Unauthorized")
	// 	return
	// }

	var employees []models.UserShow

	rows, err := config.DB.Query(`
		SELECT 
			emp_id,
			emp_name,
			email,
			phone,
			department,
			role,
			status
		FROM employee
		WHERE manager_id = ? and deleted_at IS NULL
	`,managerID)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to fetch employees")
		return
	}
	defer rows.Close()

	

	for rows.Next() {
		var emp models.UserShow

		err := rows.Scan(
			&emp.EmpID,
			&emp.EmpName,
			&emp.Email,
			&emp.Phone,
			&emp.Department,
			&emp.Role,
			&emp.Status,
			// &emp.ManagerID,
		)
		if err != nil {
			utils.Failed(c, http.StatusInternalServerError, "Error scanning employees")
			return
		}

		employees = append(employees, emp)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"employees": employees,
	})
}