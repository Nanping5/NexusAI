package middleware

import (
	ret "NexusAi/common/code"
	response "NexusAi/common/response/common"
	"NexusAi/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminJWT 管理员 JWT 认证中间件
func AdminJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 尝试从 Query 参数获取
			authHeader = c.Query("token")
		}

		var tokenString string
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			tokenString = authHeader
		}

		if tokenString == "" {
			response.Fail(c, ret.CodeNotLogin)
			c.Abort()
			return
		}

		claims, err := utils.ParseAdminJwtToken(tokenString)
		if err != nil {
			response.Fail(c, ret.CodeInvalidToken)
			c.Abort()
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
