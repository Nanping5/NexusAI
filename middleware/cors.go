package middleware

import (
	"NexusAi/config"

	"github.com/gin-gonic/gin"
)

func CORS() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// 从配置获取允许的域名，支持多域名配置
		allowedOrigins := config.GetConfig().CORSConfig.AllowedOrigins
		origin := ctx.GetHeader("Origin")

		// 检查 origin 是否在允许列表中
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			if origin != "" {
				ctx.Header("Access-Control-Allow-Origin", origin)
			} else if len(allowedOrigins) > 0 {
				ctx.Header("Access-Control-Allow-Origin", allowedOrigins[0])
			}
		}

		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Max-Age", "86400")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next()
	}
}
