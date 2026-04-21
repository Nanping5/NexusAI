package router

import (
	onlineController "NexusAi/controller/online"
	userController "NexusAi/controller/user"
	"NexusAi/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {

	r := gin.Default()
	//CORS 跨域资源共享
	r.Use(middleware.CORS())

	// 静态文件服务 - 提供音频文件访问
	r.Static("/static", "./uploads")

	// 前端静态文件服务
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/favicon.svg", "./frontend/dist/favicon.svg")

	// 前端路由 - 所有非 API 请求返回 index.html（支持 Vue Router history 模式）
	r.NoRoute(func(c *gin.Context) {
		// 如果是 API 请求但未匹配到路由，返回 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"code": 404, "msg": "接口不存在"})
			return
		}
		c.File("./frontend/dist/index.html")
	})

	enterRouter := r.Group("/api/v1")

	// 用户公开路由（无需认证）
	RegisterUserRouter(enterRouter.Group("/user"))

	// 用户认证路由
	userAuth := enterRouter.Group("/user")
	userAuth.Use(middleware.JWT())
	{
		userAuth.GET("/info", userController.GetUserInfo)
		userAuth.PUT("/nickname", userController.UpdateNickname)
		userAuth.POST("/heartbeat", onlineController.Heartbeat)
	}

	// AI 接口（需要认证）
	aiGroup := enterRouter.Group("/ai")
	aiGroup.Use(middleware.JWT())
	RegisterAIRouter(aiGroup)

	//file接口（需要认证）
	fileGroup := enterRouter.Group("/file")
	fileGroup.Use(middleware.JWT())
	RegisterFileRouter(fileGroup)

	//tts接口（需要认证）
	ttsGroup := enterRouter.Group("/voice")
	ttsGroup.Use(middleware.JWT())
	RegisterTTSRouter(ttsGroup)

	// 管理员接口
	adminGroup := enterRouter.Group("/admin")
	RegisterAdminRouter(adminGroup)

	// 在线用户统计（无需认证）
	enterRouter.GET("/online/count", onlineController.GetOnlineCount)

	return r
}
