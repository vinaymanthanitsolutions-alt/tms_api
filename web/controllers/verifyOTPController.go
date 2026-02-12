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

//this function verify OTP
func VerifyOtp(c *gin.Context) {
	// taking payload from  frontend 
	var req struct {
		EmpID string `json:"empID"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Bind error:", err)
		utils.Failed(c, 400, "Invalid request")
		return
	}
// variables used to store data temporary
	var dbOTP string
	var expiry time.Time
	var role string
	var email string

// sql query to obtain data from table
	err := config.DB.QueryRow(`
		SELECT otp, expire_at,role,email
		FROM employee
		WHERE emp_code=?
	`, req.EmpID).Scan(&dbOTP, &expiry, &role, &email)


	if err != nil {
		log.Println("DB query error:", err)
		log.Println("DB OTP error:", dbOTP)
		
		utils.Failed(c, 401, "Invalid empID or otp")
		return
	}

	// check whether the time expires or not
	if time.Now().After(expiry) {

		utils.Failed(c, 401, "OTP expired")
		return
	}

	if strings.TrimSpace(dbOTP) != strings.TrimSpace(req.OTP) {
	utils.Failed(c, 401, "Invalid otp")
	return
}
	

	// token generation
	token, err := utils.GenerateToken(req.EmpID, email)
	if err != nil {
		utils.Failed(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}
	
	// this make otp column empty in backend 
	_, err = config.DB.Exec(`
	UPDATE employee
	SET otp=NULL, expire_at=NULL
	WHERE emp_code=?
	`, req.EmpID)
	
	if err != nil {
	log.Println("OTP update error:", err)
	utils.Failed(c, 500, "OTP not updated")
	return
}

	utils.Success(c, gin.H{
		"message": "OTP verified successfully",
		"token": token,
		"role":  role,
	})
}
