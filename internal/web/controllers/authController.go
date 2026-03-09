package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"database/sql"
	"net/http"
	"strings"
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
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if input.EmpId == "" || input.Password == "" {
		utils.Failed(c, http.StatusBadRequest, "empID and password are required")
		return
	}

	var storedPassword string
	var empId string

	err := config.DB.QueryRow(
		`SELECT employee_id, employee_password 
		 FROM employee_master 
		 WHERE employee_id = ? AND deleted_at IS NULL`,
		input.EmpId,
	).Scan(&empId, &storedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.Failed(c, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		c.Error(err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(input.Password)); err != nil {
		utils.Failed(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		c.Error(err)
		return
	}

	otpExpiry := time.Now().Add(5 * time.Minute)

	result, err := config.DB.Exec(
		`UPDATE employee_master 
		 SET employee_otp = ?, otp_expire_at = ? 
		 WHERE employee_id = ?`,
		otp,
		otpExpiry,
		empId,
	)
	if err != nil {
		c.Error(err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.Failed(c, http.StatusInternalServerError, "Unable to process login")
		return
	}

	utils.Success(c, gin.H{
		"empID":   empId,
		"message": "OTP sent to your registered email",
	})
}

func ForgetPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if input.Email == "" {
		utils.Failed(c, http.StatusBadRequest, "Email is required")
		return
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		c.Error(err)
		return
	}

	otpExpiry := time.Now().Add(5 * time.Minute)

	result, err := config.DB.Exec(
		`UPDATE employee_master 
		 SET employee_otp = ?, otp_expire_at = ? 
		 WHERE employee_email = ? AND deleted_at IS NULL`,
		otp,
		otpExpiry,
		input.Email,
	)
	if err != nil {
		c.Error(err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.Error(err)
		return
	}

	if rowsAffected == 0 {
		utils.Failed(c, http.StatusNotFound, "Email not registered")
		return
	}

	var empId string
	err = config.DB.QueryRow(
		`SELECT employee_id 
		 FROM employee_master 
		 WHERE employee_email = ? AND deleted_at IS NULL`,
		input.Email,
	).Scan(&empId)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.Failed(c, http.StatusNotFound, "Email not registered")
			return
		}
		c.Error(err)
		return
	}

	utils.Success(c, gin.H{
		"empID":   empId,
		"message": "OTP sent to your registered email",
	})
}

func UpdatePassword(c *gin.Context) {
	var input struct {
		Email       string `json:"email"`
		OTP         string `json:"otp"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if input.Email == "" || input.OTP == "" || input.NewPassword == "" {
		utils.Failed(c, http.StatusBadRequest, "Email, OTP and new password are required")
		return
	}

	var dbOTP string
	var expiry time.Time

	err := config.DB.QueryRow(
		`SELECT employee_otp, otp_expire_at 
		 FROM employee_master 
		 WHERE employee_email = ? AND deleted_at IS NULL`,
		input.Email,
	).Scan(&dbOTP, &expiry)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.Failed(c, http.StatusNotFound, "Email not registered")
			return
		}
		c.Error(err)
		return
	}

	if dbOTP != input.OTP {
		utils.Failed(c, http.StatusUnauthorized, "Invalid OTP")
		return
	}

	if time.Now().After(expiry) {
		utils.Failed(c, http.StatusUnauthorized, "OTP expired")
		config.DB.Exec(`
			UPDATE employee_master
			SET employee_otp = NULL, otp_expire_at = NULL
			WHERE employee_email = ?
		`, input.Email)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := config.DB.Exec(
		`UPDATE employee_master 
		 SET employee_password = ?, 
		     employee_otp = NULL, 
		     otp_expire_at = NULL
		 WHERE employee_email = ? AND deleted_at IS NULL`,
		string(hashedPassword),
		input.Email,
	)
	if err != nil {
		c.Error(err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.Error(err)
		return
	}

	if rowsAffected == 0 {
		utils.Failed(c, http.StatusInternalServerError, "Unable to update password")
		return
	}

	utils.Success(c, gin.H{
		"message": "Password updated successfully",
	})
}

func VerifyOtp(c *gin.Context) {
	var req struct {
		EmpID string `json:"empID"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Failed(c, http.StatusBadRequest, "Invalid request")
		return
	}

	var dbOTP string
	var expiry time.Time
	var role string
	var email string

	err := config.DB.QueryRow(`
		SELECT employee_otp, otp_expire_at, employee_role, employee_email
		FROM employee_master
		WHERE employee_id = ? AND deleted_at IS NULL
	`, req.EmpID).Scan(&dbOTP, &expiry, &role, &email)

	if err != nil {
		c.Error(err)
		return
	}

	if time.Now().After(expiry) {
		utils.Failed(c, http.StatusUnauthorized, "OTP expired")
		config.DB.Exec(`
			UPDATE employee_master
			SET employee_otp = NULL, otp_expire_at = NULL
			WHERE employee_id = ?
		`, req.EmpID)
		return
	}

	if strings.TrimSpace(dbOTP) != strings.TrimSpace(req.OTP) {
		utils.Failed(c, http.StatusUnauthorized, "Invalid otp")
		return
	}

	token, err := utils.GenerateToken(req.EmpID, role)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	_, err = config.DB.Exec(`
		UPDATE employee_master
		SET employee_otp = NULL, otp_expire_at = NULL
		WHERE employee_id = ?
	`, req.EmpID)

	if err != nil {
		c.Error(err)
		return
	}

	utils.Success(c, gin.H{
		"message": "OTP verified successfully",
		"token":   token,
		"role":    role,
	})
}
