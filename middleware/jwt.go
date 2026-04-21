package middleware

import (
	ret "NexusAi/common/code"
	response "NexusAi/common/response/common"

	"NexusAi/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		// 如果 Header 中没有，尝试从 Query 参数中获取
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			response.Fail(c, ret.CodeNotLogin)
			c.Abort()
			return
		}
		claims, err := utils.ParseJwtToken(token)
		if err != nil {
			response.Fail(c, ret.CodeInvalidToken)
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Next()
	}

}
