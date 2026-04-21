package router

import (
	adminController "NexusAi/controller/admin"
	"NexusAi/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRouter 注册管理员路由
func RegisterAdminRouter(r *gin.RouterGroup) {
	// 管理员公开路由（无需认证）
	r.POST("/login", adminController.Login)

	// 管理员认证路由
	auth := r.Group("")
	auth.Use(middleware.AdminJWT())
	{
		// 管理员信息
		auth.GET("/info", adminController.GetInfo)

		// 模型配置管理
		auth.GET("/models", adminController.GetAllModelConfigs)
		auth.GET("/models/:config_id", adminController.GetModelConfig)
		auth.POST("/models", adminController.CreateModelConfig)
		auth.PUT("/models/:config_id", adminController.UpdateModelConfig)
		auth.DELETE("/models/:config_id", adminController.DeleteModelConfig)
		auth.PUT("/models/:config_id/default", adminController.SetDefaultModelConfig)
		auth.PUT("/models/:config_id/toggle", adminController.ToggleModelConfig)
	}
}
