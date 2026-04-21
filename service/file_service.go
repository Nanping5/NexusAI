package service

import (
	"NexusAi/common/rag"
	"NexusAi/config"
	"NexusAi/dao"
	mylogger "NexusAi/pkg/logger"
	"NexusAi/pkg/utils"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type fileService struct {
}

var FileService = &fileService{}

func (s *fileService) UploadRagFile(ctx context.Context, userId, sessionID string, uploadedFile *multipart.FileHeader) (string, error) {

	if err := utils.ValidateFile(uploadedFile); err != nil {
		mylogger.Logger.Error("File validation failed: " + err.Error())
		return "", err
	}

	// 文件存储路径改为 session 级别: uploads/{userId}/{sessionID}/
	sessionDir := filepath.Join("uploads", userId, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		mylogger.Logger.Error("Failed to create session directory: " + err.Error())
		return "", err
	}

	uuid := utils.GenerateUUID()
	ext := filepath.Ext(uploadedFile.Filename)
	newFilename := uuid + ext
	savePath := filepath.Join(sessionDir, newFilename)

	src, err := uploadedFile.Open()
	if err != nil {
		mylogger.Logger.Error("Failed to open uploaded file: " + err.Error())
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		mylogger.Logger.Error("Failed to create destination file: " + err.Error())
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		mylogger.Logger.Error("Failed to save uploaded file: " + err.Error())
		os.Remove(savePath)
		return "", err
	}

	// 文件保存成功后，创建RAG索引
	// Qdrant 使用单集合 + session_id 过滤，所以只需要传入 sessionID
	indexer, err := rag.NewRAGIndexer(sessionID, config.GetConfig().RagConfig.RagEmbeddingModel)
	if err != nil {
		mylogger.Logger.Error("Failed to create RAG indexer: " + err.Error())
		os.Remove(savePath)
		return "", err
	}

	// 索引文件，sessionID 已保存在 indexer 中
	if err := indexer.IndexFile(ctx, savePath); err != nil {
		mylogger.Logger.Error("Failed to index uploaded file: " + err.Error())
		os.Remove(savePath)
		// 删除该 session 的向量点
		rag.DeleteSessionPoints(ctx, sessionID)
		return "", err
	}

	// 删除该 session 下的旧文件（如果有）
	files, err := os.ReadDir(sessionDir)
	if err != nil {
		mylogger.Logger.Error("Failed to read session directory: " + err.Error())
		// 不返回错误，因为新文件已上传成功
	} else {
		for _, file := range files {
			if !file.IsDir() && file.Name() != newFilename {
				// 删除旧文件
				oldFilePath := filepath.Join(sessionDir, file.Name())
				if err := os.Remove(oldFilePath); err != nil {
					mylogger.Logger.Error("Failed to remove old file: " + err.Error())
				}
			}
		}
	}

	return savePath, nil
}

// GetSessionRAGFile 获取指定 session 的 RAG 文件信息
func (s *fileService) GetSessionRAGFile(ctx context.Context, userID, sessionID string) (filePath string, err error) {
	// 验证会话存在且属于该用户
	sessionInfo, err := dao.SessionDAO.GetSessionInfo(ctx, userID, sessionID)
	if err != nil {
		return "", err
	}
	if sessionInfo.SessionID == "" {
		return "", fmt.Errorf("session not found or access denied")
	}

	// 查找 session 目录下的文件
	sessionDir := filepath.Join("uploads", userID, sessionID)
	files, err := os.ReadDir(sessionDir)
	if err != nil {
		return "", fmt.Errorf("failed to read session directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath = filepath.Join(sessionDir, file.Name())
			return filePath, nil
		}
	}

	return "", fmt.Errorf("no RAG file found for session")
}

// DeleteSessionRAGFile 删除指定 session 的 RAG 文件和向量
func (s *fileService) DeleteSessionRAGFile(ctx context.Context, userID, sessionID string) error {
	// 删除 Qdrant 中的向量点
	if err := rag.DeleteSessionPoints(ctx, sessionID); err != nil {
		mylogger.Logger.Error("Failed to delete session points from Qdrant",
			zap.String("sessionID", sessionID),
			zap.Error(err))
		// 继续删除文件，不返回错误
	}

	// 删除文件目录
	sessionDir := filepath.Join("uploads", userID, sessionID)
	if err := os.RemoveAll(sessionDir); err != nil {
		mylogger.Logger.Error("Failed to remove session directory",
			zap.String("sessionDir", sessionDir),
			zap.Error(err))
		return err
	}

	mylogger.Logger.Info("Session RAG file deleted successfully",
		zap.String("sessionID", sessionID))
	return nil
}
