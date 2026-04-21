package search

import (
	"NexusAi/common/mcp/base"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/serpapi/serpapi-golang"
)

type SearchResult struct {
	Query       string
	Results     []SearchItem
	RelatedSearches []string // 相关搜索建议
	TotalResults string     // 总结果数
}

type SearchItem struct {
	Title       string
	Description string
	URL         string
	Source      string // 来源网站
	Date        string // 发布日期（如果有）
}

type SerpAPIClient struct {
	apiKey string
	engine string
}

func NewSerpAPIClient() *SerpAPIClient {
	apiKey := os.Getenv("SERPAPI_KEY")
	engine := os.Getenv("SEARCH_ENGINE")
	if engine == "" {
		engine = "baidu" // 默认使用百度
	}

	return &SerpAPIClient{
		apiKey: apiKey,
		engine: engine,
	}
}

func (c *SerpAPIClient) Search(ctx context.Context, query string, limit int) (*SearchResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("SERPAPI_KEY not configured, please set SERPAPI_KEY environment variable")
	}

	// 创建 SerpAPI 客户端
	setting := serpapi.NewSerpApiClientSetting(c.apiKey)
	setting.Engine = c.engine

	client := serpapi.NewClient(setting)

	// 设置搜索参数
	parameter := map[string]string{
		"q":        query,
		"no_cache": "true", // 禁用缓存，获取实时结果
	}

	// 参数
	if c.engine == "baidu" {
		parameter["ct"] = "1"  // 限制简体中文结果
		parameter["rn"] = "10" // 返回结果数量
	}

	// 执行搜索
	results, err := client.Search(parameter)
	if err != nil {
		return nil, fmt.Errorf("serpapi search failed: %w", err)
	}

	// 解析结果
	result := &SearchResult{
		Query:   query,
		Results: make([]SearchItem, 0, limit),
	}

	// 提取有机搜索结果
	if organicResults, ok := results["organic_results"].([]interface{}); ok {
		count := 0
		for _, item := range organicResults {
			if count >= limit {
				break
			}

			if resultMap, ok := item.(map[string]interface{}); ok {
				searchItem := SearchItem{}

				if title, ok := resultMap["title"].(string); ok {
					searchItem.Title = title
				}
				if snippet, ok := resultMap["snippet"].(string); ok {
					searchItem.Description = snippet
				}
				if link, ok := resultMap["link"].(string); ok {
					searchItem.URL = link
				}
				// 提取来源网站
				if source, ok := resultMap["source"].(string); ok {
					searchItem.Source = source
				} else if displayedLink, ok := resultMap["displayed_link"].(string); ok {
					searchItem.Source = displayedLink
				}
				// 提取发布日期
				if date, ok := resultMap["date"].(string); ok {
					searchItem.Date = date
				}

				if searchItem.Title != "" {
					result.Results = append(result.Results, searchItem)
					count++
				}
			}
		}
	}

	// 提取总结果数
	if searchInfo, ok := results["search_information"].(map[string]interface{}); ok {
		if totalResults, ok := searchInfo["total_results"].(string); ok {
			result.TotalResults = totalResults
		} else if totalResults, ok := searchInfo["total_results"].(float64); ok {
			result.TotalResults = fmt.Sprintf("%.0f", totalResults)
		}
	}

	// 提取相关搜索建议
	if relatedSearches, ok := results["related_searches"].([]interface{}); ok {
		for _, rs := range relatedSearches {
			if rsMap, ok := rs.(map[string]interface{}); ok {
				if query, ok := rsMap["query"].(string); ok {
					result.RelatedSearches = append(result.RelatedSearches, query)
				}
			}
		}
	}

	// 提取答案框（如果有）
	if answerBox, ok := results["answer_box"].(map[string]interface{}); ok {
		answerItem := SearchItem{
			Title:       "快速答案",
			Description: "",
			URL:         "",
		}

		if answer, ok := answerBox["answer"].(string); ok {
			answerItem.Description = answer
		} else if snippet, ok := answerBox["snippet"].(string); ok {
			answerItem.Description = snippet
		}

		if answerItem.Description != "" {
			// 将答案框放在最前面
			result.Results = append([]SearchItem{answerItem}, result.Results...)
		}
	}

	return result, nil
}

// FormatSearchResult 格式化搜索结果为文本输出
func FormatSearchResult(result *SearchResult) string {
	if result == nil || len(result.Results) == 0 {
		return fmt.Sprintf("没有找到关于 \"%s\" 的相关搜索结果。", result.Query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 搜索: %s\n", result.Query))
	if result.TotalResults != "" {
		sb.WriteString(fmt.Sprintf("📊 约找到 %s 个结果\n", result.TotalResults))
	}
	sb.WriteString("\n")

	for i, item := range result.Results {
		sb.WriteString(fmt.Sprintf("【%d】%s\n", i+1, item.Title))
		if item.Source != "" {
			sb.WriteString(fmt.Sprintf("   来源: %s\n", item.Source))
		}
		if item.Date != "" {
			sb.WriteString(fmt.Sprintf("   日期: %s\n", item.Date))
		}
		if item.Description != "" {
			sb.WriteString(fmt.Sprintf("   摘要: %s\n", item.Description))
		}
		if item.URL != "" {
			sb.WriteString(fmt.Sprintf("   链接: %s\n", item.URL))
		}
		sb.WriteString("\n")
	}

	// 添加相关搜索建议
	if len(result.RelatedSearches) > 0 {
		sb.WriteString("📌 相关搜索:\n")
		for i, query := range result.RelatedSearches {
			if i >= 5 {
				break // 最多显示5条相关搜索
			}
			sb.WriteString(fmt.Sprintf("   • %s\n", query))
		}
	}

	return sb.String()
}

// GetSearchServiceConfig 获取搜索服务配置
func GetSearchServiceConfig() base.ServiceConfig {
	client := NewSerpAPIClient()

	return base.ServiceConfig{
		Name:    "search",
		Version: "1.0.0",
		Tools: []base.ToolDefinition{
			{
				Name:        "web_search",
				Description: "使用搜索引擎搜索互联网信息，返回相关网页结果",
				Parameters: []base.ToolParameter{
					{
						Name:        "query",
						Description: "搜索关键词或问题",
						Required:    true,
					},
					{
						Name:        "limit",
						Description: "返回结果数量，默认5条",
						Required:    false,
					},
				},
				Handler: func(ctx context.Context, args map[string]any) (string, error) {
					query, ok := args["query"].(string)
					if !ok || query == "" {
						return "", fmt.Errorf("invalid argument: query is required and must be a string")
					}

					limit := 5
					if limitVal, ok := args["limit"].(float64); ok && limitVal > 0 {
						limit = int(limitVal)
					}

					result, err := client.Search(ctx, query, limit)
					if err != nil {
						return "", fmt.Errorf("search error: %w", err)
					}

					return FormatSearchResult(result), nil
				},
			},
		},
	}
}
