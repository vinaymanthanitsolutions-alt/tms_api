package controllers

import (
	"backend/internal/config"
	"backend/internal/utils"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func VerifyOtp(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Bind error:", err)
		utils.Failed(c, 400, "Invalid request")
		return
	}

	var dbOTP string
	var expiry time.Time

	err := config.DB.QueryRow(`
		SELECT otp, otp_expires_at
		FROM users
		WHERE email=?
	`, req.Email).Scan(&dbOTP, &expiry)

	if err != nil {

		utils.Failed(c, 401, "Invalid email or otp")
		return
	}

	if time.Now().After(expiry) {

		utils.Failed(c, 401, "OTP expired")
		return
	}

	if strings.TrimSpace(dbOTP) != req.OTP {
		utils.Failed(c, 401, "Invalid otp")
		return
	}

	_, err = config.DB.Exec(`
		UPDATE users
		SET otp=NULL, otp_expires_at=NULL
		WHERE email=?
	`, req.Email)

	if err != nil {
		utils.Failed(c, 500, "OTP not updated")
		return
	}

	utils.Success(c, gin.H{"message": "OTP verified succesfully"})
}
