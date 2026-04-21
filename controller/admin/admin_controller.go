package adminController

import (
	ret "NexusAi/common/code"
	"NexusAi/common/request"
	response "NexusAi/common/response/common"
	"NexusAi/service"

	"github.com/gin-gonic/gin"
)

// Login 管理员登录
func Login(c *gin.Context) {
	var req request.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	admin, err := service.AdminService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Fail(c, ret.CodeLoginError)
		return
	}

	token, err := service.AdminService.GenerateToken(admin)
	if err != nil {
		response.Fail(c, ret.CodeServerBusy)
		return
	}

	response.Success(c, gin.H{
		"token":    token,
		"admin_id": admin.AdminID,
		"username": admin.Username,
	})
}

// GetInfo 获取管理员信息
func GetInfo(c *gin.Context) {
	adminID := c.GetString("admin_id")
	username := c.GetString("username")

	response.Success(c, gin.H{
		"admin_id": adminID,
		"username": username,
	})
}
