package service

import (
	aihelper "NexusAi/common/ai_helper"
	ret "NexusAi/common/code"
	mmysql "NexusAi/common/mysql"
	myredis "NexusAi/common/redis"
	"NexusAi/common/rag"
	SessionResponse "NexusAi/common/response"
	"NexusAi/dao"
	"NexusAi/model"
	mylogger "NexusAi/pkg/logger"
	"NexusAi/pkg/utils"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type sessionService struct {
}

var SessionService = &sessionService{}

// getAgentConfig 获取 Agent 配置
func getAgentConfig(useAgent bool) aihelper.AgentConfig {
	if !useAgent {
		return aihelper.AgentConfig{
			Mode:          aihelper.AgentModeNone,
			MaxIterations: 10,
		}
	}

	maxIterations := 10
	if iter := os.Getenv("AGENT_MAX_ITERATIONS"); iter != "" {
		fmt.Sscanf(iter, "%d", &maxIterations)
	}

	return aihelper.AgentConfig{
		Mode:          aihelper.AgentModeReAct,
		MaxIterations: maxIterations,
	}
}

// getOrCreateHelper 获取或创建 AIHelper（统一逻辑，减少重复代码）
func (s *sessionService) getOrCreateHelper(ctx context.Context, manager *aihelper.AIHelperManager, userID, sessionID, configID string, useAgent bool) (*aihelper.AIHelper, error) {
	agentConfig := getAgentConfig(useAgent)
	if agentConfig.Mode == aihelper.AgentModeReAct {
		return manager.GetOrCreateAIHelperWithAgentByConfigID(ctx, userID, sessionID, configID, agentConfig)
	}
	return manager.GetOrCreateAIHelperByConfigID(ctx, userID, sessionID, configID)
}

// CreateSessionAndSendMessage 创建会话并发送消息
func (s *sessionService) CreateSessionAndSendMessage(ctx context.Context, userID, question, configID string, useAgent bool) (SessionResponse.CreateSessionAndSendMessageResponse, ret.Code) {
	newSession := &model.Session{
		SessionID: utils.GenerateSessionID(),
		UserID:    userID,
		Title:     question,
	}

	createdSessionID, err := dao.SessionDAO.CreateSession(ctx, newSession)
	if err != nil {
		mylogger.Logger.Error(err.Error())
		return SessionResponse.CreateSessionAndSendMessageResponse{}, ret.CodeServerBusy
	}

	// 使会话列表缓存失效
	if err := myredis.InvalidateUserSessionList(userID); err != nil {
		mylogger.Logger.Error("CreateSessionAndSendMessage Redis invalidate error: " + err.Error())
	}

	manager := aihelper.GetGlobalAIHelperManager()
	helper, err := s.getOrCreateHelper(ctx, manager, userID, createdSessionID, configID, useAgent)
	if err != nil {
		mylogger.Logger.Error(err.Error())
		return SessionResponse.CreateSessionAndSendMessageResponse{}, ret.AIModelFail
	}

	// 生成 AI 响应（使用队列保证顺序）
	responseMessage, err := helper.QueueGenerateResponse(ctx, userID, question)
	if err != nil {
		mylogger.Logger.Error("CreateSessionAndSendMessage GenerateResponse error: " + err.Error())
		return SessionResponse.CreateSessionAndSendMessageResponse{}, ret.AIModelFail
	}

	return SessionResponse.CreateSessionAndSendMessageResponse{
		SessionID: createdSessionID,
		AIMessage: responseMessage.Content,
	}, ret.CodeSuccess
}

// CreateStreamSessionOnly 仅创建会话（用于流式响应）
func (s *sessionService) CreateStreamSessionOnly(ctx context.Context, userID, question string) (string, ret.Code) {
	newSession := &model.Session{
		SessionID: utils.GenerateSessionID(),
		UserID:    userID,
		Title:     question,
	}
	createdSessionID, err := dao.SessionDAO.CreateSession(ctx, newSession)
	if err != nil {
		mylogger.Logger.Error("CreateStreamSessionOnly CreateSession error: " + err.Error())
		return "", ret.CodeServerBusy
	}

	// 使会话列表缓存失效
	if err := myredis.InvalidateUserSessionList(userID); err != nil {
		mylogger.Logger.Error("CreateStreamSessionOnly Redis invalidate error: " + err.Error())
	}

	return createdSessionID, ret.CodeSuccess
}

// StreamSendMessageToExistSession 向已存在的会话发送流式消息
func (s *sessionService) StreamSendMessageToExistSession(ctx context.Context, userID, sessionID, question, configID string, useAgent bool, w http.ResponseWriter) ret.Code {
	flusher, ok := w.(http.Flusher)
	if !ok {
		mylogger.Logger.Error("ResponseWriter does not support flushing")
		return ret.CodeServerBusy
	}

	// 验证用户是否有权限访问该会话
	sessionInfo, err := dao.SessionDAO.GetSessionInfo(ctx, userID, sessionID)
	if err != nil || sessionInfo.SessionID == "" {
		mylogger.Logger.Warn("StreamSendMessageToExistSession: session not found or access denied")
		return ret.CodeRecordNotFound
	}

	// 添加超时控制，防止长时间连接
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// 创建可取消的 context，用于客户端断开时取消 AI 请求
	ctx, requestCancel := context.WithCancel(ctx)
	defer requestCancel()

	manager := aihelper.GetGlobalAIHelperManager()
	helper, err := s.getOrCreateHelper(ctx, manager, userID, sessionID, configID, useAgent)
	if err != nil {
		mylogger.Logger.Error("StreamSendMessageToExistSession GetOrCreateAIHelper error: " + err.Error())
		return ret.AIModelFail
	}

	// 监控客户端断开连接
	clientDisconnected := false
	cb := func(message string) {
		if clientDisconnected {
			return
		}
		_, err := w.Write([]byte("data:" + message + "\n\n"))
		if err != nil {
			mylogger.Logger.Error("SSE Write error: " + err.Error())
			clientDisconnected = true
			requestCancel() // 取消 AI 请求
			return
		}
		flusher.Flush()
	}

	// 使用队列处理请求，防止并发消息错乱
	_, err = helper.QueueStreamResponse(ctx, userID, question, cb)
	if err != nil {
		// 如果是客户端断开连接，不算错误
		if ctx.Err() == context.Canceled {
			mylogger.Logger.Info("StreamSendMessageToExistSession: client disconnected")
			return ret.CodeSuccess
		}
		mylogger.Logger.Error("StreamSendMessageToExistSession QueueStreamResponse error: " + err.Error())
		return ret.AIModelFail
	}

	// 检查客户端是否已断开，避免向已断开的客户端写入
	if clientDisconnected {
		return ret.CodeSuccess
	}

	_, err = w.Write([]byte("data:[DONE]\n\n"))
	if err != nil {
		mylogger.Logger.Error("SSE Write [DONE] error: " + err.Error())
		return ret.CodeServerBusy
	}

	flusher.Flush()
	return ret.CodeSuccess
}

// GetUserSessionsByUserID 获取用户的所有会话（带 Redis 缓存）
func (s *sessionService) GetUserSessionsByUserID(ctx context.Context, userID string) ([]model.SessionInfo, error) {
	// 1. 先从 Redis 缓存获取
	cached, err := myredis.GetUserSessionList(userID)
	if err != nil {
		mylogger.Logger.Warn("GetUserSessionsByUserID Redis get error: " + err.Error())
		// Redis 出错时继续从数据库查询
	} else if cached != nil && len(cached) > 0 {
		// 缓存命中，转换为 SessionInfo
		sessions := make([]model.SessionInfo, 0, len(cached))
		for _, item := range cached {
			createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
			sessions = append(sessions, model.SessionInfo{
				SessionID: item.SessionID,
				Title:     item.Title,
				Model: &gorm.Model{
					CreatedAt: createdAt,
				},
			})
		}
		return sessions, nil
	}

	// 2. 从数据库查询
	sessions, err := dao.SessionDAO.GetSessionsByUserID(ctx, userID)
	if err != nil {
		mylogger.Logger.Error("GetUserSessionsByUserID error: " + err.Error())
		return nil, err
	}

	// 3. 写入 Redis 缓存
	cacheItems := make([]myredis.SessionListItem, 0, len(sessions))
	for _, s := range sessions {
		cacheItems = append(cacheItems, myredis.SessionListItem{
			SessionID: s.SessionID,
			Title:     s.Title,
			CreatedAt: s.CreatedAt.Format(time.RFC3339),
		})
	}
	if err := myredis.SetUserSessionList(userID, cacheItems); err != nil {
		mylogger.Logger.Warn("GetUserSessionsByUserID Redis set error: " + err.Error())
	}

	return sessions, nil
}

// ChatSendMessage 发送聊天消息
func (s *sessionService) ChatSendMessage(ctx context.Context, userID, sessionID, question, configID string, useAgent bool) (string, ret.Code) {
	manager := aihelper.GetGlobalAIHelperManager()

	// 检查 AIHelper 是否存在，不存在则尝试从数据库恢复
	helper, exists := manager.GetAIHelper(userID, sessionID)
	if !exists {
		// 尝试从数据库恢复会话历史
		var err error
		helper, err = s.getOrCreateHelper(ctx, manager, userID, sessionID, configID, useAgent)
		if err != nil {
			mylogger.Logger.Error("ChatSendMessage GetOrCreateAIHelper error: " + err.Error())
			return "", ret.AIModelFail
		}
		// 恢复历史消息
		if err := s.restoreSessionMessages(ctx, sessionID, helper); err != nil {
			mylogger.Logger.Error("ChatSendMessage restoreSessionMessages error: " + err.Error())
			return "", ret.AIModelFail
		}
	}

	aiResponse, err := helper.QueueGenerateResponse(ctx, userID, question)
	if err != nil {
		mylogger.Logger.Error("ChatSendMessage GenerateResponse error: " + err.Error())
		return "", ret.AIModelFail
	}

	return aiResponse.Content, ret.CodeSuccess
}

// StreamChatSendMessage 流式发送聊天消息
func (s *sessionService) StreamChatSendMessage(ctx context.Context, userID, sessionID, question, configID string, useAgent bool, w http.ResponseWriter) ret.Code {
	manager := aihelper.GetGlobalAIHelperManager()

	// 检查 AIHelper 是否存在，不存在则创建
	_, exists := manager.GetAIHelper(userID, sessionID)
	if !exists {
		_, err := s.getOrCreateHelper(ctx, manager, userID, sessionID, configID, useAgent)
		if err != nil {
			mylogger.Logger.Error("StreamChatSendMessage GetOrCreateAIHelper error: " + err.Error())
			return ret.AIModelFail
		}
	}

	return s.StreamSendMessageToExistSession(ctx, userID, sessionID, question, configID, useAgent, w)
}

// GetChatHistory 获取聊天历史
func (s *sessionService) GetChatHistory(ctx context.Context, userID, sessionID string) (SessionResponse.ChatHistoryResponse, ret.Code) {
	// 验证用户是否有权限访问该会话
	sessionInfo, err := dao.SessionDAO.GetSessionInfo(ctx, userID, sessionID)
	if err != nil {
		mylogger.Logger.Error("GetChatHistory GetSessionInfo error: " + err.Error())
		return SessionResponse.ChatHistoryResponse{}, ret.CodeServerBusy
	}
	if sessionInfo.SessionID == "" {
		return SessionResponse.ChatHistoryResponse{}, ret.CodeRecordNotFound
	}

	// 先从内存获取
	manager := aihelper.GetGlobalAIHelperManager()
	helper, exists := manager.GetAIHelper(userID, sessionID)

	if exists {
		// 从内存获取
		messages := helper.GetMessages()
		history := make([]model.History, 0, len(messages))
		for _, msg := range messages {
			history = append(history, model.History{
				Content: msg.Content,
				IsUser:  msg.IsUser,
			})
		}
		return SessionResponse.ChatHistoryResponse{SessionID: sessionID, History: history}, ret.CodeSuccess
	}

	// 内存中没有，尝试从 Redis 缓存获取
	cachedHistory, err := myredis.GetSessionHistory(sessionID)
	if err != nil {
		mylogger.Logger.Warn("GetChatHistory Redis get error: " + err.Error())
	} else if len(cachedHistory) > 0 {
		history := make([]model.History, 0, len(cachedHistory))
		for _, item := range cachedHistory {
			history = append(history, model.History{
				Content: item.Content,
				IsUser:  item.IsUser,
			})
		}
		return SessionResponse.ChatHistoryResponse{SessionID: sessionID, History: history}, ret.CodeSuccess
	}

	// 从数据库加载
	messages, err := dao.MessageDAO.GetMessagesBySessionID(ctx, sessionID)
	if err != nil {
		mylogger.Logger.Error("GetChatHistory GetMessagesBySessionID error: " + err.Error())
		return SessionResponse.ChatHistoryResponse{}, ret.CodeServerBusy
	}
	history := make([]model.History, 0, len(messages))
	cacheItems := make([]myredis.HistoryItem, 0, len(messages))
	for _, msg := range messages {
		history = append(history, model.History{
			Content: msg.Content,
			IsUser:  msg.IsUser,
		})
		cacheItems = append(cacheItems, myredis.HistoryItem{
			Content: msg.Content,
			IsUser:  msg.IsUser,
		})
	}

	// 写入 Redis 缓存
	if err := myredis.SetSessionHistory(sessionID, cacheItems); err != nil {
		mylogger.Logger.Warn("GetChatHistory Redis set error: " + err.Error())
	}

	return SessionResponse.ChatHistoryResponse{SessionID: sessionID, History: history}, ret.CodeSuccess
}

// DeleteSession 删除会话（同时删除数据库记录和内存缓存）
func (s *sessionService) DeleteSession(ctx context.Context, userID, sessionID string) ret.Code {
	// 1. 首先验证会话是否属于该用户
	sessionInfo, err := dao.SessionDAO.GetSessionInfo(ctx, userID, sessionID)
	if err != nil {
		mylogger.Logger.Error("DeleteSession GetSessionInfo error: " + err.Error())
		return ret.CodeServerBusy
	}
	if sessionInfo.SessionID == "" {
		return ret.CodeRecordNotFound
	}

	// 2. 使用事务删除数据库记录
	db := mmysql.GetDB()
	if db == nil {
		mylogger.Logger.Error("DeleteSession: database not initialized")
		return ret.CodeServerBusy
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除消息记录
		if err := tx.Where("session_id = ?", sessionID).Delete(&model.Message{}).Error; err != nil {
			return fmt.Errorf("delete messages failed: %w", err)
		}
		// 删除会话记录
		if err := tx.Where("user_id = ? AND session_id = ?", userID, sessionID).Delete(&model.Session{}).Error; err != nil {
			return fmt.Errorf("delete session failed: %w", err)
		}
		return nil
	})

	if err != nil {
		mylogger.Logger.Error("DeleteSession transaction error: " + err.Error())
		return ret.CodeServerBusy
	}

	// 3. 数据库删除成功后，从内存中移除 AIHelper
	manager := aihelper.GetGlobalAIHelperManager()
	manager.RemoveAIHelper(userID, sessionID)

	// 4. 使会话列表缓存失效
	if err := myredis.InvalidateUserSessionList(userID); err != nil {
		mylogger.Logger.Error("DeleteSession Redis invalidate error: " + err.Error())
	}

	// 5. 删除 session 级别的 RAG 文件和索引
	s.deleteSessionRAGFiles(ctx, userID, sessionID)

	return ret.CodeSuccess
}

