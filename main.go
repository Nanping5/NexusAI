package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"NexusAi/common/mcpmanager"
	mmysql "NexusAi/common/mysql"
	"NexusAi/common/qdrant"
	"NexusAi/common/rabbitmq"
	myredis "NexusAi/common/redis"
	"NexusAi/config"
	mylogger "NexusAi/pkg/logger"
	"NexusAi/router"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	Setup()
}

func Setup() {
	godotenv.Load()

	if err := config.LoadConfig(); err != nil {
		panic(err)
	}

	if err := mylogger.InitLogger(); err != nil {
		panic(err)
	}

	if err := mmysql.InitMySQL(); err != nil {
		panic(err)
	}

	if err := myredis.InitRedis(); err != nil {
		panic(err)
	}

	// 初始化 Qdrant 客户端
	qdrantCfg := config.GetConfig().QdrantConfig
	if err := qdrant.InitQdrant(qdrantCfg.Host, qdrantCfg.Port, qdrantCfg.APIKey); err != nil {
		mylogger.Logger.Error("Qdrant 初始化失败", zap.Error(err))
	}

	rabbitmq.InitRabbitMQ()

	if err := initMCPConnections(); err != nil {
		mylogger.Logger.Error("MCP 连接初始化失败", zap.Error(err))
	}

	r := router.InitRouter()
	StartServer(r)
}

func StartServer(r *gin.Engine) {
	host := config.GetConfig().MainConfig.Host
	port := config.GetConfig().MainConfig.Port
	addr := fmt.Sprintf("%s:%d", host, port)
	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	// 启动服务器
	go func() {
		mylogger.Logger.Info("NexusAi 启动成功", zap.String("host", host), zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mylogger.Logger.Error("服务器启动失败", zap.Error(err))
			panic(err)
		}
	}()
	Shutdown(srv)
}

// Shutdown 关闭服务器
func Shutdown(srv *http.Server) {

	quit := make(chan os.Signal, 1)
	// 监听 SIGINT和 SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞等待信号
	sig := <-quit
	mylogger.Logger.Info("接收到关闭信号，开始关闭...", zap.String("signal", sig.String()))
	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		mylogger.Logger.Error("服务器关闭错误", zap.Error(err))
	}
	// 关闭 Qdrant 客户端
	if err := qdrant.CloseQdrant(); err != nil {
		mylogger.Logger.Error("Qdrant 关闭错误", zap.Error(err))
	}
	// 关闭 RabbitMQ 连接
	rabbitmq.DestroyRabbitMQ()
	// 关闭 Redis 连接
	if err := myredis.CloseRedis(); err != nil {
		mylogger.Logger.Error("Redis 关闭错误", zap.Error(err))
	}
	// 关闭 MySQL 连接
	if err := mmysql.CloseMySQL(); err != nil {
		mylogger.Logger.Error("MySQL 关闭错误", zap.Error(err))
	}
	// 关闭 MCP 连接
	mcpmanager.GetGlobalMCPManager().DisconnectAll()
	// 关闭日志系统
	if err := mylogger.CloseLogger(); err != nil {
		mylogger.Logger.Error("日志系统关闭错误", zap.Error(err))
	}
	mylogger.Logger.Info("服务器已安全关闭")
}

// initMCPConnections 初始化 MCP 服务连接
func initMCPConnections() error {
	ctx := context.Background()
	mcpManager := mcpmanager.GetGlobalMCPManager()

	// 从环境变量读取 MCP 服务配置
	mcpServerURL := os.Getenv("MCP_SERVER_URL")
	if mcpServerURL == "" {
		mylogger.Logger.Info("未配置 MCP 服务，跳过 MCP 初始化")
		return nil
	}
	// 连接到 MCP 服务器（所有服务共享一个端口）
	err := mcpManager.Connect(ctx, mcpmanager.MCPClientConfig{
		Name:    "nexus-mcp",
		HTTPUrl: mcpServerURL,
	})
	if err != nil {
		mylogger.Logger.Error("连接 MCP 服务失败",
			zap.String("url", mcpServerURL),
			zap.Error(err),
		)
		return err
	}

	mylogger.Logger.Info("MCP 服务连接完成", zap.String("url", mcpServerURL))
	return nil
}
