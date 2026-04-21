package fileController

import (
	ret "NexusAi/common/code"
	response "NexusAi/common/response/common"
	mylogger "NexusAi/pkg/logger"
	"NexusAi/service"

	"github.com/gin-gonic/gin"
)

func UploadRagFile(c *gin.Context) {
	uploadedFile, err := c.FormFile("file")
	if err != nil {
		mylogger.Logger.Error("Failed to get uploaded file: " + err.Error())
		response.Fail(c, ret.CodeInvalidParams)
		return
	}
	userId := c.GetString("user_id")
	if userId == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}

	// 从请求中获取 sessionID
	sessionID := c.PostForm("session_id")
	if sessionID == "" {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	filePath, err := service.FileService.UploadRagFile(c.Request.Context(), userId, sessionID, uploadedFile)
	if err != nil {
		mylogger.Logger.Error("Failed to upload RAG file: " + err.Error())
		response.Fail(c, ret.CodeServerBusy)
		return
	}

	response.Success(c, gin.H{"file_path": filePath})
}
