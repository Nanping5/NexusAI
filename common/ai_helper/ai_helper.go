package aihelper

import (
	"NexusAi/common/rabbitmq"
	"NexusAi/config"
	"NexusAi/model"
	mylogger "NexusAi/pkg/logger"
	"NexusAi/pkg/utils"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// 系统提示词模板（modelName 和 mode 作为参数传入）
func buildSystemPrompt(defaultPrompt, currentDate, modelName string, mode AgentMode) string {
	base := fmt.Sprintf(`%s

【当前时间】今天是 %s。

【当前模型】你现在正在使用 %s 模型提供服务。当用户询问你是什么模型时，请如实告知你正在使用的模型名称。`, defaultPrompt, currentDate, modelName)

	var modeInstructions string
	switch mode {
	case AgentModeReAct:
		modeInstructions = `

【重要工具使用规则】
当用户的请求包含以下关键词时，必须调用相应的工具：
- 朗读、播放、念、读、听、语音、朗读给我听、念给我听 → 必须使用 text_to_speech 工具
- 搜索、查、最新、新闻 → 使用 search 工具
- 翻译、translate → 使用 translate 工具
- 天气、weather → 使用 weather 工具

【朗读功能特别说明】
当用户要求朗读时，你必须：
1. 先生成正常的文本回复
2. 然后调用 text_to_speech 工具，将回复内容转换为语音
3. 工具会返回音频链接，用户可以点击播放

请严格按照以上规则执行，不要忽略用户的朗读请求。`
	case AgentModePlan:
		modeInstructions = `

【规划模式说明】
请在回答前先分析问题，制定解决步骤，然后按步骤执行。`
	default:
		modeInstructions = ""
	}

	return base + modeInstructions
}

type AIHelper struct {
	model          AIModel
	messages       []*model.Message
	mutex          sync.RWMutex
	saveFuncMu     sync.RWMutex            // 保护 saveFunc 的锁
	SessionID      string
	saveFunc       func(*model.Message) error // 消息回调函数，默认存储到 rabbitmq
	pendingMsgID   string                     // 待确认的消息 ID，用于回滚
	agentService   *AgentService              // Agent 服务（可选）
	agentMode      AgentMode                  // Agent 模式
	config         *AIContextConfig           // 上下文配置
	tokenStats     *TokenStats                // Token 统计
	requestQueue   chan *chatRequest          // 请求队列
	queueCtx       context.Context           // 队列 goroutine 的 context
	queueCancel    context.CancelFunc        // 取消队列 goroutine
	queueOnce      sync.Once                // 确保队列处理器只启动一次
	lastUsage      *schema.TokenUsage         // 最近一次 API 返回的 token 使用量
	lastUsageMutex sync.RWMutex               // 保护 lastUsage
}

// chatRequest 聊天请求
type chatRequest struct {
	ctx        context.Context
	userID     string
	question   string
	cb         StreamCallback
	responseCh chan *chatResponse
}

// chatResponse 聊天响应
type chatResponse struct {
	content string
	err     error
}

// TokenStats Token 统计信息
type TokenStats struct {
	TotalInputTokens  int64 // 总输入 Token 数
	TotalOutputTokens int64 // 总输出 Token 数
	TotalTokens       int64 // 总 Token 数
	RequestCount      int64 // 请求次数
	mutex             sync.RWMutex
}

// NewTokenStats 创建 Token 统计实例
func NewTokenStats() *TokenStats {
	return &TokenStats{}
}

// AddInputTokens 添加输入 Token 数
func (t *TokenStats) AddInputTokens(count int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.TotalInputTokens += int64(count)
	t.TotalTokens += int64(count)
	t.RequestCount++
}

// AddOutputTokens 添加输出 Token 数
func (t *TokenStats) AddOutputTokens(count int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.TotalOutputTokens += int64(count)
	t.TotalTokens += int64(count)
}

// GetStats 获取统计数据
func (t *TokenStats) GetStats() (input, output, total int64, requests int64) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.TotalInputTokens, t.TotalOutputTokens, t.TotalTokens, t.RequestCount
}

// AIContextConfig AI 上下文配置
type AIContextConfig struct {
	MaxMessages int    // 最大消息轮次
	MaxTokens   int    // 最大 Token 数
	Strategy    string // 上下文策略
}

// DefaultAIContextConfig 默认上下文配置
func DefaultAIContextConfig() *AIContextConfig {
	cfg := config.GetConfig().AIConfig
	return &AIContextConfig{
		MaxMessages: cfg.MaxContextMessages,
		MaxTokens:   cfg.MaxContextTokens,
		Strategy:    cfg.ContextStrategy,
	}
}

