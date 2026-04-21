package dao

import (
	mmysql "NexusAi/common/mysql"
	"NexusAi/model"
	"NexusAi/pkg/utils"
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// 定义明确的错误类型
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
)

// UserDAO 全局用户 DAO 实例
var UserDAO = &userDAO{}

type userDAO struct {
}

// isDuplicateKeyError 判断是否为 MySQL 唯一键冲突错误
func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 // ER_DUP_ENTRY
	}
	return false
}

// Register 注册新用户（使用直接插入+唯一键冲突重试，避免竞态）
func (dao *userDAO) Register(ctx context.Context, email, password string) (*model.User, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}

	nickname := utils.GenerateNickname()
	const maxAttempts = 10

	for i := 0; i < maxAttempts; i++ {
		userID := utils.GenerateUserID()
		user := model.User{
			UserID:   userID,
			Email:    email,
			Password: password,
			Nickname: nickname,
		}
		if err := user.EncryptPassword(); err != nil {
			return nil, fmt.Errorf("encrypt password failed: %w", err)
		}
		if err := db.WithContext(ctx).Create(&user).Error; err != nil {
			if isDuplicateKeyError(err) {
				continue // 唯一键冲突，重试
			}
			return nil, fmt.Errorf("create user failed: %w", err)
		}
		return &user, nil
	}
	return nil, fmt.Errorf("failed to generate unique user id after %d attempts", maxAttempts)
}

// IsEmailExist 检查邮箱是否已存在
func (dao *userDAO) IsEmailExist(ctx context.Context, email string) (bool, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return false, fmt.Errorf("get db client failed: %w", err)
	}
	var count int64
	if err := db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check email existence failed: %w", err)
	}
	return count > 0, nil
}

// Login 邮箱登录
func (dao *userDAO) Login(ctx context.Context, email, password string) (*model.User, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}
	user := model.User{}
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query user failed: %w", err)
	}
	if !user.CheckPassword(password) {
		return nil, ErrInvalidPassword
	}
	return &user, nil
}

// GetUserInfo 获取用户信息
func (dao *userDAO) GetUserInfo(ctx context.Context, userID string) (*model.User, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}
	user := model.User{}
	if err := db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query user failed: %w", err)
	}
	return &user, nil
}

// UpdateNickname 更新用户昵称
func (dao *userDAO) UpdateNickname(ctx context.Context, userID, nickname string) error {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return fmt.Errorf("get db client failed: %w", err)
	}
	if err := db.Model(&model.User{}).Where("user_id = ?", userID).Update("nickname", nickname).Error; err != nil {
		return fmt.Errorf("update nickname failed: %w", err)
	}
	return nil
}