// deleteSessionRAGFiles 删除会话的 RAG 文件和索引
func (s *sessionService) deleteSessionRAGFiles(ctx context.Context, userID, sessionID string) {
	// 验证路径安全
	if !isValidID(userID) || !isValidID(sessionID) {
		mylogger.Logger.Error("Invalid userID or sessionID for path traversal check")
		return
	}

	// 删除 Qdrant 中的向量点（使用 session_id 过滤）
	if err := rag.DeleteSessionPoints(ctx, sessionID); err != nil {
		mylogger.Logger.Error("Failed to delete session points from Qdrant: " + err.Error())
	}

	// 删除整个 session 目录
	sessionDir := filepath.Join("uploads", userID, sessionID)
	if err := os.RemoveAll(sessionDir); err != nil {
		mylogger.Logger.Error("Failed to remove session directory: " + err.Error())
	}
}

// RestoreSessionFromDB 从数据库恢复会话历史
func (s *sessionService) RestoreSessionFromDB(ctx context.Context, userID, sessionID, configID string, useAgent bool) error {
	// 首先验证会话是否属于该用户
	sessionInfo, err := dao.SessionDAO.GetSessionInfo(ctx, userID, sessionID)
	if err != nil {
		return fmt.Errorf("verify session ownership failed: %w", err)
	}
	if sessionInfo.SessionID == "" {
		return fmt.Errorf("session not found or access denied")
	}

	// 从数据库加载历史消息
	messages, err := dao.MessageDAO.GetMessagesBySessionID(ctx, sessionID)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return nil
	}

	// 创建 AIHelper 并恢复历史
	manager := aihelper.GetGlobalAIHelperManager()
	helper, err := s.getOrCreateHelper(ctx, manager, userID, sessionID, configID, useAgent)
	if err != nil {
		return err
	}

	// 恢复历史消息到 helper（不保存到数据库，因为已经在数据库中了）
	for _, msg := range messages {
		helper.AddMessage(msg.Content, msg.UserID, msg.IsUser, false)
	}

	return nil
}

// restoreSessionMessages 恢复会话消息到 helper
func (s *sessionService) restoreSessionMessages(ctx context.Context, sessionID string, helper *aihelper.AIHelper) error {
	messages, err := dao.MessageDAO.GetMessagesBySessionID(ctx, sessionID)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		helper.AddMessage(msg.Content, msg.UserID, msg.IsUser, false)
	}

	return nil
}

// isValidID 验证 ID 是否安全（防止路径遍历攻击）
func isValidID(id string) bool {
	if id == "" {
		return false
	}
	// 禁止路径遍历字符
	return !strings.Contains(id, "..") && !strings.Contains(id, "/") && !strings.Contains(id, "\\")
}
