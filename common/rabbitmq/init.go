package rabbitmq

import (
	"context"
	mylogger "NexusAi/pkg/logger"
)

var (
	RMQMessage *RabbitMQ
)

// InitRabbitMQ 初始化 RabbitMQ 连接和消费者
func InitRabbitMQ() {
	InitConn()
	var err error
	RMQMessage, err = NewWorkerRabbitMQ("Message")
	if err != nil {
		mylogger.Logger.Error("Failed to create RabbitMQ worker: " + err.Error())
		panic(err)
	}

	// 启动带 context 控制的消费者（注意启动顺序：先 Add，再赋值 cancel，最后启动 goroutine）
	consumerWg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	consumerCancel = cancel
	go func() {
		defer func() {
			if r := recover(); r != nil {
				mylogger.Logger.Error("RabbitMQ consumer panic recovered: " + toString(r))
			}
		}()
		defer consumerWg.Done()
		RMQMessage.ConsumeContext(ctx, MQMessage)
	}()

	// 启动连接监控
	StartConnectionMonitor()
}

// DestroyRabbitMQ 关闭 RabbitMQ 连接和通道
func DestroyRabbitMQ() {
	// 先停止消费者
	if consumerCancel != nil {
		consumerCancel()
		consumerWg.Wait()
	}

	if RMQMessage != nil {
		RMQMessage.Destroy()
	}
	CloseGlobalConn()
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return "unknown error"
}
