package model

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Admin 管理员模型
type Admin struct {
	AdminID  string `gorm:"uniqueIndex;type:varchar(20);not null" json:"admin_id"`
	Username string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password string `gorm:"type:varchar(255)" json:"-"`
	*gorm.Model
}

// TableName 指定表名
func (Admin) TableName() string {
	return "admins"
}

// EncryptPassword 加密密码
func (a *Admin) EncryptPassword() error {
	if a.Password == "" {
		return nil
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码是否正确
func (a *Admin) CheckPassword(password string) bool {
	if a.Password == "" || password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}