// NewAIHelper 创建 AIHelper 实例
func NewAIHelper(aiModel AIModel, sessionID string) *AIHelper {
	queueCtx, queueCancel := context.WithCancel(context.Background())
	h := &AIHelper{
		model:    aiModel,
		messages: make([]*model.Message, 0),
		saveFunc: func(msg *model.Message) error {
			data, err := rabbitmq.GenerateMessageMQParam(msg.SessionID, msg.Content, msg.UserID, msg.IsUser)
			if err != nil {
				return err
			}
			return rabbitmq.RMQMessage.Publish(data)
		},
		SessionID:    sessionID,
		agentMode:    AgentModeNone,
		config:       DefaultAIContextConfig(),
		tokenStats:   NewTokenStats(),
		requestQueue: make(chan *chatRequest, 100), // 请求队列，最多缓存 100 个请求
		queueCtx:     queueCtx,
		queueCancel:  queueCancel,
	}
	return h
}

// NewAIHelperWithAgent 创建带 Agent 的 AIHelper 实例
func NewAIHelperWithAgent(agentService *AgentService, sessionID string, mode AgentMode) *AIHelper {
	queueCtx, queueCancel := context.WithCancel(context.Background())
	return &AIHelper{
		messages: make([]*model.Message, 0),
		saveFunc: func(msg *model.Message) error {
			data, err := rabbitmq.GenerateMessageMQParam(msg.SessionID, msg.Content, msg.UserID, msg.IsUser)
			if err != nil {
				return err
			}
			return rabbitmq.RMQMessage.Publish(data)
		},
		SessionID:    sessionID,
		agentService: agentService,
		agentMode:    mode,
		config:       DefaultAIContextConfig(),
		tokenStats:   NewTokenStats(),
		requestQueue: make(chan *chatRequest, 100),
		queueCtx:     queueCtx,
		queueCancel:  queueCancel,
	}
}

// SetAgentMode 设置 Agent 模式
func (a *AIHelper) SetAgentMode(mode AgentMode) {
	a.agentMode = mode
}

// GetAgentMode 获取 Agent 模式
func (a *AIHelper) GetAgentMode() AgentMode {
	return a.agentMode
}

// SetAgentService 设置 Agent 服务
func (a *AIHelper) SetAgentService(agentService *AgentService) {
	a.agentService = agentService
}

// AddMessage 添加消息到对话历史，并通过回调函数保存
func (a *AIHelper) AddMessage(content, userID string, isUser bool, save bool) error {
	msg := &model.Message{
		SessionID: a.SessionID,
		Content:   content,
		UserID:    userID,
		IsUser:    isUser,
	}

	a.mutex.Lock()
	a.messages = append(a.messages, msg)
	a.mutex.Unlock()

	if save {
		if err := a.getSaveFunc()(msg); err != nil {
			mylogger.Logger.Error("save message failed: " + err.Error())
			return err
		}
	}
	return nil
}

// SetSaveFunc 设置消息保存回调函数
func (a *AIHelper) SetSaveFunc(saveFunc func(*model.Message) error) {
	a.saveFuncMu.Lock()
	defer a.saveFuncMu.Unlock()
	a.saveFunc = saveFunc
}

// getSaveFunc 获取消息保存回调函数（线程安全）
func (a *AIHelper) getSaveFunc() func(*model.Message) error {
	a.saveFuncMu.RLock()
	defer a.saveFuncMu.RUnlock()
	return a.saveFunc
}

// GetMessages 获取当前对话历史的消息列表
func (a *AIHelper) GetMessages() []*model.Message {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	out := make([]*model.Message, len(a.messages))
	copy(out, a.messages)
	return out
}

// GetMessageCount 获取消息数量
func (a *AIHelper) GetMessageCount() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.messages)
}

// RemoveLastMessage 移除最后一条消息（用于回滚）
func (a *AIHelper) RemoveLastMessage() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if len(a.messages) > 0 {
		a.messages = a.messages[:len(a.messages)-1]
	}
}

// ========== 上下文管理 ==========

