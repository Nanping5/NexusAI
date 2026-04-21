# MCP 服务模块文档

> MCP (Model Context Protocol) 服务提供外部工具集成能力
> 版本: v1.0.0
> 更新日期: 2026-03-27

---

## 概述

MCP 服务是 NexusAi 的外部工具集成层，允许 AI Agent 调用各种外部服务。基于 [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) 实现。

### 架构

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   AI Agent      │────▶│   MCP Manager   │────▶│   MCP Server    │
│  (CloudWeGo)    │     │   (客户端管理)    │     │   (工具服务)     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                        │
                        ┌───────────────────────────────┼───────────────────────────────┐
                        │               │               │               │
                        ▼               ▼               ▼               ▼
                    ┌───────┐       ┌───────┐       ┌───────┐       ┌───────┐
                    │Weather│       │Search │       │Translate│     │  TTS  │
                    └───────┘       └───────┘       └───────┘       └───────┘
```

---

## 可用服务

### 1. Weather 服务

天气查询服务，支持查询指定城市的天气信息。

**工具名称:** `get_weather`

**参数:**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| city | string | 是 | 城市名称（如：北京、上海） |

**返回示例:**
```json
{
    "city": "北京",
    "weather": "晴",
    "temperature": "25°C",
    "humidity": "45%"
}
```

---

### 2. Search 服务

网络搜索服务，支持搜索互联网信息。

**工具名称:** `search`

**参数:**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| query | string | 是 | 搜索关键词 |

**返回示例:**
```json
{
    "results": [
        {
            "title": "搜索结果标题",
            "url": "https://example.com",
            "snippet": "内容摘要..."
        }
    ]
}
```

---

### 3. Translate 服务

文本翻译服务，支持多语言互译。

**工具名称:** `translate`

**参数:**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| text | string | 是 | 待翻译文本 |
| source_lang | string | 否 | 源语言（默认自动检测） |
| target_lang | string | 是 | 目标语言（如：en、zh、ja） |

**返回示例:**
```json
{
    "original": "你好世界",
    "translated": "Hello World",
    "source_lang": "zh",
    "target_lang": "en"
}
```

---

### 4. TTS 服务

语音合成服务，将文本转换为语音。

**工具名称:** `tts`

**参数:**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| text | string | 是 | 待合成文本（最大 5000 字符） |

**返回示例:**
```json
{
    "audio_url": "/static/tts_xxx.wav",
    "format": "wav",
    "duration": 3.5
}
```

---

## 启动方式

### 启动所有服务

```bash
# 默认在 localhost:8081 启动
go run common/mcp/main.go
```

### 启动指定服务

```bash
# 只启动天气和搜索服务
go run common/mcp/main.go --services=weather,search
```

### 指定端口

```bash
go run common/mcp/main.go --address=localhost:9090
```

---

## 目录结构

```
common/mcp/
├── main.go              # MCP 服务入口
├── base/                # 基础框架
│   ├── server.go        # MCP 服务器实现
│   └── service.go       # 服务接口定义
└── services/            # 具体服务实现
    ├── weather/         # 天气服务
    ├── search/          # 搜索服务
    ├── translate/       # 翻译服务
    └── tts/             # TTS 服务
```

---

## 添加新服务

### 1. 创建服务目录

在 `common/mcp/services/` 下创建新目录：

```bash
mkdir common/mcp/services/yourservice
```

### 2. 实现服务接口

创建 `service.go` 文件：

```go
package yourservice

import (
    "NexusAi/common/mcp/base"
)

func GetYourServiceConfig() base.ServiceConfig {
    return base.ServiceConfig{
        Name:        "yourservice",
        Description: "服务描述",
        Tools:       []base.Tool{yourTool},
        Handler:     yourHandler,
    }
}
```

### 3. 注册服务

在 `common/mcp/main.go` 的 `ServiceRegistry` 中注册：

```go
var ServiceRegistry = map[string]func() base.ServiceConfig{
    "weather":   weatherserver.GetWeatherServiceConfig,
    "search":    searchserver.GetSearchServiceConfig,
    "translate": translateserver.GetTranslateServiceConfig,
    "tts":       ttsserver.GetTTSServiceConfig,
    "yourservice": yourservice.GetYourServiceConfig,  // 添加这行
}
```

---

## 配置说明

### 环境变量

| 变量名 | 说明 |
|--------|------|
| `MCP_SERVER_PORT` | MCP 服务端口（默认 8081） |
| `WEATHER_API_KEY` | 天气服务 API Key（如需） |
| `SEARCH_API_KEY` | 搜索服务 API Key（如需） |
| `TRANSLATE_API_KEY` | 翻译服务 API Key（如需） |

### 与主服务集成

在主应用中通过 MCP Manager 连接：

```go
// 创建 MCP 客户端
manager := mcpmanager.NewMCPManager()
manager.Connect("localhost:8081")

// Agent 可通过 manager 调用工具
result, err := manager.CallTool("get_weather", map[string]interface{}{
    "city": "北京",
})
```

---

## 与 AI Agent 集成

MCP 服务与 CloudWeGo Eino Agent 集成，AI 可以自动调用这些工具：

```go
// Agent 配置
agentConfig := &eino.AgentConfig{
    Model: model,
    Tools: []Tool{
        mcpmanager.GetMCPTool("get_weather"),
        mcpmanager.GetMCPTool("search"),
        mcpmanager.GetMCPTool("translate"),
        mcpmanager.GetMCPTool("tts"),
    },
}
```

**示例对话:**

```
用户: 北京今天天气怎么样？
AI: [调用 get_weather 工具]
AI: 北京今天天气晴朗，气温25°C，湿度45%。

用户: 帮我搜索一下 Go 语言的最佳实践
AI: [调用 search 工具]
AI: 我找到了以下 Go 语言最佳实践...
```
