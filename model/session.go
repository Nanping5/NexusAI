package model

import (
	"gorm.io/gorm"
)

type Session struct {
	SessionID   string `gorm:"uniqueIndex;type:varchar(36);not null" json:"session_id"`
	UserID      string `gorm:"index;type:varchar(20);not null" json:"user_id"`
	Title       string `gorm:"type:varchar(100)" json:"title"`
	RagFileID   string `gorm:"type:varchar(36)" json:"rag_file_id"`     // RAG 文件 UUID
	RagFileName string `gorm:"type:varchar(255)" json:"rag_file_name"` // RAG 文件原始文件名
	*gorm.Model
}

func (Session) TableName() string {
	return "sessions"
}

type SessionInfo struct {
	SessionID string `json:"session_id"`
	Title     string `json:"title"`
	*gorm.Model
}
