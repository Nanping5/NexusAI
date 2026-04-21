package service

import (
	ret "NexusAi/common/code"
	myredis "NexusAi/common/redis"
	"NexusAi/dao"
	"NexusAi/pkg/utils"
	"context"
	"errors"
)

type userService struct {
}

var UserService = &userService{}

// Register 用户注册（仅支持邮箱）
func (s *userService) Register(ctx context.Context, email, password, captcha string) (string, ret.Code) {
	if email == "" || password == "" || captcha == "" {
		return "", ret.CodeInvalidParams
	}

	// 验证验证码
	valid, err := myredis.CheckCaptchaForEmail(email, captcha)
	if err != nil {
		return "", ret.CodeServerBusy
	}
	if !valid {
		return "", ret.CodeInvalidCaptcha
	}

	userdao := dao.UserDAO

	// 检查邮箱是否已注册
	exist, err := userdao.IsEmailExist(ctx, email)
	if err != nil {
		return "", ret.CodeServerBusy
	}
	if exist {
		return "", ret.CodeUserExist
	}

	user, err := userdao.Register(ctx, email, password)
	if err != nil {
		return "", ret.CodeServerBusy
	}

	// 生成 JWT token（使用 UserID）
	token, err := utils.GenerateJwtToken(user.UserID)
	if err != nil {
		return "", ret.CodeServerBusy
	}

	return token, ret.CodeSuccess
}

// Login 用户登录（仅支持邮箱）
func (s *userService) Login(ctx context.Context, email, password string) (string, ret.Code) {
	if email == "" || password == "" {
		return "", ret.CodeInvalidParams
	}

	user, err := dao.UserDAO.Login(ctx, email, password)
	if err != nil {
		// 使用 errors.Is 进行错误类型判断
		if errors.Is(err, dao.ErrUserNotFound) || errors.Is(err, dao.ErrInvalidPassword) {
			// 用户不存在或密码错误，返回相同错误码（安全考虑，不暴露用户是否存在）
			return "", ret.CodeInvalidPassword
		}
		// 其他错误（数据库错误等）返回服务繁忙
		return "", ret.CodeServerBusy
	}

	// 生成 JWT token（使用 UserID）
	token, err := utils.GenerateJwtToken(user.UserID)
	if err != nil {
		return "", ret.CodeServerBusy
	}

	return token, ret.CodeSuccess
}

func (s *userService) GetUserInfo(ctx context.Context, userID string) (map[string]any, ret.Code) {
	// 先从 Redis 缓存获取
	cachedUser, err := myredis.GetUserInfo(userID)
	if err != nil {
		// Redis 出错继续从数据库查询
	} else if cachedUser != nil {
		return map[string]any{
			"user_id":  cachedUser.UserID,
			"email":    cachedUser.Email,
			"nickname": cachedUser.Nickname,
		}, ret.CodeSuccess
	}

	// 从数据库获取
	user, err := dao.UserDAO.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, ret.CodeUserNotExist
	}

	// 写入 Redis 缓存
	cacheUser := myredis.UserCacheInfo{
		UserID:   user.UserID,
		Email:    user.Email,
		Nickname: user.Nickname,
	}
	if err := myredis.SetUserInfo(userID, cacheUser); err != nil {
		// 缓存失败不影响返回
	}

	return map[string]any{
		"user_id":  user.UserID,
		"email":    user.Email,
		"nickname": user.Nickname,
	}, ret.CodeSuccess
}

// UpdateNickname 更新用户昵称
func (s *userService) UpdateNickname(ctx context.Context, userID, nickname string) ret.Code {
	if nickname == "" {
		return ret.CodeInvalidParams
	}
	if err := dao.UserDAO.UpdateNickname(ctx, userID, nickname); err != nil {
		return ret.CodeServerBusy
	}

	// 使用户信息缓存失效
	myredis.InvalidateUserInfo(userID)

	return ret.CodeSuccess
}
