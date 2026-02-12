package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// this function verifies OTP
func VerifyOtp(c *gin.Context) {
	// taking payload from frontend
	var req struct {
		EmpID string `json:"empID"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Bind error:", err)
		utils.Failed(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// variables used to store data temporarily
	var dbOTP string
	var expiry time.Time
	var role string
	var email string

	// sql query to obtain data from table (UPDATED COLUMN NAME)
	err := config.DB.QueryRow(`
		SELECT otp, expire_at, role, email
		FROM employee
		WHERE emp_id = ?
	`, req.EmpID).Scan(&dbOTP, &expiry, &role, &email)

	if err != nil {
		log.Println("DB query error:", err)
		utils.Failed(c, http.StatusUnauthorized, "Invalid empID or otp")
		return
	}

	// check whether the time expires or not
	if time.Now().After(expiry) {
		utils.Failed(c, http.StatusUnauthorized, "OTP expired")
		return
	}

	// compare OTP
	if strings.TrimSpace(dbOTP) != strings.TrimSpace(req.OTP) {
		utils.Failed(c, http.StatusUnauthorized, "Invalid otp")
		return
	}

	// token generation
	token, err := utils.GenerateToken(req.EmpID, email)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// clear OTP after successful verification
	_, err = config.DB.Exec(`
<<<<<<< HEAD
	UPDATE employee
	SET otp=NULL, expire_at=NULL
	WHERE emp_id=?
=======
		UPDATE employee
		SET otp = NULL, expire_at = NULL
		WHERE emp_id = ?
>>>>>>> adbc3ce0adaa7acb4650904d0d2287d0afa5c8bb
	`, req.EmpID)

	if err != nil {
		log.Println("OTP update error:", err)
		utils.Failed(c, http.StatusInternalServerError, "OTP not updated")
		return
	}

	utils.Success(c, gin.H{
		"message": "OTP verified successfully",
		"token":   token,
		"role":    role,
	})
}
