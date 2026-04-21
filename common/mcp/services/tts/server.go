package tts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"NexusAi/common/mcp/base"
	"NexusAi/common/tts"
	mylogger "NexusAi/pkg/logger"

	"go.uber.org/zap"
)

// 音频文件保存目录
const audioDir = "../../uploads/audio"

// init 确保音频目录存在
func init() {
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		mylogger.Logger.Error("Failed to create audio directory", zap.Error(err))
	}
}

// saveAudioFile 保存音频文件并返回相对路径
func saveAudioFile(audioData []byte, filename string) (string, error) {
	filePath := filepath.Join(audioDir, filename)
	if err := os.WriteFile(filePath, audioData, 0644); err != nil {
		return "", fmt.Errorf("failed to save audio file: %w", err)
	}
	return filename, nil
}

// GetTTSServiceConfig 获取 TTS 服务配置
func GetTTSServiceConfig() base.ServiceConfig {
	ttsService := tts.GetTTSService()

	return base.ServiceConfig{
		Name:    "tts",
		Version: "1.0.0",
		Tools: []base.ToolDefinition{
			{
				Name: "text_to_speech",
				Description: `[语音合成工具] 将文本转换为可播放的语音文件。

【触发条件】当用户的消息中包含以下任一关键词时，必须调用此工具：
- "朗读"、"播放"、"念"、"读"、"听"
- "朗读给我听"、"念给我听"、"读给我听"
- "语音"、"声音"、"念出来"、"读出来"
- "播放语音"、"语音播放"

【使用方法】
1. 先正常生成文本回复
2. 然后立即调用此工具，将你刚才生成的回复内容作为参数传入
3. 工具会返回音频文件的URL

【重要】这是用户明确要求的功能，必须执行！`,
				Parameters: []base.ToolParameter{
					{
						Name:        "text",
						Description: "要转换为语音的文本内容",
						Required:    true,
					},
				},
				Handler: func(ctx context.Context, args map[string]any) (string, error) {
					text, ok := args["text"].(string)
					if !ok || text == "" {
						return "", fmt.Errorf("invalid argument: text is required and must be a string")
					}

					// 限制文本长度
					if len(text) > 1000 {
						text = text[:1000]
						mylogger.Logger.Warn("Text truncated to 1000 characters")
					}

					mylogger.Logger.Info("TTS tool called",
						zap.String("text", text),
						zap.Int("textLen", len(text)),
					)

					// 调用 TTS 服务合成语音
					audioData, err := ttsService.Synthesize(ctx, text)
					if err != nil {
						mylogger.Logger.Error("TTS synthesis failed", zap.Error(err))
						return "", fmt.Errorf("语音合成失败: %w", err)
					}

					// 生成唯一文件名
					filename := fmt.Sprintf("tts_%d.wav", time.Now().UnixNano())

					// 保存音频文件
					relativePath, err := saveAudioFile(audioData, filename)
					if err != nil {
						return "", err
					}

					mylogger.Logger.Info("TTS audio generated",
						zap.String("path", relativePath),
						zap.Int("size", len(audioData)),
					)

					// 返回 JSON 格式的音频信息，前端可以解析并显示播放器
					return fmt.Sprintf(`{"audio_url": "/static/%s", "text_length": %d}`, relativePath, len(text)), nil
				},
			},
		},
	}
}
