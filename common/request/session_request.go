package request

type CreateSessionAndSendMessageRequest struct {
	UserQuestion string `json:"user_question" binding:"required"`
	ConfigID     string `json:"config_id"`               // 模型配置 ID
	UseAgent     bool   `json:"use_agent"`               // 是否使用 Agent 模式
}

type ChatSendMessageRequest struct {
	SessionID    string `json:"session_id" binding:"required"`
	UserQuestion string `json:"user_question" binding:"required"`
	ConfigID     string `json:"config_id"`               // 模型配置 ID
	UseAgent     bool   `json:"use_agent"`               // 是否使用 Agent 模式
}
