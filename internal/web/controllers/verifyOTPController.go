package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// all done
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
