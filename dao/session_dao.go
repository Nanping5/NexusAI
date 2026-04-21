package dao

import (
	mmysql "NexusAi/common/mysql"
	"NexusAi/model"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type sessionDAO struct {
}

var SessionDAO = &sessionDAO{}

// CreateSession 创建新会话
func (d *sessionDAO) CreateSession(ctx context.Context, session *model.Session) (string, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return "", fmt.Errorf("get db client failed: %w", err)
	}
	if err := db.Create(session).Error; err != nil {
		return "", fmt.Errorf("create session failed: %w", err)
	}
	return session.SessionID, nil
}

// GetSessionInfo 获取会话信息
func (d *sessionDAO) GetSessionInfo(ctx context.Context, userID, sessionID string) (model.SessionInfo, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return model.SessionInfo{}, fmt.Errorf("get db client failed: %w", err)
	}
	var session model.SessionInfo
	err = db.Model(&model.Session{}).
		Where("user_id = ? AND session_id = ?", userID, sessionID).
		First(&session).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.SessionInfo{}, nil
		}
		return model.SessionInfo{}, fmt.Errorf("get session info failed: %w", err)
	}
	return session, nil
}

// GetSessionsByUserID 获取用户的所有会话
func (d *sessionDAO) GetSessionsByUserID(ctx context.Context, userID string) ([]model.SessionInfo, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}
	var sessions []model.SessionInfo
	if err := db.Model(&model.Session{}).Where("user_id = ?", userID).Order("created_at desc").Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("get sessions by user_id failed: %w", err)
	}
	return sessions, nil
}

// DeleteSession 删除会话
func (d *sessionDAO) DeleteSession(ctx context.Context, userID, sessionID string) error {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return fmt.Errorf("get db client failed: %w", err)
	}
	if err := db.Where("user_id = ? AND session_id = ?", userID, sessionID).Delete(&model.Session{}).Error; err != nil {
		return fmt.Errorf("delete session failed: %w", err)
	}
	return nil
}
