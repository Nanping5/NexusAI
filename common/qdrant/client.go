package qdrant

import (
	"context"
	"fmt"

	mylogger "NexusAi/pkg/logger"

	"github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"
)

var Client *qdrant.Client

// InitQdrant 初始化 Qdrant 客户端
func InitQdrant(host string, port int, apiKey string) error {
	var err error

	config := &qdrant.Config{
		Host: host,
		Port: port,
	}

	// 如果有 API Key（云服务），添加认证
	if apiKey != "" {
		config.APIKey = apiKey
	}

	Client, err = qdrant.NewClient(config)
	if err != nil {
		mylogger.Logger.Error("Failed to create Qdrant client", zap.Error(err))
		return err
	}

	// 测试连接
	ctx := context.Background()
	_, err = Client.ListCollections(ctx)
	if err != nil {
		mylogger.Logger.Error("Failed to connect to Qdrant", zap.Error(err))
		return err
	}

	mylogger.Logger.Info("Qdrant client initialized successfully",
		zap.String("host", host),
		zap.Int("port", port))

	return nil
}

// CloseQdrant 关闭 Qdrant 客户端
func CloseQdrant() error {
	if Client == nil {
		return nil
	}
	return Client.Close()
}

// CreateCollectionIfNotExists 创建集合（如果不存在）
func CreateCollectionIfNotExists(ctx context.Context, collectionName string, vectorSize uint64) error {
	// 检查集合是否存在
	exists, err := Client.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection exists: %w", err)
	}

	if exists {
		mylogger.Logger.Info("Collection already exists, skipping creation",
			zap.String("collection", collectionName))
		return nil
	}

	// 创建新集合
	err = Client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})

	if err != nil {
		mylogger.Logger.Error("Failed to create collection",
			zap.String("collection", collectionName),
			zap.Error(err))
		return fmt.Errorf("failed to create collection: %w", err)
	}

	mylogger.Logger.Info("Collection created successfully",
		zap.String("collection", collectionName),
		zap.Uint64("vectorSize", vectorSize))

	return nil
}

// DeleteCollection 删除集合
func DeleteCollection(ctx context.Context, collectionName string) error {
	err := Client.DeleteCollection(ctx, collectionName)
	if err != nil {
		mylogger.Logger.Error("Failed to delete collection",
			zap.String("collection", collectionName),
			zap.Error(err))
		return err
	}

	mylogger.Logger.Info("Collection deleted successfully",
		zap.String("collection", collectionName))
	return nil
}

// CollectionExists 检查集合是否存在
func CollectionExists(ctx context.Context, collectionName string) (bool, error) {
	return Client.CollectionExists(ctx, collectionName)
}
