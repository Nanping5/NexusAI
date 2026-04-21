package dao

import (
	"context"
	"errors"
	"fmt"

	mmysql "NexusAi/common/mysql"
	"NexusAi/model"

	"gorm.io/gorm"
)

// AIModelConfigDAO 全局模型配置 DAO 实例
var AIModelConfigDAO = &aiModelConfigDAO{}

type aiModelConfigDAO struct{}

// Create 创建模型配置
func (dao *aiModelConfigDAO) Create(ctx context.Context, config *model.AIModelConfig) error {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return fmt.Errorf("get db client failed: %w", err)
	}
	return db.Create(config).Error
}

// GetByConfigID 根据 ConfigID 获取模型配置
func (dao *aiModelConfigDAO) GetByConfigID(ctx context.Context, configID string) (*model.AIModelConfig, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}

	var config model.AIModelConfig
	err = db.Where("config_id = ?", configID).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("query model config failed: %w", err)
	}
	return &config, nil
}

// GetAll 获取所有模型配置
func (dao *aiModelConfigDAO) GetAll(ctx context.Context) ([]model.AIModelConfig, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}

	var configs []model.AIModelConfig
	err = db.Order("sort_order ASC, created_at ASC").Find(&configs).Error
	return configs, err
}

// GetEnabled 获取所有启用的模型配置
func (dao *aiModelConfigDAO) GetEnabled(ctx context.Context) ([]model.AIModelConfig, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}

	var configs []model.AIModelConfig
	err = db.Where("is_enabled = ?", true).Order("sort_order ASC, created_at ASC").Find(&configs).Error
	return configs, err
}

// GetDefault 获取默认模型配置
func (dao *aiModelConfigDAO) GetDefault(ctx context.Context) (*model.AIModelConfig, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}

	var config model.AIModelConfig
	err = db.Where("is_enabled = ? AND is_default = ?", true, true).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("query default model config failed: %w", err)
	}
	return &config, nil
}

// Update 更新模型配置
func (dao *aiModelConfigDAO) Update(ctx context.Context, configID string, updates map[string]interface{}) error {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return fmt.Errorf("get db client failed: %w", err)
	}
	return db.Model(&model.AIModelConfig{}).Where("config_id = ?", configID).Updates(updates).Error
}

// Delete 删除模型配置
func (dao *aiModelConfigDAO) Delete(ctx context.Context, configID string) error {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return fmt.Errorf("get db client failed: %w", err)
	}
	return db.Where("config_id = ?", configID).Delete(&model.AIModelConfig{}).Error
}

// ClearDefault 清除所有默认标记
func (dao *aiModelConfigDAO) ClearDefault(ctx context.Context) error {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return fmt.Errorf("get db client failed: %w", err)
	}
	return db.Model(&model.AIModelConfig{}).Where("is_default = ?", true).Update("is_default", false).Error
}

// SetDefault 设置默认模型
func (dao *aiModelConfigDAO) SetDefault(ctx context.Context, configID string) error {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return fmt.Errorf("get db client failed: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 先清除所有默认标记
		if err := tx.Model(&model.AIModelConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		// 设置新的默认模型
		return tx.Model(&model.AIModelConfig{}).Where("config_id = ?", configID).Update("is_default", true).Error
	})
}

// IsNameExist 检查模型名称是否已存在
func (dao *aiModelConfigDAO) IsNameExist(ctx context.Context, name string, excludeConfigID string) (bool, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return false, fmt.Errorf("get db client failed: %w", err)
	}

	query := db.Model(&model.AIModelConfig{}).Where("name = ?", name)
	if excludeConfigID != "" {
		query = query.Where("config_id != ?", excludeConfigID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("check name existence failed: %w", err)
	}
	return count > 0, nil
}
