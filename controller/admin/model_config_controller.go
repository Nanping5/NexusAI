package adminController

import (
	ret "NexusAi/common/code"
	"NexusAi/common/request"
	response "NexusAi/common/response/common"
	"NexusAi/model"
	"NexusAi/service"

	"github.com/gin-gonic/gin"
)

// GetAllModelConfigs 获取所有模型配置
func GetAllModelConfigs(c *gin.Context) {
	configs, err := service.ModelConfigService.GetAllModelConfigs(c.Request.Context())
	if err != nil {
		response.Fail(c, ret.CodeServerBusy)
		return
	}

	// 转换为响应格式（不包含 API Key）
	responses := make([]model.AIModelConfigResponse, 0, len(configs))
	for _, config := range configs {
		responses = append(responses, config.ToResponse())
	}

	response.Success(c, gin.H{"models": responses})
}

// GetModelConfig 获取单个模型配置
func GetModelConfig(c *gin.Context) {
	configID := c.Param("config_id")
	if configID == "" {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	config, err := service.ModelConfigService.GetModelConfig(c.Request.Context(), configID)
	if err != nil {
		response.Fail(c, ret.CodeServerBusy)
		return
	}
	if config == nil {
		response.Fail(c, ret.CodeRecordNotFound)
		return
	}

	response.Success(c, config.ToResponse())
}

// CreateModelConfig 创建模型配置
func CreateModelConfig(c *gin.Context) {
	var req request.CreateModelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	config := &model.AIModelConfig{
		Name:        req.Name,
		ModelType:   req.ModelType,
		BaseURL:     req.BaseURL,
		ModelName:   req.ModelName,
		APIKey:      req.APIKey,
		Description: req.Description,
		IsEnabled:   req.IsEnabled,
		IsDefault:   req.IsDefault,
		SortOrder:   req.SortOrder,
	}

	created, err := service.ModelConfigService.CreateModelConfig(c.Request.Context(), config)
	if err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	response.Success(c, gin.H{"config_id": created.ConfigID})
}

// UpdateModelConfig 更新模型配置
func UpdateModelConfig(c *gin.Context) {
	configID := c.Param("config_id")
	if configID == "" {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	var req request.UpdateModelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	updates := make(map[string]any)
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ModelType != "" {
		updates["model_type"] = req.ModelType
	}
	if req.BaseURL != "" {
		updates["base_url"] = req.BaseURL
	}
	if req.ModelName != "" {
		updates["model_name"] = req.ModelName
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if err := service.ModelConfigService.UpdateModelConfig(c.Request.Context(), configID, updates); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	response.Success(c, nil)
}

// DeleteModelConfig 删除模型配置
func DeleteModelConfig(c *gin.Context) {
	configID := c.Param("config_id")
	if configID == "" {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	if err := service.ModelConfigService.DeleteModelConfig(c.Request.Context(), configID); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	response.Success(c, nil)
}

// SetDefaultModelConfig 设置默认模型
func SetDefaultModelConfig(c *gin.Context) {
	configID := c.Param("config_id")
	if configID == "" {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	if err := service.ModelConfigService.SetDefaultModelConfig(c.Request.Context(), configID); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	response.Success(c, nil)
}

// ToggleModelConfig 启用/禁用模型配置
func ToggleModelConfig(c *gin.Context) {
	configID := c.Param("config_id")
	if configID == "" {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	var req request.ToggleModelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	updates := map[string]any{
		"is_enabled": req.IsEnabled,
	}

	if err := service.ModelConfigService.UpdateModelConfig(c.Request.Context(), configID, updates); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	response.Success(c, nil)
}
