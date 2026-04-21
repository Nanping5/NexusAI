package router

import (
	ttsController "NexusAi/controller/tts"

	"github.com/gin-gonic/gin"
)

// RegisterTTSRouter 注册 TTS 语音合成路由
func RegisterTTSRouter(r *gin.RouterGroup) {
	r.POST("/tts", ttsController.SynthesizeSpeech)
	r.GET("/tts/status", ttsController.TTSServiceStatus)
}
