package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MessageEmbedding 向量嵌入接口
type MessageEmbedding interface {
	EmbedStrings(ctx context.Context, texts []string) ([][]float32, error)
}

// DashScopeEmbedder DashScope 向量生成器实现（OpenAI兼容格式）
type DashScopeEmbedder struct {
	apiKey     string
	baseURL   string
	model     string
	dimension int
}

// NewDashScopeEmbedder 创建 DashScope 向量生成器
func NewDashScopeEmbedder(ctx context.Context, baseURL, apiKey, model string, dimension int) (MessageEmbedding, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("api key is empty")
	}

	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}

	return &DashScopeEmbedder{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		dimension:  dimension,
	}, nil
}

// EmbedStrings 生成文本向量
func (d *DashScopeEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// 过滤空字符串
	var validTexts []string
	for _, t := range texts {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			validTexts = append(validTexts, trimmed)
		}
	}

	if len(validTexts) == 0 {
		return nil, nil
	}

	// 构建请求（OpenAI兼容格式）
	reqBody := map[string]any{
		"model":           d.model,
		"input":            validTexts,
		"encoding_format": "float",
	}
	if d.dimension > 0 {
		reqBody["dimensions"] = d.dimension
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", d.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dashscope api error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int        `json:"index"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	embeddings := make([][]float32, len(result.Data))
	for i, item := range result.Data {
		embeddings[i] = item.Embedding
	}

	return embeddings, nil
}
