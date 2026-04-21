package base

import (
	"context"
	"fmt"

	mylogger "NexusAi/pkg/logger"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// ToolDefinition 工具定义
type ToolDefinition struct {
	Name        string              // 工具名称
	Description string              // 工具描述
	Parameters  []ToolParameter     // 工具参数
	Handler     ToolHandlerFunc     // 工具处理函数
}

// ToolParameter 工具参数定义
type ToolParameter struct {
	Name        string // 参数名称
	Description string // 参数描述
	Required    bool   // 是否必填
}

// ToolHandlerFunc 工具处理函数类型
type ToolHandlerFunc func(ctx context.Context, args map[string]any) (string, error)

// ServiceConfig 单个服务配置
type ServiceConfig struct {
	Name    string            // 服务名称
	Version string            // 服务版本
	Tools   []ToolDefinition  // 工具列表
}

// McpServerWrapper MCP 服务包装器（支持多服务）
type McpServerWrapper struct {
	server   *server.MCPServer
	services []ServiceConfig
}

// NewMcpServer 创建 MCP 服务实例（支持多服务注册）
func NewMcpServer(serverName, serverVersion string, services []ServiceConfig) *McpServerWrapper {
	mcpServer := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	wrapper := &McpServerWrapper{
		server:   mcpServer,
		services: services,
	}

	// 注册所有服务的所有工具
	totalTools := 0
	for _, svc := range services {
		for _, tool := range svc.Tools {
			wrapper.registerTool(svc.Name, tool)
			totalTools++
		}
	}

	mylogger.Logger.Info("MCP server created",
		zap.String("name", serverName),
		zap.String("version", serverVersion),
		zap.Int("serviceCount", len(services)),
		zap.Int("toolCount", totalTools),
	)

	return wrapper
}

// registerTool 注册单个工具
func (w *McpServerWrapper) registerTool(serviceName string, def ToolDefinition) {
	// 构建工具选项
	opts := []mcp.ToolOption{
		mcp.WithDescription(def.Description),
	}

	// 添加参数
	for _, param := range def.Parameters {
		opts = append(opts, mcp.WithString(
			param.Name,
			mcp.Description(param.Description),
			mcp.Required(),
		))
	}

	// 创建工具
	tool := mcp.NewTool(def.Name, opts...)

	// 注册工具处理函数
	w.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		// 调用业务处理函数
		result, err := def.Handler(ctx, args)
		if err != nil {
			mylogger.Logger.Error("tool execution failed",
				zap.String("service", serviceName),
				zap.String("tool", def.Name),
				zap.Error(err),
			)
			return nil, fmt.Errorf("tool '%s' execution failed: %w", def.Name, err)
		}

		mylogger.Logger.Info("tool executed successfully",
			zap.String("service", serviceName),
			zap.String("tool", def.Name),
		)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: result,
				},
			},
		}, nil
	})

	mylogger.Logger.Debug("tool registered",
		zap.String("service", serviceName),
		zap.String("tool", def.Name),
	)
}

// GetServer 获取底层 MCP Server 实例
func (w *McpServerWrapper) GetServer() *server.MCPServer {
	return w.server
}

// StartHTTP 启动 HTTP 服务
func (w *McpServerWrapper) StartHTTP(httpAddr string) error {
	httpServer := server.NewStreamableHTTPServer(w.server)
	mylogger.Logger.Info("Starting MCP HTTP server",
		zap.Int("services", len(w.services)),
		zap.String("addr", httpAddr),
	)
	return httpServer.Start(httpAddr)
}
