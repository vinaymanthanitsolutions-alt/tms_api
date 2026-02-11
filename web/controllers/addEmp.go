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

	// 3️⃣ INSERT query (MATCHES TABLE STRUCTURE)
	query := `
		INSERT INTO employee (
			emp_code,
			emp_name,
			email,
			phone,
			emp_password,
			department,
			role,
			admin_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := config.DB.Exec(
		query,
		data.EmpCode,
		data.EmpName,
		data.Email,
		data.Phone,
		string(hashedPassword),
		data.Department,
		data.Role,
		data.AdminID,
	)
	if err != nil {
		utils.Failed(c, http.StatusConflict, "Employee code, email or phone already exists")
		return
	}

	// 4️⃣ Get inserted ID
	id, _ := result.LastInsertId()

	// 5️⃣ Success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Employee created successfully",
		"emp_id":  id,
	})
}