// trimMessagesByCount 根据消息轮次裁剪上下文（滑动窗口策略）
func (a *AIHelper) trimMessagesByCount() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.config == nil || a.config.MaxMessages <= 0 {
		return
	}

	// 计算最大消息数量（一问一答为1轮，所以消息数是轮次的2倍）
	maxCount := a.config.MaxMessages * 2

	if len(a.messages) > maxCount {
		// 保留最近的消息
		a.messages = a.messages[len(a.messages)-maxCount:]
		mylogger.Logger.Info("context trimmed by sliding window",
			zap.Int("before", len(a.messages)+maxCount),
			zap.Int("after", len(a.messages)),
			zap.Int("maxRounds", a.config.MaxMessages),
		)
	}
}

// estimateTokenCount 粗略估算消息的 Token 数量
// 使用简单的启发式方法：平均 4 个字符约等于 1 个 Token（中文）
func (a *AIHelper) estimateTokenCount(messages []*schema.Message) int {
	totalTokens := 0
	for _, msg := range messages {
		// 粗略估算：中文约 1.5 字符/Token，英文约 4 字符/Token
		// 取折中值：3 字符/Token
		totalTokens += len(msg.Content) / 3
	}
	return totalTokens
}

// trimMessagesByTokens 根据 Token 数量裁剪上下文
func (a *AIHelper) trimMessagesByTokens(messages []*schema.Message) []*schema.Message {
	if a.config == nil || a.config.MaxTokens <= 0 {
		return messages
	}

	totalTokens := a.estimateTokenCount(messages)
	if totalTokens <= a.config.MaxTokens {
		return messages
	}

	// 从头部开始移除消息，直到 Token 数量符合限制
	trimmed := make([]*schema.Message, len(messages))
	copy(trimmed, messages)

	for len(trimmed) > 0 {
		firstMsgTokens := len(trimmed[0].Content) / 3
		if totalTokens-firstMsgTokens <= a.config.MaxTokens {
			break
		}
		totalTokens -= firstMsgTokens
		trimmed = trimmed[1:]
	}

	mylogger.Logger.Info("context trimmed by token limit",
		zap.Int("before", len(messages)),
		zap.Int("after", len(trimmed)),
		zap.Int("tokens", totalTokens),
		zap.Int("maxTokens", a.config.MaxTokens),
	)

	return trimmed
}

// applyContextStrategy 应用上下文策略
func (a *AIHelper) applyContextStrategy(messages []*schema.Message) []*schema.Message {
	if a.config == nil {
		return messages
	}

	switch a.config.Strategy {
	case "sliding_window":
		// 滑动窗口策略：在内存中已经通过 trimMessagesByCount 处理
		// 这里再根据 Token 数量做二次裁剪
		return a.trimMessagesByTokens(messages)
	case "summary":
		// 摘要压缩策略（TODO：需要额外的 LLM 调用）
		return a.trimMessagesByTokens(messages)
	default:
		return a.trimMessagesByTokens(messages)
	}
}

// ========== 系统提示词 ==========

// addSystemPrompt 添加系统提示词（包含当前日期和模型名称）
func (a *AIHelper) addSystemPrompt(messages []*schema.Message) []*schema.Message {
	// 获取当前日期和星期
	now := time.Now()
	weekdays := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	weekday := weekdays[now.Weekday()]
	currentDate := fmt.Sprintf("%d年%d月%d日（%s）", now.Year(), now.Month(), now.Day(), weekday)

	// 获取用户自定义的系统提示词（如果有的话）
	defaultPrompt := os.Getenv("DEFAULT_SYSTEM_PROMPT")
	if defaultPrompt == "" {
		defaultPrompt = "你是一个智能助手，帮助用户解决问题。"
	}

	// 获取当前模型名称
	modelName := a.getModelName()

	// 构建系统提示词
	systemPrompt := buildSystemPrompt(defaultPrompt, currentDate, modelName, a.agentMode)

	// 创建系统消息
	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: systemPrompt,
	}

	// 将系统消息插入到消息列表最前面
	result := make([]*schema.Message, 0, len(messages)+1)
	result = append(result, systemMsg)
	result = append(result, messages...)

	return result
}

// getModelName 获取当前模型名称
func (a *AIHelper) getModelName() string {
	if a.agentService != nil {
		return a.agentService.GetModelName()
	}
	if a.model != nil {
		return a.model.GetModelName()
	}
	return "未知模型"
}

// enhanceWithRAG 使用 RAG 检索增强消息上下文（统一在 AIHelper 层调用）
func (a *AIHelper) enhanceWithRAG(messages []*schema.Message) ([]*schema.Message, error) {
	return EnhanceWithRAG(context.Background(), a.SessionID, messages)
}

// ========== 提取的公共逻辑 ==========

