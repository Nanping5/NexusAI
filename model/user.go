package model

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	UserID   string `gorm:"uniqueIndex;type:varchar(20);not null" json:"user_id"`
	Nickname string `gorm:"type:varchar(50)" json:"nickname"`
	Email    string `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	Password string `gorm:"type:varchar(255)" json:"-"`
	*gorm.Model
}

// EncryptPassword 加密密码，使用 bcrypt 算法
func (u *User) EncryptPassword() error {
	if u.Password == "" {
		return nil
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码是否正确
func (u *User) CheckPassword(password string) bool {
	if u.Password == "" || password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
