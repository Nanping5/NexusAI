package userController

import (
	ret "NexusAi/common/code"
	myemail "NexusAi/common/email"
	"NexusAi/common/request"
	response "NexusAi/common/response/common"
	"NexusAi/service"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var req request.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	token, errCode := service.UserService.Register(c.Request.Context(), req.Email, req.Password, req.Captcha)
	if errCode != ret.CodeSuccess {
		response.Fail(c, errCode)
		return
	}

	response.Success(c, gin.H{"token": token})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req request.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	token, errCode := service.UserService.Login(c.Request.Context(), req.Email, req.Password)
	if errCode != ret.CodeSuccess {
		response.Fail(c, errCode)
		return
	}

	response.Success(c, gin.H{"token": token})
}

// GetUserInfo 获取用户信息
func GetUserInfo(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}
	userInfo, errCode := service.UserService.GetUserInfo(c.Request.Context(), userID)
	if errCode != ret.CodeSuccess {
		response.Fail(c, errCode)
		return
	}

	response.Success(c, userInfo)
}

// UpdateNickname 更新用户昵称
func UpdateNickname(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}

	var req struct {
		Nickname string `json:"nickname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	errCode := service.UserService.UpdateNickname(c.Request.Context(), userID, req.Nickname)
	if errCode != ret.CodeSuccess {
		response.Fail(c, errCode)
		return
	}

	response.Success(c, nil)
}

// SendCaptcha 发送验证码
func SendCaptcha(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	msg, err := myemail.SendCaptcha(email)
	if err != nil {
		response.FailWithMsg(c, ret.CodeServerBusy, msg)
		return
	}

	response.SuccessWithMsg(c, msg, nil)
}
