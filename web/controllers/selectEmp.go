package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"net/http"
	"time"


	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func SelectUser(c *gin.Context) {
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
		`SELECT emp_code, emp_password FROM employee WHERE emp_id=? AND deleted_at IS NULL`,
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

	// ✅ Store OTP in DB
	_, err = config.DB.Exec(
		"UPDATE employee SET otp = ?, expire_at = ? WHERE emp_id = ?",
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

func GetUser (c *gin.Context){
	
}