// prepareMessages 前置处理：添加用户消息、裁剪上下文、添加系统提示词
// 返回处理后的 schema.Message 列表
func (a *AIHelper) prepareMessages(userQuestion, userID string) ([]*schema.Message, error) {
	if err := a.AddMessage(userQuestion, userID, true, false); err != nil {
		return nil, err
	}

	a.trimMessagesByCount()

	a.mutex.RLock()
	messages := utils.ConvertToSchemaMessages(a.messages)
	a.mutex.RUnlock()

	messages = a.applyContextStrategy(messages)
	messages = a.addSystemPrompt(messages)

	// RAG 增强（统一在 AIHelper 层调用）
	ragMessages, err := a.enhanceWithRAG(messages)
	if err != nil {
		mylogger.Logger.Warn("RAG enhancement failed, using original messages", zap.Error(err))
		return messages, nil
	}
	return ragMessages, nil
}

// saveUserMessage 保存用户消息到外部存储
func (a *AIHelper) saveUserMessage(userID, userQuestion string) {
	userMsg := &model.Message{
		SessionID: a.SessionID,
		Content:   userQuestion,
		UserID:    userID,
		IsUser:    true,
	}
	if err := a.getSaveFunc()(userMsg); err != nil {
		mylogger.Logger.Error("save user message failed: " + err.Error())
	}
}

// recordTokenStats 记录 token 统计（非流式响应）
func (a *AIHelper) recordTokenStats(schemaMsg *schema.Message, messages []*schema.Message) {
	if a.tokenStats == nil {
		return
	}

	if schemaMsg.ResponseMeta != nil && schemaMsg.ResponseMeta.Usage != nil {
		usage := schemaMsg.ResponseMeta.Usage
		a.tokenStats.AddInputTokens(usage.PromptTokens)
		a.tokenStats.AddOutputTokens(usage.CompletionTokens)
		mylogger.Logger.Info("token stats from usage",
			zap.Int("input", usage.PromptTokens),
			zap.Int("output", usage.CompletionTokens),
			zap.Int("total", usage.TotalTokens),
		)
	} else {
		inputTokens := a.estimateTokenCount(messages)
		outputTokens := len(schemaMsg.Content) / 3
		a.tokenStats.AddInputTokens(inputTokens)
		a.tokenStats.AddOutputTokens(outputTokens)
		mylogger.Logger.Info("token stats (estimated)",
			zap.Int("input", inputTokens),
			zap.Int("output", outputTokens),
		)
	}
}

// recordStreamTokenStats 记录 token 统计（流式响应）
func (a *AIHelper) recordStreamTokenStats(content string, messages []*schema.Message) {
	if a.tokenStats == nil {
		return
	}

	var usage *schema.TokenUsage
	if a.agentMode == AgentModeReAct && a.agentService != nil {
		usage = a.agentService.GetLastUsage()
	} else if a.model != nil {
		usage = a.model.GetLastUsage()
	}

	if usage != nil {
		a.tokenStats.AddInputTokens(usage.PromptTokens)
		a.tokenStats.AddOutputTokens(usage.CompletionTokens)
		mylogger.Logger.Info("token stats from usage",
			zap.Int("input", usage.PromptTokens),
			zap.Int("output", usage.CompletionTokens),
			zap.Int("total", usage.TotalTokens),
		)
	} else {
		inputTokens := a.estimateTokenCount(messages)
		outputTokens := len(content) / 3
		a.tokenStats.AddInputTokens(inputTokens)
		a.tokenStats.AddOutputTokens(outputTokens)
		mylogger.Logger.Info("token stats (estimated)",
			zap.Int("input", inputTokens),
			zap.Int("output", outputTokens),
		)
	}
}

// GenerateResponse 生成 AI 响应，并将用户问题和 AI 回答添加到对话历史中
func (a *AIHelper) GenerateResponse(ctx context.Context, userID string, userQuestion string) (*model.Message, error) {
	messages, err := a.prepareMessages(userQuestion, userID)
	if err != nil {
		return nil, err
	}

	var schemaMsg *schema.Message

	if a.agentMode == AgentModeReAct && a.agentService != nil {
		schemaMsg, err = a.agentService.GenerateResponse(ctx, messages)
	} else if a.model != nil {
		schemaMsg, err = a.model.GenerateResponse(ctx, messages)
	} else {
		a.RemoveLastMessage()
		return nil, fmt.Errorf("no model or agent service configured")
	}

	if err != nil {
		a.RemoveLastMessage()
		return nil, err
	}

	a.saveUserMessage(userID, userQuestion)

	modelMsg := utils.ConvertToModelMessage(a.SessionID, userID, schemaMsg)
	if err := a.AddMessage(modelMsg.Content, userID, false, true); err != nil {
		return nil, err
	}

	a.recordTokenStats(schemaMsg, messages)

	return modelMsg, nil
}

