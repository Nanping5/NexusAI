package ttsController

import (
	ret "NexusAi/common/code"
	"NexusAi/common/request"
	response "NexusAi/common/response/common"
	mylogger "NexusAi/pkg/logger"
	"NexusAi/service"

	"github.com/gin-gonic/gin"
)

// SynthesizeSpeech 语音合成接口
func SynthesizeSpeech(c *gin.Context) {
	userId := c.GetString("user_id")
	if userId == "" {
		response.Fail(c, ret.CodeNotLogin)
		return
	}

	req := request.TTSRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	// 文本长度检查
	if len(req.Text) > 5000 {
		response.Fail(c, ret.CodeInvalidParams)
		return
	}

	audioData, err := service.TTSService.SynthesizeSpeechData(c.Request.Context(), req.Text)
	if err != nil {
		mylogger.Logger.Error("TTS synthesis failed: " + err.Error())
		response.Fail(c, ret.TTSFail)
		return
	}

	// 直接返回音频文件流
	c.Data(200, "audio/wav", audioData)
}

// TTSServiceStatus TTS 服务状态
func TTSServiceStatus(c *gin.Context) {
	response.Success(c, gin.H{
		"format":      "wav",
		"sample_rate": 16000,
		"voice":       "zhixiaoxia",
		"status":      "ok",
	})
}
