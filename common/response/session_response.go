package SessionResponse

import "NexusAi/model"

type CreateSessionAndSendMessageResponse struct {
	SessionID string `json:"session_id"`
	AIMessage string `json:"ai_message"`
}

type ChatHistoryResponse struct {
	SessionID string          `json:"session_id"`
	History   []model.History `json:"history"`
}