// StreamResponse 实现流式响应，实时回调消息片段，并将完整响应添加到对话历史中
func (a *AIHelper) StreamResponse(ctx context.Context, userID string, userQuestion string, cb StreamCallback) (string, error) {
	messages, err := a.prepareMessages(userQuestion, userID)
	if err != nil {
		return "", err
	}

	var content string

	if a.agentMode == AgentModeReAct && a.agentService != nil {
		content, err = a.agentService.StreamResponse(ctx, messages, cb)
	} else if a.model != nil {
		content, err = a.model.StreamResponse(ctx, messages, cb)
	} else {
		a.RemoveLastMessage()
		return "", fmt.Errorf("no model or agent service configured")
	}

	if err != nil {
		a.RemoveLastMessage()
		return "", err
	}

	a.saveUserMessage(userID, userQuestion)

	if err := a.AddMessage(content, userID, false, true); err != nil {
		return "", err
	}

	a.recordStreamTokenStats(content, messages)

	return content, nil
}

// ========== 会话级请求队列（防止并发消息错乱） ==========

// ensureQueueProcessorStarted 懒启动队列处理器（确保只有一个长活 goroutine）
func (a *AIHelper) ensureQueueProcessorStarted() {
	a.queueOnce.Do(func() {
		go a.processQueue()
	})
}

// QueueGenerateResponse 非流式请求队列
func (a *AIHelper) QueueGenerateResponse(ctx context.Context, userID, question string) (*model.Message, error) {
	// 创建请求（cb 为空，表示非流式）
	req := &chatRequest{
		ctx:        ctx,
		userID:     userID,
		question:   question,
		cb:         nil, // 非流式
		responseCh: make(chan *chatResponse, 1),
	}

	// 确保队列处理器已启动
	a.ensureQueueProcessorStarted()

	// 将请求加入队列
	select {
	case a.requestQueue <- req:
		// 成功加入队列
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 等待响应
	select {
	case resp := <-req.responseCh:
		if resp.err != nil {
			return nil, resp.err
		}
		// 返回消息结构
		return &model.Message{
			SessionID: a.SessionID,
			Content:   resp.content,
			UserID:    userID,
			IsUser:    false,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// QueueStreamResponse 将请求加入队列并等待响应
// 确保同一会话内的消息按顺序处理
func (a *AIHelper) QueueStreamResponse(ctx context.Context, userID, question string, cb StreamCallback) (string, error) {
	// 创建请求
	req := &chatRequest{
		ctx:        ctx,
		userID:     userID,
		question:   question,
		cb:         cb,
		responseCh: make(chan *chatResponse, 1),
	}

	// 确保队列处理器已启动
	a.ensureQueueProcessorStarted()

	// 将请求加入队列
	select {
	case a.requestQueue <- req:
		// 成功加入队列
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// 等待响应
	select {
	case resp := <-req.responseCh:
		return resp.content, resp.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// processQueue 处理请求队列（长活 goroutine，通过 context 退出）
func (a *AIHelper) processQueue() {
	for {
		select {
		case <-a.queueCtx.Done():
			return // context 取消时优雅退出
		case req := <-a.requestQueue:
			if req.cb != nil {
				content, err := a.StreamResponse(req.ctx, req.userID, req.question, req.cb)
				req.responseCh <- &chatResponse{content: content, err: err}
			} else {
				msg, err := a.GenerateResponse(req.ctx, req.userID, req.question)
				if err != nil {
					req.responseCh <- &chatResponse{content: "", err: err}
				} else {
					req.responseCh <- &chatResponse{content: msg.Content, err: nil}
				}
			}
		}
	}
}

// GetTokenStats 获取 Token 统计信息
func (a *AIHelper) GetTokenStats() (input, output, total int64, requests int64) {
	if a.tokenStats == nil {
		return 0, 0, 0, 0
	}
	return a.tokenStats.GetStats()
}

// GetModelType 获取当前使用的 AI 模型类型
func (a *AIHelper) GetModelType() string {
	if a.agentService != nil {
		return a.agentService.GetModelType()
	}
	if a.model != nil {
		return a.model.GetModelType()
	}
	return "unknown"
}
