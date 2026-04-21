package model

import (
	"gorm.io/gorm"
)

type Message struct {
	SessionID string `gorm:"index;not null;type:varchar(36)" json:"session_id"`
	UserID    string `gorm:"index;type:varchar(20);not null" json:"user_id"`
	Content   string `gorm:"type:text" json:"content"`
	IsUser    bool   `gorm:"not null" json:"is_user"`
	*gorm.Model
}

// History 聊天历史记录
type History struct {
	IsUser  bool   `json:"is_user"`
	Content string `json:"content"`
}
