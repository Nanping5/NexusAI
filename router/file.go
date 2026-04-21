package router

import (
	fileController "NexusAi/controller/file"

	"github.com/gin-gonic/gin"
)

func RegisterFileRouter(r *gin.RouterGroup) {
	// RAG 文件上传接口
	r.POST("/rag/upload", fileController.UploadRagFile)
}
