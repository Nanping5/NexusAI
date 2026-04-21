package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// ErrRandomIntFailed 随机数生成失败错误
var ErrRandomIntFailed = fmt.Errorf("failed to generate secure random int")

// GenerateUUID 生成标准的 UUID (36位)
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateShortID 生成随机数字 ID
func GenerateShortID(length int) (string, error) {
	format := fmt.Sprintf("%%0%dd", length)
	max := int64(1)
	for i := 0; i < length; i++ {
		max *= 10
	}
	n, err := generateSecureRandomInt(max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(format, n), nil
}

// GenerateUserID 生成 U+6位数字的用户 ID (如 U123456)
func GenerateUserID() string {
	id, err := GenerateShortID(6)
	if err != nil {
		// 如果加密随机数生成失败，使用时间戳作为后备方案
		return fmt.Sprintf("U%d", time.Now().UnixNano()%1000000)
	}
	return "U" + id
}

// GenerateGroupID 生成 G+6位数字的群组 ID (如 G123456)
func GenerateGroupID() string {
	id, err := GenerateShortID(6)
	if err != nil {
		return fmt.Sprintf("G%d", time.Now().UnixNano()%1000000)
	}
	return "G" + id
}

// GenerateSessionID 生成 S+8位数字的会话 ID
func GenerateSessionID() string {
	id, err := GenerateShortID(8)
	if err != nil {
		return fmt.Sprintf("S%d", time.Now().UnixNano()%100000000)
	}
	return "S" + id
}

// GenerateMessageID 生成 M+10位数字的消息 ID
func GenerateMessageID() string {
	id, err := GenerateShortID(10)
	if err != nil {
		return fmt.Sprintf("M%d", time.Now().UnixNano()%10000000000)
	}
	return "M" + id
}

// GenerateApplyID 生成 A+8位数字的申请 ID
func GenerateApplyID() string {
	id, err := GenerateShortID(8)
	if err != nil {
		return fmt.Sprintf("A%d", time.Now().UnixNano()%100000000)
	}
	return "A" + id
}

func generateSecureRandomInt(max int64) (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, ErrRandomIntFailed
	}
	return n.Int64(), nil
}
