package model

import "gorm.io/gorm"

// AIModelConfig AI模型配置表
type AIModelConfig struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigID    string `gorm:"uniqueIndex;type:varchar(36);not null" json:"config_id"` // UUID，对外暴露的ID
	Name        string `gorm:"type:varchar(100);not null" json:"name"`                  // 显示名称，如 "GPT-4o"
	ModelType   string `gorm:"type:varchar(50);not null" json:"model_type"`             // 模型类型：openai_compatible, ollama
	BaseURL     string `gorm:"type:varchar(255);not null" json:"base_url"`              // API 请求地址
	ModelName   string `gorm:"type:varchar(100);not null" json:"model_name"`            // 模型名称，如 "gpt-4o"
	APIKey      string `gorm:"type:varchar(255)" json:"-"`                              // API Key（敏感信息，不返回给前端）
	Description string `gorm:"type:text" json:"description"`                             // 模型描述
	IsEnabled   bool   `gorm:"default:true;not null" json:"is_enabled"`                 // 是否启用
	IsDefault   bool   `gorm:"default:false;not null" json:"is_default"`                // 是否为默认模型
	SortOrder   int    `gorm:"default:0;not null" json:"sort_order"`                    // 排序权重
	*gorm.Model
}

// TableName 指定表名
func (AIModelConfig) TableName() string {
	return "ai_model_configs"
}

// AIModelConfigResponse 返回给前端的模型配置（不包含敏感信息）
type AIModelConfigResponse struct {
	ConfigID    string `json:"config_id"`
	Name        string `json:"name"`
	ModelType   string `json:"model_type"`
	BaseURL     string `json:"base_url"`
	ModelName   string `json:"model_name"`
	Description string `json:"description"`
	IsEnabled   bool   `json:"is_enabled"`
	IsDefault   bool   `json:"is_default"`
	SortOrder   int    `json:"sort_order"`
}

// ToResponse 转换为响应结构
func (c *AIModelConfig) ToResponse() AIModelConfigResponse {
	return AIModelConfigResponse{
		ConfigID:    c.ConfigID,
		Name:        c.Name,
		ModelType:   c.ModelType,
		BaseURL:     c.BaseURL,
		ModelName:   c.ModelName,
		Description: c.Description,
		IsEnabled:   c.IsEnabled,
		IsDefault:   c.IsDefault,
		SortOrder:   c.SortOrder,
	}
}

// AIModelConfigBrief 简要模型信息（供用户选择）
type AIModelConfigBrief struct {
	ConfigID  string `json:"config_id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// ToBrief 转换为简要信息
func (c *AIModelConfig) ToBrief() AIModelConfigBrief {
	return AIModelConfigBrief{
		ConfigID:  c.ConfigID,
		Name:      c.Name,
		IsDefault: c.IsDefault,
	}
}
