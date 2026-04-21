package utils

import (
	"NexusAi/model"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// 默认系统提示词
const defaultSystemPrompt = "You are a helpful, friendly, and knowledgeable AI assistant. You provide clear, accurate, and thoughtful responses. You communicate in a natural and engaging way."

// GetSystemPrompt 获取系统提示词（优先从环境变量读取，否则使用默认值）
func GetSystemPrompt() string {
	if prompt := os.Getenv("DEFAULT_SYSTEM_PROMPT"); prompt != "" {
		return prompt
	}
	return defaultSystemPrompt
}

// MD5 MD5加密
func MD5(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	return hex.EncodeToString(m.Sum(nil))
}

// 将 schema 消息转换为数据库可存储的格式
func ConvertToModelMessage(sessionID string, userId string, msg *schema.Message) *model.Message {
	return &model.Message{
		SessionID: sessionID,
		UserID:    userId,
		Content:   msg.Content,
	}
}

// 将数据库消息转换为 schema 消息（供 AI 使用），自动添加系统提示词
func ConvertToSchemaMessages(msgs []*model.Message) []*schema.Message {
	schemaMsgs := make([]*schema.Message, 0, len(msgs)+1)

	// 添加系统提示词作为第一条消息
	schemaMsgs = append(schemaMsgs, &schema.Message{
		Role:    schema.System,
		Content: GetSystemPrompt(),
	})

	for _, m := range msgs {
		role := schema.Assistant
		if m.IsUser {
			role = schema.User
		}
		schemaMsgs = append(schemaMsgs, &schema.Message{
			Role:    role,
			Content: m.Content,
		})
	}
	return schemaMsgs
}

func ValidateFile(file *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".txt", ".pdf", ".md"}
	if !slices.Contains(allowedExts, ext) {
		return fmt.Errorf("invalid file type")
	}
	if file.Size > 20<<20 { // 20MB
		return fmt.Errorf("file too large")
	}
	return nil

}

// RemoveAllFilesInDir 删除目录中的所有文件（不删除子目录）
func RemoveAllFilesInDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 目录不存在就算了
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return err
			}
		}
	}
	return nil
}
