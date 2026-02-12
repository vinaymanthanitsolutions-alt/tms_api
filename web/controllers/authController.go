package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func LoginUser(c *gin.Context) {
	var input struct {
		EmpId    string `json:"empID"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	var storedPassword string
	var empId string

	err := config.DB.QueryRow(
		`SELECT emp_id, emp_password FROM employee 
		 WHERE emp_id = ? AND deleted_at IS NULL`,
		input.EmpId,
	).Scan(&empId, &storedPassword)

	if err != nil {
		utils.Failed(c, http.StatusNotFound, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(storedPassword),
		[]byte(input.Password),
	); err != nil {
		utils.Failed(c, http.StatusUnauthorized, "Incorrect password")
		return
	}

	// ✅ Generate OTP
	otp, err := utils.GenerateOTP()
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to generate OTP")
		return
	}

	otpExpiry := time.Now().Add(5 * time.Minute)

	// ✅ Store OTP
	_, err = config.DB.Exec(
		`UPDATE employee SET otp = ?, expire_at = ? WHERE emp_id = ?`,
		otp,
		otpExpiry,
		empId,
	)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to save OTP")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"empID":   empId,
		"success": true,
	})
}


func ForgetPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid payload")
		return
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to generate OTP")
		return
	}

	otpExpiry := time.Now().Add(5 * time.Minute)

	result, err := config.DB.Exec(
		`UPDATE employee SET otp = ?, expire_at = ? WHERE email = ?`,
		otp,
		otpExpiry,
		input.Email,
	)

	rows, _ := result.RowsAffected()
	if err != nil || rows == 0 {
		utils.Failed(c, http.StatusNotFound, "Email is not registered")
		return
	}

	var empId string
	err = config.DB.QueryRow(
		`SELECT emp_id FROM employee WHERE email = ?`,
		input.Email,
	).Scan(&empId)

	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Emp ID not fetched")
		return
	}

	utils.Success(c, gin.H{
		"empID":   empId,
		"message": "OTP generated successfully and sent to email",
	})
}
