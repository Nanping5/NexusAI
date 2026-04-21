package translate

import (
	"NexusAi/common/mcp/base"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// AlibabaTranslateClient 阿里云翻译客户端
type AlibabaTranslateClient struct {
	accessKeyID     string
	accessKeySecret string
	httpClient      *http.Client
}

// TranslateResponse 翻译响应
type TranslateResponse struct {
	RequestId string `json:"RequestId"`
	Data      struct {
		Translated       string `json:"Translated"`
		WordCount        string `json:"WordCount"`
		DetectedLanguage string `json:"DetectedLanguage"`
	} `json:"Data"`
	Code    int    `json:"Code"`
	Message string `json:"Message"`
}

// LanguagePair 支持的语言对
var supportedLanguages = map[string]string{
	"zh": "中文",
	"en": "英文",
	"ja": "日语",
	"ko": "韩语",
	"es": "西班牙语",
	"fr": "法语",
	"de": "德语",
	"ru": "俄语",
	"pt": "葡萄牙语",
	"it": "意大利语",
	"ar": "阿拉伯语",
	"th": "泰语",
	"vi": "越南语",
	"id": "印尼语",
	"ms": "马来语",
}

// NewAlibabaTranslateClient 翻译客户端
func NewAlibabaTranslateClient() *AlibabaTranslateClient {
	return &AlibabaTranslateClient{
		accessKeyID:     os.Getenv("ALIBABA_ACCESS_KEY_ID"),
		accessKeySecret: os.Getenv("ALIBABA_ACCESS_KEY_SECRET"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Translate 执行翻译
func (c *AlibabaTranslateClient) Translate(ctx context.Context, sourceText, sourceLanguage, targetLanguage string) (string, error) {
	if c.accessKeyID == "" || c.accessKeySecret == "" {
		return "", fmt.Errorf("ALIBABA_ACCESS_KEY_ID or ALIBABA_ACCESS_KEY_SECRET not configured")
	}

	// 请求参数
	params := map[string]string{
		"Action":           "TranslateGeneral",
		"FormatType":       "text",
		"Scene":            "general",
		"SourceLanguage":   sourceLanguage,
		"SourceText":       sourceText,
		"TargetLanguage":   targetLanguage,
		"Version":          "2018-10-12",
		"Format":           "JSON",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"AccessKeyId":      c.accessKeyID,
	}

	signature := c.generateSignature(params, "GET")
	params["Signature"] = signature

	// URL
	queryString := c.buildQueryString(params)
	apiURL := fmt.Sprintf("https://mt.aliyuncs.com/?%s", queryString)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 解析
	var result struct {
		RequestId string `json:"RequestId"`
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		Data      struct {
			Translated       string `json:"Translated"`
			WordCount        string `json:"WordCount"`
			DetectedLanguage string `json:"DetectedLanguage"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if result.Code != "200" {
		return "", fmt.Errorf("translate failed: code=%s, message=%s", result.Code, result.Message)
	}

	return result.Data.Translated, nil
}

// generateSignature 生成阿里云 API 签名
func (c *AlibabaTranslateClient) generateSignature(params map[string]string, method string) string {
	// 按参数名排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建请求
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", specialURLEncode(k), specialURLEncode(params[k])))
	}
	canonicalizedQueryString := strings.Join(pairs, "&")

	// 构建待签名字符串
	stringToSign := fmt.Sprintf("%s&%%2F&%s", method, specialURLEncode(canonicalizedQueryString))

	// HMAC-SHA1 签名
	mac := hmac.New(sha1.New, []byte(c.accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}

// buildQueryString 构建查询字符串
func (c *AlibabaTranslateClient) buildQueryString(params map[string]string) string {
	var pairs []string
	for k, v := range params {
		pairs = append(pairs, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
	}
	return strings.Join(pairs, "&")
}

// specialURLEncode 阿里要求URL 编码）
func specialURLEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

// DetectLanguage 检测语言
func (c *AlibabaTranslateClient) DetectLanguage(ctx context.Context, text string) (string, error) {
	// 使用 auto 作为源语言，阿里云会自动检测
	params := map[string]string{
		"Action":           "TranslateGeneral",
		"FormatType":       "text",
		"Scene":            "general",
		"SourceLanguage":   "auto",
		"SourceText":       text,
		"TargetLanguage":   "zh", // 翻译成中文以检测源语言
		"Version":          "2018-10-12",
		"Format":           "JSON",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"AccessKeyId":      c.accessKeyID,
	}

	signature := c.generateSignature(params, "GET")
	params["Signature"] = signature

	queryString := c.buildQueryString(params)
	apiURL := fmt.Sprintf("https://mt.cn-hangzhou.aliyuncs.com/?%s", queryString)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var wrapper struct {
		TranslateGeneralResponse TranslateResponse `json:"TranslateGeneralResponse"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return "", err
	}

	return wrapper.TranslateGeneralResponse.Data.DetectedLanguage, nil
}

// GetTranslateServiceConfig 获取翻译服务配置
func GetTranslateServiceConfig() base.ServiceConfig {
	client := NewAlibabaTranslateClient()

	return base.ServiceConfig{
		Name:    "translate",
		Version: "1.0.0",
		Tools: []base.ToolDefinition{
			{
				Name:        "translate_text",
				Description: "翻译文本到指定语言，支持多种语言互译（中、英、日、韩、法、德、西等）。如果不指定源语言，会自动检测。",
				Parameters: []base.ToolParameter{
					{
						Name:        "text",
						Description: "需要翻译的文本内容",
						Required:    true,
					},
					{
						Name:        "target_language",
						Description: "目标语言代码，如：zh(中文)、en(英文)、ja(日语)、ko(韩语)、es(西班牙语)、fr(法语)、de(德语)",
						Required:    true,
					},
					{
						Name:        "source_language",
						Description: "源语言代码，可选。不填则自动检测。如：zh(中文)、en(英文)、ja(日语)",
						Required:    false,
					},
				},
				Handler: func(ctx context.Context, args map[string]any) (string, error) {
					text, ok := args["text"].(string)
					if !ok || text == "" {
						return "", fmt.Errorf("invalid argument: text is required and must be a string")
					}

					targetLang, ok := args["target_language"].(string)
					if !ok || targetLang == "" {
						return "", fmt.Errorf("invalid argument: target_language is required")
					}

					sourceLang := "auto" // 默认自动检测
					if sl, ok := args["source_language"].(string); ok && sl != "" {
						sourceLang = sl
					}

					result, err := client.Translate(ctx, text, sourceLang, targetLang)
					if err != nil {
						return "", fmt.Errorf("translation error: %w", err)
					}

					// 格式化输出
					var output strings.Builder
					output.WriteString(fmt.Sprintf("📝 原文: %s\n", text))
					if sourceLang == "auto" {
						output.WriteString("🔍 已自动检测源语言\n")
					} else {
						if langName, ok := supportedLanguages[sourceLang]; ok {
							output.WriteString(fmt.Sprintf("源语言: %s\n", langName))
						}
					}
					if langName, ok := supportedLanguages[targetLang]; ok {
						output.WriteString(fmt.Sprintf("目标语言: %s\n", langName))
					}
					output.WriteString(fmt.Sprintf("译文: %s", result))

					return output.String(), nil
				},
			},
			{
				Name:        "detect_language",
				Description: "检测文本的语言类型",
				Parameters: []base.ToolParameter{
					{
						Name:        "text",
						Description: "需要检测语言的文本",
						Required:    true,
					},
				},
				Handler: func(ctx context.Context, args map[string]any) (string, error) {
					text, ok := args["text"].(string)
					if !ok || text == "" {
						return "", fmt.Errorf("invalid argument: text is required")
					}

					langCode, err := client.DetectLanguage(ctx, text)
					if err != nil {
						return "", fmt.Errorf("language detection error: %w", err)
					}

					langName, ok := supportedLanguages[langCode]
					if !ok {
						langName = "未知语言"
					}

					return fmt.Sprintf("检测结果: %s (%s)", langName, langCode), nil
				},
			},
		},
	}
}
