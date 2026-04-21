package aihelper

import (
	"context"
	"fmt"
	"sync"

	"NexusAi/model"
)

// DynamicModelCreater 动态模型创建函数，接收配置参数
type DynamicModelCreater func(ctx context.Context, config *model.AIModelConfig, userID, sessionID string) (AIModel, error)

type AiModelFactory struct {
	dynamicCreators map[string]DynamicModelCreater // 动态模型创建器
}

var (
	globalFactory *AiModelFactory
	factoryOnce   sync.Once
)

// GetGlobalFactory 获取全局 AI 模型工厂实例，单例模式（线程安全）
func GetGlobalFactory() *AiModelFactory {
	factoryOnce.Do(func() {
		globalFactory = &AiModelFactory{
			dynamicCreators: make(map[string]DynamicModelCreater),
		}
		globalFactory.registerDynamicCreaters()
	})
	return globalFactory
}

// registerDynamicCreaters 注册动态模型创建器
func (f *AiModelFactory) registerDynamicCreaters() {
	f.dynamicCreators["openai_compatible"] = func(ctx context.Context, config *model.AIModelConfig, userID, sessionID string) (AIModel, error) {
		return NewOpenAICompatibleModel(ctx, config, userID, sessionID)
	}
}

// CreateAIModelByConfigID 根据配置 ID 创建模型
func (f *AiModelFactory) CreateAIModelByConfigID(ctx context.Context, configID, userID, sessionID string) (AIModel, error) {
	// 获取配置
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

	return f.CreateAIModelFromConfig(ctx, config, userID, sessionID)
}

// CreateAIModelFromConfig 根据配置创建模型
func (f *AiModelFactory) CreateAIModelFromConfig(ctx context.Context, config *model.AIModelConfig, userID, sessionID string) (AIModel, error) {
	creater, exists := f.dynamicCreators[config.ModelType]
	if !exists {
		return nil, fmt.Errorf("unsupported model type: %s", config.ModelType)
	}
	return creater(ctx, config, userID, sessionID)
}

// CreateAIHelperByConfigID 根据配置 ID 创建 AIHelper 实例
func (f *AiModelFactory) CreateAIHelperByConfigID(ctx context.Context, configID, sessionID, userID string) (*AIHelper, error) {
	aiModel, err := f.CreateAIModelByConfigID(ctx, configID, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return NewAIHelper(aiModel, sessionID), nil
}

// CreateAIHelperWithAgentByConfigID 根据配置 ID 创建带 Agent 的 AIHelper 实例
func (f *AiModelFactory) CreateAIHelperWithAgentByConfigID(ctx context.Context, configID, sessionID, userID string, agentConfig AgentConfig) (*AIHelper, error) {
	agentService, err := NewAgentServiceByConfigID(ctx, configID, userID, sessionID, agentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent service: %w", err)
	}

	// 初始化 Agent（加载 MCP 工具）
	if err := agentService.InitializeAgent(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize agent: %w", err)
	}

	return NewAIHelperWithAgent(agentService, sessionID, AgentModeReAct), nil
}

// RegisterDynamicModel 注册动态模型创建函数
func (f *AiModelFactory) RegisterDynamicModel(modelType string, creater DynamicModelCreater) {
	f.dynamicCreators[modelType] = creater
}
