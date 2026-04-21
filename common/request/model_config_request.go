package request

// CreateModelConfigRequest 创建模型配置请求
type CreateModelConfigRequest struct {
	Name        string `json:"name" binding:"required"`
	ModelType   string `json:"model_type" binding:"required,oneof=openai_compatible"`
	BaseURL     string `json:"base_url" binding:"required"`
	ModelName   string `json:"model_name" binding:"required"`
	APIKey      string `json:"api_key"`
	Description string `json:"description"`
	IsEnabled   bool   `json:"is_enabled"`
	IsDefault   bool   `json:"is_default"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateModelConfigRequest 更新模型配置请求
type UpdateModelConfigRequest struct {
	Name        string `json:"name"`
	ModelType   string `json:"model_type" binding:"omitempty,oneof=openai_compatible"`
	BaseURL     string `json:"base_url"`
	ModelName   string `json:"model_name"`
	APIKey      string `json:"api_key"` // 空表示不更新
	Description string `json:"description"`
	IsEnabled   *bool  `json:"is_enabled"`
	IsDefault   *bool  `json:"is_default"`
	SortOrder   *int   `json:"sort_order"`
}

// ToggleModelConfigRequest 启用/禁用模型配置请求
type ToggleModelConfigRequest struct {
	IsEnabled bool `json:"is_enabled"`
}
