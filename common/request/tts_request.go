package request

type TTSRequest struct {
	Text string `json:"text" binding:"required"` // 待合成的文本
}
