package online

import (
	"NexusAi/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetOnlineCount(c *gin.Context) {
	count := service.OnlineService.GetOnlineCount()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"count": count,
		},
	})
}

func Heartbeat(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "unauthorized",
		})
		return
	}
	service.OnlineService.UserOnline(userID.(string))
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
	})
}
