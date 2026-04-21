package rag

import (
	"NexusAi/common/qdrant"
	"NexusAi/config"
	mylogger "NexusAi/pkg/logger"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/schema"
	qdrantgo "github.com/qdrant/go-client/qdrant"
	"go.uber.org/zap"
)

// RAGIndexer RAG 索引器，负责文档向量化存储
type RAGIndexer struct {
	embedding  MessageEmbedding
	client     *qdrantgo.Client
	collection string
	sessionID  string
	vectorSize uint64
}

// RAGQuery RAG 查询器，负责向量检索
type RAGQuery struct {
	embedding  MessageEmbedding
	client     *qdrantgo.Client
	collection string
	sessionID  string
}

// getCollectionName 获取集合名称
func getCollectionName() string {
	collectionName := config.GetConfig().QdrantConfig.Collection
	if collectionName == "" {
		collectionName = "nexus_rag"
	}
	return collectionName
}

// getEmbeddingAPIKey 获取 embedding 使用的 API Key
// 优先使用 DASHSCOPE_API_KEY，如果没有则尝试其他环境变量
func getEmbeddingAPIKey() string {
	// 优先使用 DashScope 标准环境变量名
	if key := os.Getenv("DASHSCOPE_API_KEY"); key != "" {
		return key
	}
	return ""
}

// NewRAGIndexer 创建一个新的 RAG 索引器实例
func NewRAGIndexer(sessionID, embeddingModel string) (*RAGIndexer, error) {
	ctx := context.Background()

	cfg := config.GetConfig()
	apiKey := getEmbeddingAPIKey()
	vectorSize := uint64(cfg.RagConfig.RagDimension)
	collectionName := getCollectionName()

	if apiKey == "" {
		mylogger.Logger.Error("Embedding API key is empty, please set DASHSCOPE_API_KEY environment variable")
		return nil, fmt.Errorf("embedding API key is empty")
	}

	// 使用 DashScope 向量生成器
	embedder, err := NewDashScopeEmbedder(ctx, cfg.RagConfig.RagBaseURL, apiKey, embeddingModel, cfg.RagConfig.RagDimension)
	if err != nil {
		mylogger.Logger.Error("Failed to create embedding instance", zap.Error(err))
		return nil, err
	}

	// 确保 Qdrant 集合存在
	if err := qdrant.CreateCollectionIfNotExists(ctx, collectionName, vectorSize); err != nil {
		mylogger.Logger.Error("Failed to ensure Qdrant collection exists", zap.Error(err))
		return nil, err
	}

	return &RAGIndexer{
		embedding:  embedder,
		client:     qdrant.Client,
		collection: collectionName,
		sessionID:  sessionID,
		vectorSize: vectorSize,
	}, nil
}

// IndexFile 读取文件内容并创建 RAG 索引
func (r *RAGIndexer) IndexFile(ctx context.Context, filePath string) error {
	// 使用结构体中保存的 sessionID
	sessionID := r.sessionID

	// 索引前先清理该 session 的旧向量，确保单会话单文档
	if err := DeleteSessionPoints(ctx, sessionID); err != nil {
		mylogger.Logger.Warn("Failed to delete old session points before indexing, continuing anyway",
			zap.String("sessionID", sessionID),
			zap.Error(err))
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		mylogger.Logger.Error("Failed to read file for indexing", zap.Error(err))
		return err
	}

	mylogger.Logger.Info("File content read",
		zap.String("sessionID", sessionID),
		zap.String("filePath", filePath),
		zap.Int("contentLength", len(content)))

	// 跳过空文件
	if len(content) == 0 {
		mylogger.Logger.Warn("Skipping empty file",
			zap.String("sessionID", sessionID),
			zap.String("filePath", filePath))
		return nil
	}

	// 从配置获取切分参数
	cfg := config.GetConfig()
	chunkSize := cfg.RagConfig.RagChunkSize
	if chunkSize <= 0 {
		chunkSize = 800
	}
	overlapSize := cfg.RagConfig.RagChunkOverlap
	if overlapSize < 0 {
		overlapSize = 0
	}

	// 使用语义切分 + 滑动窗口
	docs := SplitDocument(string(content), filePath, chunkSize, overlapSize)

	mylogger.Logger.Info("Document split result",
		zap.String("sessionID", sessionID),
		zap.Int("totalChunks", len(docs)))

	// 准备所有点用于批量上传
	var points []*qdrantgo.PointStruct

	// 为每个文档块生成向量
	for i, doc := range docs {
		// 跳过空内容
		doc.Content = strings.TrimSpace(doc.Content)
		if doc.Content == "" {
			mylogger.Logger.Warn("Skipping empty document chunk",
				zap.String("docID", doc.ID),
				zap.Int("chunkIndex", i))
			continue
		}

		// 生成向量
		vectors, err := r.embedding.EmbedStrings(ctx, []string{doc.Content})
		if err != nil {
			mylogger.Logger.Error("Failed to generate embedding",
				zap.String("docID", doc.ID),
				zap.String("content", doc.Content[:min(100, len(doc.Content))]),
				zap.Error(err))
			return err
		}

		// vectors[0] 已经是 float32 类型
		vectorFloat32 := vectors[0]

		// 构建点结构 - 使用数字 ID（基于索引）
		point := &qdrantgo.PointStruct{
			Id:      qdrantgo.NewIDNum(uint64(i + 1)), // 使用数字 ID，从 1 开始
			Vectors: qdrantgo.NewVectors(vectorFloat32...),
			Payload: qdrantgo.NewValueMap(map[string]any{
				"content":    doc.Content,
				"session_id": sessionID,
				"source":     doc.MetaData["source"],
				"chunk":      doc.MetaData["chunk"],
			}),
		}
		points = append(points, point)
	}

	// 批量存储到 Qdrant
	_, err = r.client.Upsert(ctx, &qdrantgo.UpsertPoints{
		CollectionName: r.collection,
		Points:         points,
	})

	if err != nil {
		mylogger.Logger.Error("Failed to upsert points to Qdrant", zap.Error(err))
		return err
	}

	mylogger.Logger.Info("Document indexed successfully",
		zap.String("sessionID", sessionID),
		zap.Int("chunks", len(docs)))

	return nil
}

