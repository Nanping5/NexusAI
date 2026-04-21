package dao

import (
	"context"
	"errors"
	"fmt"

	mmysql "NexusAi/common/mysql"
	"NexusAi/model"
	"NexusAi/pkg/utils"

	"gorm.io/gorm"
)

// AdminDAO 全局管理员 DAO 实例
var AdminDAO = &adminDAO{}

type adminDAO struct{}

// Create 创建管理员
func (dao *adminDAO) Create(ctx context.Context, admin *model.Admin) error {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return fmt.Errorf("get db client failed: %w", err)
	}

	// 生成唯一的 AdminID
	adminID, err := dao.generateUniqueAdminID(db)
	if err != nil {
		return fmt.Errorf("generate admin id failed: %w", err)
	}
	admin.AdminID = adminID

	if err := admin.EncryptPassword(); err != nil {
		return fmt.Errorf("encrypt password failed: %w", err)
	}

	return db.Create(admin).Error
}

// generateUniqueAdminID 生成唯一的管理员 ID (A + 6位数字)
func (dao *adminDAO) generateUniqueAdminID(db *gorm.DB) (string, error) {
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		adminID := "A" + utils.GenerateRandomDigits(6)

		var count int64
		if err := db.Model(&model.Admin{}).Where("admin_id = ?", adminID).Count(&count).Error; err != nil {
			return "", fmt.Errorf("check admin id existence failed: %w", err)
		}
		if count == 0 {
			return adminID, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique admin id after %d attempts", maxAttempts)
}

// GetByUsername 根据用户名获取管理员
func (dao *adminDAO) GetByUsername(ctx context.Context, username string) (*model.Admin, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}

	var admin model.Admin
	err = db.Where("username = ?", username).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("query admin failed: %w", err)
	}
	return &admin, nil
}

// GetByAdminID 根据 AdminID 获取管理员
func (dao *adminDAO) GetByAdminID(ctx context.Context, adminID string) (*model.Admin, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}

	var admin model.Admin
	err = db.Where("admin_id = ?", adminID).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("query admin failed: %w", err)
	}
	return &admin, nil
}

// IsUsernameExist 检查用户名是否已存在
func (dao *adminDAO) IsUsernameExist(ctx context.Context, username string) (bool, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return false, fmt.Errorf("get db client failed: %w", err)
	}

	var count int64
	if err := db.Model(&model.Admin{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check username existence failed: %w", err)
	}
	return count > 0, nil
}
