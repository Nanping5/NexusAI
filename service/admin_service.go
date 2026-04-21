package service

import (
	"context"
	"errors"

	"NexusAi/dao"
	"NexusAi/model"
	"NexusAi/pkg/utils"
)

type adminService struct{}

// AdminService 管理员服务实例
var AdminService = &adminService{}

// Login 管理员登录
func (s *adminService) Login(ctx context.Context, username, password string) (*model.Admin, error) {
	admin, err := dao.AdminDAO.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, errors.New("用户名或密码错误")
	}

	if !admin.CheckPassword(password) {
		return nil, errors.New("用户名或密码错误")
	}

	return admin, nil
}

// CreateAdmin 创建管理员
func (s *adminService) CreateAdmin(ctx context.Context, username, password string) (*model.Admin, error) {
	// 检查用户名是否已存在
	exists, err := dao.AdminDAO.IsUsernameExist(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	admin := &model.Admin{
		Username: username,
		Password: password,
	}

	if err := dao.AdminDAO.Create(ctx, admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// GenerateToken 生成管理员 Token
func (s *adminService) GenerateToken(admin *model.Admin) (string, error) {
	return utils.GenerateAdminJwtToken(admin.AdminID, admin.Username)
}