// DeleteSessionPoints 删除指定 session 的所有向量点
func DeleteSessionPoints(ctx context.Context, sessionID string) error {
	collectionName := getCollectionName()

	// 使用过滤器删除该 session 的所有点
	_, err := qdrant.Client.Delete(ctx, &qdrantgo.DeletePoints{
		CollectionName: collectionName,
		Points: qdrantgo.NewPointsSelectorFilter(&qdrantgo.Filter{
			Must: []*qdrantgo.Condition{
				qdrantgo.NewMatch("session_id", sessionID),
			},
		}),
	})

	if err != nil {
		mylogger.Logger.Error("Failed to delete session points",
			zap.String("sessionID", sessionID),
			zap.Error(err))
		return err
	}

	mylogger.Logger.Info("Session points deleted successfully",
		zap.String("sessionID", sessionID))
	return nil
}

// NewRAGQuery 创建一个新的 RAG 查询实例
func NewRAGQuery(ctx context.Context, sessionID string) (*RAGQuery, error) {
	cfg := config.GetConfig()
	apiKey := os.Getenv("DASHSCOPE_API_KEY")

	// 使用 DashScope 向量生成器
	embedder, err := NewDashScopeEmbedder(ctx, cfg.RagConfig.RagBaseURL, apiKey, cfg.RagConfig.RagEmbeddingModel, cfg.RagConfig.RagDimension)
	if err != nil {
		mylogger.Logger.Error("Failed to create embedding instance for query", zap.Error(err))
		return nil, err
	}

	collectionName := getCollectionName()

	return &RAGQuery{
		embedding:  embedder,
		client:     qdrant.Client,
		collection: collectionName,
		sessionID:  sessionID,
	}, nil
}

// RetrieveDocuments 根据查询语句检索相关文档
func (r *RAGQuery) RetrieveDocuments(ctx context.Context, query string) ([]*schema.Document, error) {
	// 为查询生成向量
	vectors, err := r.embedding.EmbedStrings(ctx, []string{query})
	if err != nil {
		mylogger.Logger.Error("Failed to generate query embedding", zap.Error(err))
		return nil, err
	}

	// vectors[0] 已经是 float32 类型
	vectorFloat32 := vectors[0]

	// 在 Qdrant 中搜索，过滤出当前 session 的文档
	searchResult, err := r.client.Query(ctx, &qdrantgo.QueryPoints{
		CollectionName: r.collection,
		Query:          qdrantgo.NewQuery(vectorFloat32...),
		Filter: &qdrantgo.Filter{
			Must: []*qdrantgo.Condition{
				qdrantgo.NewMatch("session_id", r.sessionID),
			},
		},
		WithPayload: qdrantgo.NewWithPayload(true),
		Limit:       qdrantgo.PtrOf(uint64(5)),
	})

	if err != nil {
		mylogger.Logger.Error("Failed to query Qdrant", zap.Error(err))
		return nil, err
	}

	// 转换搜索结果为 Document
	var docs []*schema.Document
	for _, point := range searchResult {
		doc := &schema.Document{
			ID:       "",
			Content:  "",
			MetaData: map[string]any{},
		}

		// 获取 ID
		if point.Id != nil {
			doc.ID = point.Id.GetUuid()
		}

		// 提取 payload
		if point.Payload != nil {
			if content, ok := point.Payload["content"]; ok {
				if strVal := content.GetStringValue(); strVal != "" {
					doc.Content = strVal
				}
			}
			if source, ok := point.Payload["source"]; ok {
				if strVal := source.GetStringValue(); strVal != "" {
					doc.MetaData["source"] = strVal
				}
			}
			if chunk, ok := point.Payload["chunk"]; ok {
				doc.MetaData["chunk"] = chunk.GetIntegerValue()
			}
		}

		// 添加相似度分数
		doc.MetaData["score"] = point.Score

		docs = append(docs, doc)
	}

	mylogger.Logger.Info("Documents retrieved successfully",
		zap.String("sessionID", r.sessionID),
		zap.Int("count", len(docs)))

	return docs, nil
}

// BuildRagPrompt 构建 RAG 提示语
func BuildRagPrompt(query string, docs []*schema.Document) string {
	if len(docs) == 0 {
		return query
	}

	var contextBuilder strings.Builder
	for i, doc := range docs {
		contextBuilder.WriteString(fmt.Sprintf("【参考文档 %d】\n%s\n\n", i+1, doc.Content))
	}

	prompt := fmt.Sprintf(`以下是用户上传的参考文档，你可以参考这些内容来回答问题：

%s

用户问题：%s

请注意：
1. 如果问题与参考文档相关，请结合文档内容给出详细、专业的回答
2. 如果问题超出参考文档的范围，你可以根据你的知识灵活回答，不必局限于文档
3. 回答要自然流畅，像正常对话一样，不要公式化
4. 如果参考文档有帮助，可以适当提及"根据你上传的文档..."，但不要生硬`, contextBuilder.String(), query)

	return prompt
}
