package middleware

import (
	myredis "NexusAi/common/redis"
	ret "NexusAi/common/code"
	"NexusAi/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Limit   int64  // 时间窗口内最大请求数
	APIName string // API 标识
}

// RateLimit 接口限流中间件
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 JWT 获取用户 ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": ret.CodeUnauthorized,
				"msg":  "未授权访问",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": ret.CodeUnauthorized,
				"msg":  "用户信息无效",
			})
			c.Abort()
			return
		}

		// 检查是否超过限流
		exceeded, err := myredis.CheckLimit(userIDStr, config.APIName, config.Limit)
		if err != nil {
			// Redis 出错时不阻止请求，仅记录日志
			// 继续处理请求
		} else if exceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		// 增加请求计数
		_, _ = myredis.IncrLimit(userIDStr, config.APIName)

		c.Next()
	}
}

// OptionalRateLimit 可选限流中间件（无 JWT 时跳过）
func OptionalRateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从 JWT 获取用户 ID
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// 提取 token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}
		tokenString := parts[1]

		claims, err := utils.ParseJwtToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// 检查是否超过限流
		exceeded, err := myredis.CheckLimit(claims.UserID, config.APIName, config.Limit)
		if err != nil {
			// Redis 出错时不阻止请求
			c.Next()
			return
		}

		if exceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		// 增加请求计数
		_, _ = myredis.IncrLimit(claims.UserID, config.APIName)

		c.Next()
	}
}
