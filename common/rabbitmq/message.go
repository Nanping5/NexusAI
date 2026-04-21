package rabbitmq

import (
	"NexusAi/dao"
	"NexusAi/model"
	mylogger "NexusAi/pkg/logger"
	"context"
	"encoding/json"

	"github.com/streadway/amqp"
)

// MessageMQParam 定义发送到 RabbitMQ 的消息参数结构体
type MessageMQParam struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
	UserID    string `json:"user_id"`
	IsUser    bool   `json:"is_user"`
}

// GenerateMessageMQParam 生成发送到 RabbitMQ 的消息参数
func GenerateMessageMQParam(sessionID, content, userID string, isUser bool) ([]byte, error) {
	param := MessageMQParam{
		SessionID: sessionID,
		Content:   content,
		UserID:    userID,
		IsUser:    isUser,
	}
	data, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MQMessage 处理从 RabbitMQ 接收到的消息
func MQMessage(msg *amqp.Delivery) error {

	var param MessageMQParam
	err := json.Unmarshal(msg.Body, &param)
	if err != nil {
		mylogger.Logger.Error("Failed to unmarshal message: " + err.Error())
		return err
	}
	newMsg := &model.Message{
		SessionID: param.SessionID,
		Content:   param.Content,
		UserID:    param.UserID,
		IsUser:    param.IsUser,
	}
	// 将消息保存到数据库
	_, err = dao.MessageDAO.CreateMessage(context.Background(), newMsg)
	if err != nil {
		mylogger.Logger.Error("Failed to save message to database: " + err.Error())
		return err
	}
	return nil
}
