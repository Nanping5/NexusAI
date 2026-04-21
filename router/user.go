package router

import (
	userController "NexusAi/controller/user"

	"github.com/gin-gonic/gin"
)

func RegisterUserRouter(r *gin.RouterGroup) {
	{
		r.POST("/register", userController.Register)
		r.POST("/login", userController.Login)
		r.GET("/send_captcha", userController.SendCaptcha)

	}
}
