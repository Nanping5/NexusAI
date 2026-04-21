package dao

import (
	mmysql "NexusAi/common/mysql"
	"NexusAi/model"
	"context"
	"fmt"
)

type messageDAO struct {
}

var MessageDAO = &messageDAO{}

func (dao *messageDAO) CreateMessage(ctx context.Context, msg *model.Message) (*model.Message, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}
	if err := db.Create(msg).Error; err != nil {
		return nil, fmt.Errorf("create message failed: %w", err)
	}
	return msg, nil
}

func (dao *messageDAO) GetMessagesBySessionID(ctx context.Context, sessionID string) ([]model.Message, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}
	var messages []model.Message
	if err := db.Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("get messages by session_id failed: %w", err)
	}
	return messages, nil
}

func (dao *messageDAO) GetMessagesBySessionIDs(ctx context.Context, sessionIDs []string) ([]model.Message, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db client failed: %w", err)
	}
	var messages []model.Message
	if err := db.Where("session_id IN ?", sessionIDs).Order("created_at asc").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("get messages by session_ids failed: %w", err)
	}
	return messages, nil
}

// GetAllMessages 获取所有消息（带分页）
func (dao *messageDAO) GetAllMessages(ctx context.Context, page, pageSize int) ([]model.Message, int64, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("get db client failed: %w", err)
	}
	var messages []model.Message
	var total int64

	// 统计总数
	if err := db.Model(&model.Message{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count messages failed: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Order("created_at asc").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("get all messages failed: %w", err)
	}
	return messages, total, nil
}

// GetMessagesBySessionIDWithPage 根据会话ID分页获取消息
func (dao *messageDAO) GetMessagesBySessionIDWithPage(ctx context.Context, sessionID string, page, pageSize int) ([]model.Message, int64, error) {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("get db client failed: %w", err)
	}
	var messages []model.Message
	var total int64

	query := db.Model(&model.Message{}).Where("session_id = ?", sessionID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count messages by session_id failed: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Where("session_id = ?", sessionID).Order("created_at asc").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("get messages by session_id with page failed: %w", err)
	}
	return messages, total, nil
}

// DeleteMessagesBySessionID 删除会话的所有消息
func (dao *messageDAO) DeleteMessagesBySessionID(ctx context.Context, sessionID string) error {
	db, err := mmysql.NewDbClient(ctx)
	if err != nil {
		return fmt.Errorf("get db client failed: %w", err)
	}
	if err := db.Where("session_id = ?", sessionID).Delete(&model.Message{}).Error; err != nil {
		return fmt.Errorf("delete messages by session_id failed: %w", err)
	}
	return nil
}
