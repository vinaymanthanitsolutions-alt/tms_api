package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)
// All working as new db
func Routes(server *gin.Engine) {
	server.POST("/signup",controllers.Add) //by sarthak and divya
	server.POST("/login",controllers.LoginUser)   
	server.POST("/forgetPassword",controllers.ForgetPassword)	
	server.POST("/verify-OTP",controllers.VerifyOtp)
	server.POST("/updatePassword",controllers.UpdatePassword)
}
