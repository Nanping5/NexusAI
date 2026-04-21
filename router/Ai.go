package router

import (
	ret "NexusAi/common/code"
	sessionController "NexusAi/controller/session"
	response "NexusAi/common/response/common"
	"NexusAi/middleware"
	"NexusAi/service"

	"github.com/gin-gonic/gin"
)

// RegisterAIRouter 注册 AI 相关路由
func RegisterAIRouter(r *gin.RouterGroup) {
	// 会话管理
	r.GET("/sessions", sessionController.GetUserSessionsByUserID)
	r.DELETE("/session", sessionController.DeleteSession)

	// 聊天历史
	r.GET("/history", sessionController.ChatHistory)

	// 获取可用模型列表
	r.GET("/models", func(c *gin.Context) {
		models, err := service.ModelConfigService.GetAvailableModelsForUser(c.Request.Context())
		if err != nil {
			response.Fail(c, ret.CodeServerBusy)
			return
		}
		response.Success(c, gin.H{"models": models})
	})

	// 创建会话并发送消息（带限流保护）
	r.POST("/session/create", middleware.RateLimit(middleware.RateLimitConfig{
		Limit:   10, // 每分钟最多 10 次
		APIName: "ai_create",
	}), sessionController.CreateSessionAndSendMessage)
	r.POST("/session/create/stream", middleware.RateLimit(middleware.RateLimitConfig{
		Limit:   10,
		APIName: "ai_create_stream",
	}), sessionController.CreateStreamSessionAndSendMessage)

	// 发送消息（已存在会话，带限流保护）
	r.POST("/chat", middleware.RateLimit(middleware.RateLimitConfig{
		Limit:   30, // 每分钟最多 30 次
		APIName: "ai_chat",
	}), sessionController.ChatSendMessage)
	r.POST("/chat/stream", middleware.RateLimit(middleware.RateLimitConfig{
		Limit:   30,
		APIName: "ai_chat_stream",
	}), sessionController.StreamChatSendMessage)
}
