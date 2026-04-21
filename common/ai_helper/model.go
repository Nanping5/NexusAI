package aihelper

import (
	"NexusAi/common/rag"
	"NexusAi/model"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type StreamCallback func(msg string)

// UsageCallback Token 使用量回调函数（用于流式响应）
type UsageCallback func(usage *schema.TokenUsage)

type AIModel interface {
	GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
	StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error)
	GetModelType() string
	GetModelName() string
	GetLastUsage() *schema.TokenUsage
}

// EnhanceWithRAG 使用 RAG 检索增强消息上下文（公共函数）
func EnhanceWithRAG(ctx context.Context, sessionID string, messages []*schema.Message) ([]*schema.Message, error) {
	ragQuery, err := rag.NewRAGQuery(ctx, sessionID)
	if err != nil {
		log.Printf("[RAG DEBUG] Failed to create RAG query: %v", err)
		return nil, fmt.Errorf("failed to create RAG query: %v", err)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	var query string
	foundUserMsg := false
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == schema.User {
			query = messages[i].Content
			foundUserMsg = true
			log.Printf("[RAG DEBUG] Found user message at index %d: %s", i, query)
			break
		}
	}

	if !foundUserMsg {
		return nil, fmt.Errorf("no user message found in conversation")
	}

	docs, err := ragQuery.RetrieveDocuments(ctx, query)
	if err != nil {
		log.Printf("[RAG DEBUG] Failed to retrieve documents: %v", err)
		return nil, fmt.Errorf("failed to retrieve documents: %v", err)
	}

	log.Printf("[RAG DEBUG] Retrieved %d documents for session %s", len(docs), sessionID)

	ragPrompt := rag.BuildRagPrompt(query, docs)
	log.Printf("[RAG DEBUG] RAG prompt length: %d", len(ragPrompt))

	ragMessages := make([]*schema.Message, len(messages))
	copy(ragMessages, messages)

	for i := len(ragMessages) - 1; i >= 0; i-- {
		if ragMessages[i].Role == schema.User {
			ragMessages[i] = &schema.Message{
				Role:    schema.User,
				Content: ragPrompt,
			}
			break
		}
	}

	return ragMessages, nil
}

// streamResponseFromModel 通用的流式响应处理逻辑
func streamResponseFromModel(ctx context.Context, chatModel einomodel.ToolCallingChatModel, messages []*schema.Message, cb StreamCallback, usageCb UsageCallback, modelType string) (string, error) {
	stream, err := chatModel.Stream(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("%s stream failed: %v", modelType, err)
	}
	defer stream.Close()

	var fullResponse strings.Builder
	var lastUsage *schema.TokenUsage

	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接或请求被取消
			return "", ctx.Err()
		default:
			msg, err := stream.Recv()

			if err == io.EOF {
				// 流结束时，如果有 usage 信息则触发回调
				if usageCb != nil && lastUsage != nil {
					usageCb(lastUsage)
				}
				return fullResponse.String(), nil
			}
			if err != nil {
				return "", fmt.Errorf("%s stream recv failed: %v", modelType, err)
			}

			// 保留最后的 usage 信息
			if msg != nil && msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
				lastUsage = msg.ResponseMeta.Usage
			}

			if msg != nil && len(msg.Content) > 0 {
				fullResponse.WriteString(msg.Content) // 累积完整响应
				cb(msg.Content)                       // 实时回调当前消息片段
			}
		}
	}
}

// OpenAICompatibleModel OpenAI 兼容模型实现（支持动态配置）
type OpenAICompatibleModel struct {
	llm       einomodel.ToolCallingChatModel
	config    *model.AIModelConfig
	userID    string
	sessionID string
	lastUsage *schema.TokenUsage // 最近一次流式响应的 token 使用量
}

// NewOpenAICompatibleModel 创建 OpenAI 兼容模型实例（动态配置）
func NewOpenAICompatibleModel(ctx context.Context, config *model.AIModelConfig, userID, sessionID string) (*OpenAICompatibleModel, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for model %s", config.Name)
	}

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   config.ModelName,
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create openai compatible model failed: %v", err)
	}
	return &OpenAICompatibleModel{llm: llm, config: config, userID: userID, sessionID: sessionID}, nil
}

// GenerateResponse 生成完整响应（RAG 在 AIHelper 层统一处理）
func (o *OpenAICompatibleModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	resp, err := o.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("model generate failed: %v", err)
	}
	return resp, nil
}

// StreamResponse 实现流式响应（RAG 在 AIHelper 层统一处理）
func (o *OpenAICompatibleModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error) {
	// 使用回调接收 usage 并存储
	usageCb := func(usage *schema.TokenUsage) {
		o.lastUsage = usage
	}
	return streamResponseFromModel(ctx, o.llm, messages, cb, usageCb, o.config.Name)
}

func (o *OpenAICompatibleModel) GetModelType() string {
	return o.config.ModelType
}

// GetModelName 获取模型显示名称
func (o *OpenAICompatibleModel) GetModelName() string {
	return o.config.Name
}

// GetLastUsage 获取最近一次流式响应的 token 使用量
func (o *OpenAICompatibleModel) GetLastUsage() *schema.TokenUsage {
	return o.lastUsage
}
