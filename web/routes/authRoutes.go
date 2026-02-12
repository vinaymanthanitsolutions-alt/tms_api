package routes

import (
	"backend/web/controllers"

	"github.com/gin-gonic/gin"
)

func Routes(server *gin.Engine) {
	server.POST("/signup",controllers.Add)
	server.POST("/login",controllers.LoginUser)
	server.POST("/forgetPassword",controllers.ForgetPassword)
}
