package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"NexusAi/common/mcp/base"
	searchserver "NexusAi/common/mcp/services/search"
	ttsserver "NexusAi/common/mcp/services/tts"
	translateserver "NexusAi/common/mcp/services/translate"
	weatherserver "NexusAi/common/mcp/services/weather"
	mylogger "NexusAi/pkg/logger"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// ServiceRegistry 服务注册表
// 新增服务时在此注册
var ServiceRegistry = map[string]func() base.ServiceConfig{
	"weather":   weatherserver.GetWeatherServiceConfig,
	"search":    searchserver.GetSearchServiceConfig,
	"translate": translateserver.GetTranslateServiceConfig,
	"tts":       ttsserver.GetTTSServiceConfig,
	// 添加新服务时在这里注册：
}

func main() {
	// 加载 .env 文件（优先加载当前目录，再加载项目根目录）
	godotenv.Load(".env")       // common/mcp/.env
	godotenv.Load("../../.env") // 项目根目录 .env

	// 初始化日志
	if err := mylogger.InitLogger(); err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer mylogger.CloseLogger()

	// 解析命令行参数
	address := flag.String("address", "localhost:8081", "服务器地址，格式为 host:port")
	servicesFlag := flag.String("services", "", "要启动的服务列表，逗号分隔（为空则启动所有服务）")
	flag.Parse()

	// 确定要启动的服务
	var servicesToStart []string
	if *servicesFlag != "" {
		// 用户指定了服务列表
		servicesToStart = strings.Split(*servicesFlag, ",")
		for i, s := range servicesToStart {
			servicesToStart[i] = strings.TrimSpace(s)
		}
	} else {
		// 启动所有已注册的服务
		for name := range ServiceRegistry {
			servicesToStart = append(servicesToStart, name)
		}
	}

	// 验证并收集服务配置
	var serviceConfigs []base.ServiceConfig
	for _, name := range servicesToStart {
		getConfig, exists := ServiceRegistry[name]
		if !exists {
			mylogger.Logger.Error("Unknown service requested",
				zap.String("service", name),
				zap.Strings("available", getAvailableServices()),
			)
			os.Exit(1)
		}
		serviceConfigs = append(serviceConfigs, getConfig())
	}

	if len(serviceConfigs) == 0 {
		fmt.Println("No services to start. Available services:")
		for name := range ServiceRegistry {
			fmt.Printf("  - %s\n", name)
		}
		os.Exit(1)
	}

	// 创建并启动 MCP 服务器（单端口多服务）
	mcpServer := base.NewMcpServer(
		"nexus-mcp-server",
		"1.0.0",
		serviceConfigs,
	)

	mylogger.Logger.Info("Starting MCP server",
		zap.Strings("services", servicesToStart),
		zap.String("addr", *address),
	)

	if err := mcpServer.StartHTTP(*address); err != nil {
		mylogger.Logger.Error("Failed to start MCP server", zap.Error(err))
		panic(err)
	}
}

// getAvailableServices 获取可用服务列表
func getAvailableServices() []string {
	var services []string
	for name := range ServiceRegistry {
		services = append(services, name)
	}
	return services
}
