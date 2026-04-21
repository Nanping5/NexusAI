package sessionController

import (
	ret "NexusAi/common/code"
	"NexusAi/common/request"
	response "NexusAi/common/response/common"
	"NexusAi/service"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserSessionsByUserID(c *gin.Context) {

	userId := c.GetString("user_id")
	if userId == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}
	userSessions, err := service.SessionService.GetUserSessionsByUserID(c.Request.Context(), userId)
	if err != nil {
		response.Fail(c, ret.CodeServerBusy)
		return
	}
	response.Success(c, gin.H{"sessions": userSessions})
}

// CreateSessionAndSendMessage 首次创建会话并发送消息
func CreateSessionAndSendMessage(c *gin.Context) {
	userId := c.GetString("user_id")
	if userId == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}
	req := request.CreateSessionAndSendMessageRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}
	resp, errCode := service.SessionService.CreateSessionAndSendMessage(c.Request.Context(), userId, req.UserQuestion, req.ConfigID, req.UseAgent)
	if errCode != ret.CodeSuccess {
		response.Fail(c, errCode)
		return
	}
	response.Success(c, resp)

}

// CreateStreamSessionAndSendMessage 创建会话并流式发送消息
func CreateStreamSessionAndSendMessage(c *gin.Context) {
	userId := c.GetString("user_id")
	if userId == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}
	req := request.CreateSessionAndSendMessageRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	// SSE 头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no") // 禁止代理缓存

	// 先创建会话拿到 sessionID，再流式输出
	SessionID, errCode := service.SessionService.CreateStreamSessionOnly(c.Request.Context(), userId, req.UserQuestion)
	if errCode != ret.CodeSuccess {
		c.SSEvent("error", gin.H{"message": "Failed To Create Session!"})
		return
	}

	// 把 sessionID 给前端
	jsonData, _ := json.Marshal(map[string]string{"session_id": SessionID})
	c.Writer.WriteString("data: " + string(jsonData) + "\n\n")
	c.Writer.Flush()

	errcode := service.SessionService.StreamSendMessageToExistSession(c.Request.Context(), userId, SessionID, req.UserQuestion, req.ConfigID, req.UseAgent, http.ResponseWriter(c.Writer))
	if errcode != ret.CodeSuccess {
		c.SSEvent("error", gin.H{"message": "Failed To Send Message!"})
		return
	}

}

// ChatSendMessage 发送消息接口
func ChatSendMessage(c *gin.Context) {
	req := request.ChatSendMessageRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}
	userId := c.GetString("user_id")
	if userId == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}

	aiMessage, errCode := service.SessionService.ChatSendMessage(c.Request.Context(), userId, req.SessionID, req.UserQuestion, req.ConfigID, req.UseAgent)
	if errCode != ret.CodeSuccess {
		response.Fail(c, errCode)
		return
	}
	response.Success(c, gin.H{"ai_message": aiMessage})
}

// StreamChatSendMessage 流式发送消息
func StreamChatSendMessage(c *gin.Context) {

	req := request.ChatSendMessageRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}
	userId := c.GetString("user_id")
	if userId == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}

	// SSE 头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no") // 禁止代理缓存

	errcode := service.SessionService.StreamChatSendMessage(c.Request.Context(), userId, req.SessionID, req.UserQuestion, req.ConfigID, req.UseAgent, http.ResponseWriter(c.Writer))
	if errcode != ret.CodeSuccess {
		c.SSEvent("error", gin.H{"message": "Failed To Send Message!"})
		return
	}
}

// ChatHistory 获取聊天历史
func ChatHistory(c *gin.Context) {
	userId := c.GetString("user_id")
	if userId == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}
	sessionId := c.Query("session_id")
	if sessionId == "" {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	resp, errCode := service.SessionService.GetChatHistory(c.Request.Context(), userId, sessionId)
	if errCode != ret.CodeSuccess {
		response.Fail(c, errCode)
		return
	}
	response.Success(c, resp)
}

// DeleteSession 删除会话
func DeleteSession(c *gin.Context) {

	userId := c.GetString("user_id")
	if userId == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}
	sessionId := c.Query("session_id")
	if sessionId == "" {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	errCode := service.SessionService.DeleteSession(c.Request.Context(), userId, sessionId)
	if errCode != ret.CodeSuccess {
		response.Fail(c, errCode)
		return
	}
	response.Success(c, nil)
}
