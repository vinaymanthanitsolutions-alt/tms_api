package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ALL CHECKED
func LoginUser(c *gin.Context) {
	var input struct {
		EmpId    string `json:"empID"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Login Bind Error: %v", err)
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
			log.Printf("Login Failed - User not found: %s", input.EmpId)
			utils.Failed(c, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		log.Printf("Database Error while fetching user %s: %v", input.EmpId, err)
		utils.LogError(c, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(storedPassword),
		[]byte(input.Password),
	); err != nil {
		log.Printf("Password mismatch for user %s", input.EmpId)
		utils.Failed(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		log.Printf("OTP Generation Error for user %s: %v", empId, err)
		utils.Failed(c, http.StatusInternalServerError, "Unable to process login")
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
		log.Printf("OTP Save Error for user %s: %v", empId, err)
		utils.LogError(c, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("OTP update failed, no rows affected for user %s", empId)
		utils.Failed(c, http.StatusInternalServerError, "Unable to process login")
		return
	}

	log.Printf("Login successful, OTP generated for user %s", empId)

	c.JSON(http.StatusOK, gin.H{
		"empID":   empId,
		"message": "OTP sent to your registered email",
		"success": true,
	})

}

func ForgetPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("ForgetPassword Bind Error: %v", err)
		utils.Failed(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if input.Email == "" {
		utils.Failed(c, http.StatusBadRequest, "Email is required")
		return
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		log.Printf("OTP Generation Error for email %s: %v", input.Email, err)
		utils.Failed(c, http.StatusInternalServerError, "Unable to process request")
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
		log.Printf("Database Error while updating OTP for email %s: %v", input.Email, err)
		utils.LogError(c, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("RowsAffected Error for email %s: %v", input.Email, err)
		utils.Failed(c, http.StatusInternalServerError, "Unable to process request")
		return
	}

	if rowsAffected == 0 {
		log.Printf("ForgetPassword attempt for unregistered email: %s", input.Email)
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
			log.Printf("EmpID fetch failed, no record found for email %s", input.Email)
			utils.Failed(c, http.StatusNotFound, "Email not registered")
			return
		}

		log.Printf("Database Error while fetching emp_id for email %s: %v", input.Email, err)
		utils.LogError(c, err)
		return
	}

	log.Printf("ForgetPassword OTP generated successfully for email %s", input.Email)

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
		log.Printf("UpdatePassword Bind Error: %v", err)
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
			log.Printf("UpdatePassword attempt for unregistered email: %s", input.Email)
			utils.Failed(c, http.StatusNotFound, "Email not registered")
			return
		}

		log.Printf("Database Error while fetching OTP for email %s: %v", input.Email, err)
		utils.LogError(c, err)
		return
	}

	if dbOTP != input.OTP {
		log.Printf("Invalid OTP attempt for email %s", input.Email)
		utils.Failed(c, http.StatusUnauthorized, "Invalid OTP")
		return
	}

	if time.Now().After(expiry) {
		log.Printf("Expired OTP attempt for email %s", input.Email)
		utils.Failed(c, http.StatusUnauthorized, "OTP expired")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(input.NewPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		log.Printf("Password Hashing Error for email %s: %v", input.Email, err)
		utils.Failed(c, http.StatusInternalServerError, "Unable to update password")
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
		log.Printf("Database Error while updating password for email %s: %v", input.Email, err)
		utils.Failed(c, http.StatusInternalServerError, "Unable to update password")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("RowsAffected Error while updating password for email %s: %v", input.Email, err)
		utils.LogError(c, err)
		return
	}

	if rowsAffected == 0 {
		log.Printf("Password update failed, no rows affected for email %s", input.Email)
		utils.Failed(c, http.StatusInternalServerError, "Unable to update password")
		return
	}

	log.Printf("Password updated successfully for email %s", input.Email)

	utils.Success(c, gin.H{
		"message": "Password updated successfully",
	})
}
