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
		utils.Failed(c, http.StatusConflict, "Employee ID, email or phone already exists")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Employee created successfully",
	})
}

