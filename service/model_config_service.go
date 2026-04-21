package service

import (
	"context"
	"errors"

	aihelper "NexusAi/common/ai_helper"
	"NexusAi/dao"
	"NexusAi/model"
	"NexusAi/pkg/utils"

	"gorm.io/gorm"
)

type modelConfigService struct{}

// ModelConfigService 模型配置服务实例
var ModelConfigService = &modelConfigService{}

// CreateModelConfig 创建模型配置
func (s *modelConfigService) CreateModelConfig(ctx context.Context, req *model.AIModelConfig) (*model.AIModelConfig, error) {
	// 检查名称是否已存在
	exists, err := dao.AIModelConfigDAO.IsNameExist(ctx, req.Name, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("模型名称已存在")
	}

	// 生成配置 ID
	req.ConfigID = utils.GenerateUUID()

	// 如果设置为默认模型，先清除其他默认标记
	if req.IsDefault {
		if err := dao.AIModelConfigDAO.ClearDefault(ctx); err != nil {
			return nil, err
		}
	}

	if err := dao.AIModelConfigDAO.Create(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

// UpdateModelConfig 更新模型配置
func (s *modelConfigService) UpdateModelConfig(ctx context.Context, configID string, updates map[string]interface{}) error {
	// 检查配置是否存在
	config, err := dao.AIModelConfigDAO.GetByConfigID(ctx, configID)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("模型配置不存在")
	}

	// 如果更新名称，检查是否重复
	if name, ok := updates["name"].(string); ok && name != config.Name {
		exists, err := dao.AIModelConfigDAO.IsNameExist(ctx, name, configID)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("模型名称已存在")
		}
	}

	// 如果设置为默认模型，先清除其他默认标记
	if isDefault, ok := updates["is_default"].(bool); ok && isDefault {
		if err := dao.AIModelConfigDAO.ClearDefault(ctx); err != nil {
			return err
		}
	}

	if err := dao.AIModelConfigDAO.Update(ctx, configID, updates); err != nil {
		return err
	}

	// 清除缓存
	aihelper.InvalidateConfigCache(configID)

	return nil
}

// DeleteModelConfig 删除模型配置
func (s *modelConfigService) DeleteModelConfig(ctx context.Context, configID string) error {
	// 检查配置是否存在
	config, err := dao.AIModelConfigDAO.GetByConfigID(ctx, configID)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("模型配置不存在")
	}

	if err := dao.AIModelConfigDAO.Delete(ctx, configID); err != nil {
		return err
	}

	// 清除缓存
	aihelper.InvalidateConfigCache(configID)

	return nil
}

// GetModelConfig 获取单个模型配置
func (s *modelConfigService) GetModelConfig(ctx context.Context, configID string) (*model.AIModelConfig, error) {
	return dao.AIModelConfigDAO.GetByConfigID(ctx, configID)
}

// GetAllModelConfigs 获取所有模型配置
func (s *modelConfigService) GetAllModelConfigs(ctx context.Context) ([]model.AIModelConfig, error) {
	return dao.AIModelConfigDAO.GetAll(ctx)
}

// GetEnabledModelConfigs 获取所有启用的模型配置
func (s *modelConfigService) GetEnabledModelConfigs(ctx context.Context) ([]model.AIModelConfig, error) {
	return dao.AIModelConfigDAO.GetEnabled(ctx)
}

// SetDefaultModelConfig 设置默认模型
func (s *modelConfigService) SetDefaultModelConfig(ctx context.Context, configID string) error {
	// 检查配置是否存在
	config, err := dao.AIModelConfigDAO.GetByConfigID(ctx, configID)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("模型配置不存在")
	}

	if !config.IsEnabled {
		return errors.New("无法将禁用的模型设置为默认模型")
	}

	if err := dao.AIModelConfigDAO.SetDefault(ctx, configID); err != nil {
		return err
	}

	// 清除缓存
	aihelper.InvalidateConfigCache(configID)

	return nil
}

// GetDefaultModelConfig 获取默认模型配置
func (s *modelConfigService) GetDefaultModelConfig(ctx context.Context) (*model.AIModelConfig, error) {
	return dao.AIModelConfigDAO.GetDefault(ctx)
}

// GetAvailableModelsForUser 获取用户可用的模型列表（简要信息）
func (s *modelConfigService) GetAvailableModelsForUser(ctx context.Context) ([]model.AIModelConfigBrief, error) {
	configs, err := dao.AIModelConfigDAO.GetEnabled(ctx)
	if err != nil {
		return nil, err
	}

	briefs := make([]model.AIModelConfigBrief, 0, len(configs))
	for _, config := range configs {
		briefs = append(briefs, config.ToBrief())
	}

	return briefs, nil
}

// EnsureDefaultModel 确保至少有一个默认模型（如果没有则设置第一个启用的模型为默认）
func (s *modelConfigService) EnsureDefaultModel(ctx context.Context) error {
	defaultConfig, err := dao.AIModelConfigDAO.GetDefault(ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 如果已有默认模型，无需处理
	if defaultConfig != nil {
		return nil
	}

	// 获取第一个启用的模型
	configs, err := dao.AIModelConfigDAO.GetEnabled(ctx)
	if err != nil {
		return err
	}

	if len(configs) > 0 {
		return dao.AIModelConfigDAO.SetDefault(ctx, configs[0].ConfigID)
	}

	return nil
}
