package mcpmanager

import (
	"context"
	"fmt"
	"sync"

	mylogger "NexusAi/pkg/logger"

	einomcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

// MCPClientConfig MCP 客户端配置
type MCPClientConfig struct {
	Name    string // MCP 服务名称
	HTTPUrl string // MCP HTTP 服务地址，格式: http://host:port/mcp
}

// MCPClientWrapper MCP 客户端包装器
type MCPClientWrapper struct {
	client    *client.Client
	transport *transport.StreamableHTTP
}

// MCPManager MCP 客户端管理器，管理多个 MCP 服务连接
type MCPManager struct {
	clients map[string]*MCPClientWrapper // map[serviceName]*MCPClientWrapper
	tools   map[string][]tool.BaseTool   // map[serviceName][]tool.BaseTool
	mutex   sync.RWMutex
}

var (
	globalMCPManager *MCPManager
	mcpOnce          sync.Once
)

// GetGlobalMCPManager 获取全局 MCP 管理器实例（单例）
func GetGlobalMCPManager() *MCPManager {
	mcpOnce.Do(func() {
		globalMCPManager = &MCPManager{
			clients: make(map[string]*MCPClientWrapper),
			tools:   make(map[string][]tool.BaseTool),
		}
	})
	return globalMCPManager
}

// Connect 连接到 MCP 服务器
func (m *MCPManager) Connect(ctx context.Context, config MCPClientConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否已连接
	if _, exists := m.clients[config.Name]; exists {
		return fmt.Errorf("MCP client '%s' already connected", config.Name)
	}

	// 创建 HTTP 传输层
	httpTransport, err := transport.NewStreamableHTTP(config.HTTPUrl)
	if err != nil {
		return fmt.Errorf("failed to create HTTP transport: %w", err)
	}

	// 创建 MCP 客户端
	cli := client.NewClient(httpTransport)

	// 初始化连接
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "NexusAi-MCP-Client",
		Version: "1.0.0",
	}

	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		httpTransport.Close()
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// 保存客户端
	m.clients[config.Name] = &MCPClientWrapper{
		client:    cli,
		transport: httpTransport,
	}

	mylogger.Logger.Info("MCP client connected successfully",
		zap.String("name", config.Name),
		zap.String("url", config.HTTPUrl),
	)

	return nil
}

// GetTools 获取指定 MCP 服务的工具列表
func (m *MCPManager) GetTools(ctx context.Context, serviceName string) ([]tool.BaseTool, error) {
	m.mutex.RLock()
	wrapper, exists := m.clients[serviceName]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP client '%s' not found", serviceName)
	}

	// 检查缓存
	m.mutex.RLock()
	cachedTools, cached := m.tools[serviceName]
	m.mutex.RUnlock()
	if cached {
		return cachedTools, nil
	}

	// 从 MCP 服务获取工具
	tools, err := einomcp.GetTools(ctx, &einomcp.Config{Cli: wrapper.client})
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP tools: %w", err)
	}

	// 缓存工具列表
	m.mutex.Lock()
	m.tools[serviceName] = tools
	m.mutex.Unlock()

	mylogger.Logger.Info("MCP tools loaded",
		zap.String("serviceName", serviceName),
		zap.Int("toolCount", len(tools)),
	)

	return tools, nil
}

// GetAllTools 获取所有已连接 MCP 服务的工具
func (m *MCPManager) GetAllTools(ctx context.Context) ([]tool.BaseTool, error) {
	m.mutex.RLock()
	serviceNames := make([]string, 0, len(m.clients))
	for name := range m.clients {
		serviceNames = append(serviceNames, name)
	}
	m.mutex.RUnlock()

	var allTools []tool.BaseTool
	for _, name := range serviceNames {
		tools, err := m.GetTools(ctx, name)
		if err != nil {
			mylogger.Logger.Warn("failed to get tools from MCP service",
				zap.String("serviceName", name),
				zap.Error(err),
			)
			continue
		}
		allTools = append(allTools, tools...)
	}

	return allTools, nil
}

// Disconnect 断开与指定 MCP 服务的连接
func (m *MCPManager) Disconnect(serviceName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	wrapper, exists := m.clients[serviceName]
	if !exists {
		return fmt.Errorf("MCP client '%s' not found", serviceName)
	}

	// 关闭连接
	if err := wrapper.transport.Close(); err != nil {
		mylogger.Logger.Warn("failed to close MCP transport",
			zap.String("serviceName", serviceName),
			zap.Error(err),
		)
	}

	// 删除客户端和缓存的工具
	delete(m.clients, serviceName)
	delete(m.tools, serviceName)

	mylogger.Logger.Info("MCP client disconnected", zap.String("serviceName", serviceName))
	return nil
}

// DisconnectAll 断开所有 MCP 服务连接
func (m *MCPManager) DisconnectAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for name, wrapper := range m.clients {
		if err := wrapper.transport.Close(); err != nil {
			mylogger.Logger.Warn("failed to close MCP transport",
				zap.String("serviceName", name),
				zap.Error(err),
			)
		}
	}

	m.clients = make(map[string]*MCPClientWrapper)
	m.tools = make(map[string][]tool.BaseTool)

	mylogger.Logger.Info("All MCP clients disconnected")
}

// GetConnectedServices 获取已连接的 MCP 服务列表
func (m *MCPManager) GetConnectedServices() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	services := make([]string, 0, len(m.clients))
	for name := range m.clients {
		services = append(services, name)
	}
	return services
}

// Ping 检查 MCP 服务连接状态
func (m *MCPManager) Ping(ctx context.Context, serviceName string) error {
	m.mutex.RLock()
	wrapper, exists := m.clients[serviceName]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("MCP client '%s' not found", serviceName)
	}

	return wrapper.client.Ping(ctx)
}
