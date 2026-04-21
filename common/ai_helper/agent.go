package aihelper

import (
	"context"
	"fmt"
	"io"
	"strings"

	"NexusAi/common/mcpmanager"
	"NexusAi/model"
	mylogger "NexusAi/pkg/logger"

	"github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// AgentMode Agent 模式类型
type AgentMode string

const (
	AgentModeNone  AgentMode = "none"
	AgentModeReAct AgentMode = "react"
	AgentModePlan  AgentMode = "plan"
)

// AgentConfig Agent 配置
type AgentConfig struct {
	Mode          AgentMode // Agent 模式
	MaxIterations int       // 最大迭代次数（Agent 执行步骤限制）
}

// AgentService Agent 服务，封装 ReAct Agent 逻辑
type AgentService struct {
	agent     *react.Agent
	model     einomodel.ToolCallingChatModel
	modelName string
	userID    string
	sessionID string
	config    AgentConfig
	lastUsage *schema.TokenUsage
}

// NewAgentServiceByConfigID 根据配置 ID 创建 Agent 服务实例
func NewAgentServiceByConfigID(ctx context.Context, configID, userID, sessionID string, agentConfig AgentConfig) (*AgentService, error) {
	// 获取模型配置
	config, err := GetConfigByConfigID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("model config not found: %s", configID)
	}
	if !config.IsEnabled {
		return nil, fmt.Errorf("model %s is disabled", config.Name)
	}

	return NewAgentServiceFromConfig(ctx, config, userID, sessionID, agentConfig)
}

// NewAgentServiceFromConfig 根据配置创建 Agent 服务实例
func NewAgentServiceFromConfig(ctx context.Context, config *model.AIModelConfig, userID, sessionID string, agentConfig AgentConfig) (*AgentService, error) {
	// 创建聊天模型
	chatModel, err := createChatModelFromConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	// 设置默认配置
	if agentConfig.MaxIterations <= 0 {
		agentConfig.MaxIterations = 10
	}
	if agentConfig.Mode == "" {
		agentConfig.Mode = AgentModeReAct
	}

	return &AgentService{
		model:     chatModel,
		modelName: config.Name,
		userID:    userID,
		sessionID: sessionID,
		config:    agentConfig,
	}, nil
}

// createChatModelFromConfig 根据配置创建聊天模型
func createChatModelFromConfig(ctx context.Context, config *model.AIModelConfig) (einomodel.ToolCallingChatModel, error) {
	if config.ModelType != "openai_compatible" {
		return nil, fmt.Errorf("unsupported model type: %s, only openai_compatible is supported", config.ModelType)
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for model %s", config.Name)
	}
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   config.ModelName,
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
	})
}

// InitializeAgent 初始化 Agent（加载 MCP 工具）
func (a *AgentService) InitializeAgent(ctx context.Context) error {
	// 获取 MCP 管理器
	mcpManager := mcpmanager.GetGlobalMCPManager()

	// 获取所有 MCP 工具
	tools, err := mcpManager.GetAllTools(ctx)
	if err != nil {
		mylogger.Logger.Warn("failed to get MCP tools, agent will run without tools",
			zap.Error(err),
		)
		tools = nil
	}

	// 创建 ReAct Agent（系统提示词由 AIHelper 统一提供）
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: a.model,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools,
		},
		MaxStep: a.config.MaxIterations,
	})
	if err != nil {
		return fmt.Errorf("failed to create react agent: %w", err)
	}

	a.agent = agent
	mylogger.Logger.Info("Agent initialized successfully",
		zap.String("modelName", a.modelName),
		zap.String("mode", string(a.config.Mode)),
		zap.Int("toolCount", len(tools)),
	)

	return nil
}

// GenerateResponse 使用 Agent 生成响应
func (a *AgentService) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if a.agent == nil {
		if err := a.InitializeAgent(ctx); err != nil {
			return nil, err
		}
	}

	// RAG 已在 AIHelper 层统一处理，直接调用 agent
	resp, err := a.agent.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("agent generate failed: %w", err)
	}

	return resp, nil
}

// StreamResponse 使用 Agent 生成流式响应
func (a *AgentService) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	if a.agent == nil {
		if err := a.InitializeAgent(ctx); err != nil {
			return "", err
		}
	}

	// RAG 已在 AIHelper 层统一处理，直接调用 agent
	stream, err := a.agent.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("agent stream failed: %w", err)
	}
	defer stream.Close()

	var fullResponse strings.Builder

	for {
		select {
		case <-ctx.Done():
			return fullResponse.String(), ctx.Err()
		default:
			msg, err := stream.Recv()

			if err == io.EOF {
				// 保存 usage（io.EOF 时 msg 可能为 nil，需要检查）
				if msg != nil && msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
					a.lastUsage = msg.ResponseMeta.Usage
				}
				return fullResponse.String(), nil
			}
			if err != nil {
				mylogger.Logger.Error("agent stream recv error", zap.Error(err))
				// 如果已经有部分响应，返回它而不是错误
				if fullResponse.Len() > 0 {
					return fullResponse.String(), nil
				}
				return "", fmt.Errorf("agent stream recv failed: %w", err)
			}

			// 保存 usage
			if msg != nil && msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
				a.lastUsage = msg.ResponseMeta.Usage
			}

			// 处理消息内容
			if msg != nil && len(msg.Content) > 0 {
				fullResponse.WriteString(msg.Content)
				cb(msg.Content)
			}
		}
	}
}

// GetModelType 获取模型类型
func (a *AgentService) GetModelType() string {
	return a.modelName
}

// GetModelName 获取模型显示名称
func (a *AgentService) GetModelName() string {
	return a.modelName
}

// GetLastUsage 获取最近一次流式响应的 token 使用量
func (a *AgentService) GetLastUsage() *schema.TokenUsage {
	return a.lastUsage
}
