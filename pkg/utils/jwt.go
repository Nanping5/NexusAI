package utils

import (
	"NexusAi/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims 定义 JWT 载荷结构
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// AdminClaims 管理员 JWT 载荷结构
type AdminClaims struct {
	AdminID  string `json:"admin_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateJwtToken 生成JWT令牌，有效期为24小时
func GenerateJwtToken(userID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    config.GetConfig().JwtConfig.Issuer,
			Subject:   config.GetConfig().JwtConfig.Subject,
			ID:        GenerateUUID(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetConfig().JwtConfig.SecretKey))
}

// ParseJwtToken 解析JWT令牌并返回Claims
func ParseJwtToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JwtConfig.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if token != nil {
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, jwt.ErrSignatureInvalid
}

// GenerateAdminJwtToken 生成管理员 JWT 令牌
func GenerateAdminJwtToken(adminID, username string) (string, error) {
	now := time.Now()
	claims := AdminClaims{
		AdminID:  adminID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    config.GetConfig().JwtConfig.Issuer,
			Subject:   "admin",
			ID:        GenerateUUID(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetConfig().JwtConfig.SecretKey))
}

// ParseAdminJwtToken 解析管理员 JWT 令牌
func ParseAdminJwtToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JwtConfig.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if token != nil {
		if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, jwt.ErrSignatureInvalid
}
