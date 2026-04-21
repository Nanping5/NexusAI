# NexusAi Backend

一个功能完整的 AI 对话平台后端，集成用户认证、AI 对话、RAG 检索增强、图像识别、语音合成等多种功能。

## 技术栈

| 组件 | 技术选型 | 版本 |
|------|----------|------|
| 编程语言 | Go | 1.25.4 |
| Web 框架 | Gin | v1.12.0 |
| ORM | GORM | v1.31.1 |
| 数据库 | MySQL | 8.0+ |
| 缓存 | Redis | 6.0+ |
| 向量数据库 | Qdrant | - |
| 消息队列 | RabbitMQ | 3.x |
| AI 框架 | CloudWeGo Eino | v0.8.5 |
| LLM 支持 | OpenAI 兼容 API | - |
| AI 推理 | - | - |
| MCP 协议 | mark3labs/mcp-go | v0.45.0 |
| 认证 | JWT | golang-jwt/v4 |
| 配置管理 | Viper | TOML 格式 |
| 日志 | Zap | v1.27.1 |
| 语音合成 | 阿里云 NLS | - |

## 核心亮点 ✨

### 1. 多级 Redis 缓存架构
- **模型配置缓存**：纯 Redis 缓存，5 分钟 TTL
- **会话列表缓存**：用户会话列表 Redis 存储
- **消息历史缓存**：会话历史 Redis 缓存，支持 Redis → 数据库回源
- **用户信息缓存**：用户信息 30 分钟缓存
- **在线用户追踪**：Redis Set 集合管理在线用户
- **接口限流**：基于 Redis 的滑动窗口限流
- **缓存失效机制**：写操作时自动失效，保证一致性

### 2. 上下文长度管理
- **滑动窗口策略**：自动裁剪超出限制的历史消息
- **Token 统计**：输入/输出 Token 实时统计
- **可配置策略**：支持 `sliding_window` / `summary` 策略
- **智能裁剪**：基于消息轮次和 Token 数双重限制

### 3. 消息队列异步处理
- **消息持久化**：RabbitMQ 异步保存聊天消息
- **自动重连**：连接监控和自动恢复机制（指数退避）
- **手动确认**：消息处理成功才 Ack，保证可靠性
- **失败重试**：处理失败消息重新入队重试
- **并发安全**：互斥锁保护消费者状态，防止竞态和死锁

### 4. 完整的 AI 对话能力
- 支持创建会话和多轮对话
- 流式响应（SSE）
- 对话历史持久化
- 动态模型切换（通过管理后台配置）
- Agent 模式支持（ReAct）
- MCP 工具调用
- RAG 检索增强生成

## 项目结构

```
NexusAi/
├── main.go                    # 应用入口
├── config/                    # 配置管理
│   ├── config.go              # 配置结构定义和加载
│   └── config.toml            # 配置文件
├── common/                    # 公共组件
│   ├── ai_helper/             # AI 模型封装（Eino 框架）
│   │   ├── ai_helper.go       # AIHelper 核心类（消息队列、上下文管理）
│   │   ├── agent.go           # ReAct Agent 实现
│   │   ├── manager.go         # AIHelper 生命周期管理（自动清理）
│   │   ├── factory.go         # 模型工厂
│   │   ├── model.go           # AI 模型定义
│   │   └── model_config_cache.go  # 模型配置缓存
│   ├── code/                  # 错误码定义
│   ├── email/                 # 邮件服务
│   ├── image/                 # 图像识别
│   ├── mcp/                   # MCP 服务实现
│   │   ├── main.go            # MCP 服务入口
│   │   ├── base/server.go     # MCP 基础服务
│   │   └── services/          # 各类 MCP 服务
│   │       ├── weather/       # 天气查询
│   │       ├── search/        # 网络搜索
│   │       ├── translate/     # 文本翻译
│   │       └── tts/           # 语音合成
│   ├── mcpmanager/            # MCP 客户端管理器
│   ├── mysql/                 # MySQL 连接
│   ├── qdrant/                # Qdrant 向量数据库客户端
│   ├── rabbitmq/              # RabbitMQ 消息队列
│   │   ├── init.go           # 初始化和消费者管理
│   │   ├── rabbitmq.go       # RabbitMQ 连接和操作封装
│   │   └── message.go        # 消息处理
│   ├── rag/                   # RAG 检索增强生成
│   │   ├── rag.go           # RAG 核心实现
│   │   ├── splitter.go       # 文档分块
│   │   └── embedder.go       # 向量化嵌入
│   ├── redis/                 # Redis 连接和缓存操作
│   │   ├── redis.go         # Redis 客户端封装
│   │   └── key.go           # Key 命名规范
│   ├── request/               # 请求结构体定义
│   ├── response/              # 响应结构体定义
│   └── tts/                   # 阿里云语音合成
├── controller/                # 控制器层
│   ├── user/                  # 用户控制器
│   ├── session/               # 会话控制器
│   ├── image/                 # 图像控制器
│   ├── file/                  # 文件控制器
│   ├── tts/                   # TTS 控制器
│   ├── online/                # 在线用户控制器
│   └── admin/                 # 管理员控制器
├── dao/                       # 数据访问层
├── middleware/                # 中间件
│   ├── jwt.go                 # JWT 认证中间件
│   ├── admin_jwt.go           # 管理员 JWT 中间件
│   ├── cors.go                # CORS 跨域中间件
│   └── rate_limit.go          # 接口限流中间件
├── model/                     # 数据模型定义
├── pkg/                       # 工具包
│   ├── logger/                # Zap 日志封装
│   └── utils/                 # 工具函数
├── router/                    # 路由定义
│   ├── router.go            # 路由注册入口
│   ├── user.go              # 用户路由
│   ├── ai.go                # AI 对话路由
│   ├── file.go              # 文件路由
│   ├── tts.go               # TTS 路由
│   └── admin.go             # 管理员路由
├── service/                   # 服务层
│   ├── session_service.go    # 会话服务（对话、消息历史、流式响应）
│   ├── user_service.go       # 用户服务
│   ├── file_service.go       # 文件服务（RAG 上传）
│   ├── admin_service.go      # 管理员服务
│   ├── model_config_service.go # 模型配置服务
│   ├── tts_service.go        # 语音合成服务
│   └── online_service.go     # 在线用户服务
├── logs/                      # 日志目录
├── uploads/                   # 上传文件存储
└── docs/                      # API 文档
```

## 核心功能

### 1. 用户认证
- 邮箱注册/登录
- JWT Token 认证
- 验证码邮件发送（Redis 缓存，3分钟有效期）
- 用户信息管理（Redis 缓存）
- 密码 bcrypt 加密

### 2. AI 对话
- 支持创建会话和多轮对话
- 流式响应（SSE）+ 非流式响应
- 对话历史持久化（RabbitMQ 异步写入 + 数据库回源）
- 动态模型切换（通过管理后台配置）
- Agent 模式支持（ReAct）可配置最大迭代次数
- MCP 工具调用
- **上下文长度管理**：滑动窗口自动裁剪 + Token 双重限制
- **Token 统计**：实时统计输入/输出 Token
- **请求队列**：防止并发消息错乱的流控机制
- **智能清理**：AIHelper 自动过期回收（可配置间隔和时长）

### 3. RAG 检索增强生成
- 知识库文件上传（PDF/TXT/DOCX）
- 文档向量化存储（Qdrant）
- 会话级别知识隔离
- 自动检索相关文档增强上下文
- 文档分块处理（可配置 chunk size 和 overlap）
- 向量维度可配置（默认 1024）

### 4. 图像识别
- 基于 MobileNetV2 的图像分类
- 本地模型推理
- 支持 JPEG/PNG/GIF/BMP/TIFF
- 1000 种 ImageNet 分类

### 5. 语音合成 (TTS)
- 基于阿里云智能语音服务
- 文本转语音（WAV 格式，16000Hz）
- 支持最大 5000 字符

### 6. 管理员功能
- 管理员登录认证
- AI 模型配置 CRUD
- 模型启用/禁用
- 设置默认模型

### 7. MCP 服务集成
- 天气查询服务
- 网络搜索服务
- 文本翻译服务
- TTS 服务

### 8. 在线用户管理
- Redis Set 实时追踪在线用户
- 用户状态上报（心跳机制）
- 分布式在线状态同步

### 9. 文件服务
- 用户上传文件管理
- RAG 文档上传和存储
- 会话级别文件隔离
- 目录遍历安全防护

## 快速开始

### 环境要求
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+
- Qdrant (可选，用于 RAG)
- RabbitMQ 3.x（可选）

### 配置

1. 复制配置文件模板
```bash
cp config/config.toml.example config/config.toml
```

2. 修改配置文件 `config/config.toml`
```toml
[main_config]
app_name = "NexusAi"
host = "0.0.0.0"
port = 8080

[mysql_config]
host = "localhost"
port = 3306
db_name = "nexus_ai"
user = "root"
password = "your_password"

[redis_config]
host = "localhost"
port = 6379
password = ""
db = 0

[jwt_config]
secret_key = "your-jwt-secret"
issuer = "nexus-ai"

[smtp]
email_addr = "your-email@example.com"
smtp_key = "your-smtp-key"
smtp_server = "smtp.example.com"

[image_recognition_config]
model_path = "./common/image/models/mobilenetv2.onnx"
label_path = "./common/image/labels/imagenet_labels.txt"

# 注意: 图像识别模型请参考 common/image/ 目录下的说明

[rag_config]
rag_dimension = 1024
rag_chat_model_name = "deepseek-chat"
rag_doc_dir = "./Rag_docs"
rag_base_url = "https://api.openai.com/v1"
rag_embedding_model = "text-embedding-v3"

[voice_service_config]
access_key_id = "your-access-key-id"
access_key_secret = "your-access-key-secret"
app_key = "your-app-key"

# AI 上下文配置
[ai_config]
max_context_messages = 20     # 最大保留消息轮次
max_context_tokens = 16000    # 最大上下文 Token 数
context_strategy = "sliding_window"  # 上下文策略
request_queue_size = 100     # 请求队列大小
default_max_iterations = 10   # Agent 默认最大迭代次数
helper_cleanup_interval = 5   # Helper 清理间隔（分钟）
helper_expire_duration = 30   # Helper 过期时间（分钟）

[rabbitmq_config]
host = "localhost"
port = 5672
username = "guest"
password = "guest"
vhost = "/"
```

3. 设置环境变量（可选）
```bash
# Agent 模式（可选）
export AGENT_MODE=true
export AGENT_MAX_ITERATIONS=10

# MCP 服务（可选）
export MCP_SERVER_URL="localhost:8081"

# 默认系统提示词（可选）
export DEFAULT_SYSTEM_PROMPT="You are a helpful assistant."

# 图像识别模型路径（可选）
# export IMAGE_MODEL_PATH="./common/image/models/mobilenetv2.onnx"
```

> **注意**: AI 模型配置现在通过管理后台动态管理，不再需要在 .env 中配置。请登录管理后台添加模型配置。

### 运行

```bash
# 安装依赖
go mod download

# 运行主服务
go run main.go

# 运行 MCP 服务（可选）
go run common/mcp/main.go --address=localhost:8081
go run common/mcp/main.go --services=weather,search
```

服务默认运行在 `http://localhost:8080`

## API 文档

详细 API 文档请参阅 [docs/api/](docs/api/) 目录：

| 模块 | 文档 | 说明 |
|------|------|------|
| 用户模块 | [user_api.md](docs/api/user_api.md) | 注册、登录、用户信息 |
| AI 会话模块 | [ai_api.md](docs/api/ai_api.md) | 对话、流式响应、历史记录、模型列表 |
| 图像识别模块 | [image_api.md](docs/api/image_api.md) | 图像分类识别 |
| RAG 文件模块 | [rag_api.md](docs/api/rag_api.md) | 知识库文件上传 |
| 语音合成模块 | [tts_api.md](docs/api/tts_api.md) | 文本转语音 |
| 管理员模块 | [admin_api.md](docs/api/admin_api.md) | 管理员登录、模型配置管理 |

### API 快速示例

```bash
# 1. 发送验证码
curl "http://localhost:8080/api/v1/user/send_captcha?email=user@example.com"

# 2. 注册
curl -X POST "http://localhost:8080/api/v1/user/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"123456","captcha":"123456"}'

# 3. 登录获取 Token
curl -X POST "http://localhost:8080/api/v1/user/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"123456"}'

# 4. 获取可用模型列表
curl "http://localhost:8080/api/v1/ai/models" \
  -H "Authorization: Bearer <token>"

# 5. AI 对话（流式）
curl -X POST "http://localhost:8080/api/v1/ai/session/create/stream" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_question":"你好","model_config_id":"<config_id>"}'

# 6. 图像识别
curl -X POST "http://localhost:8080/api/v1/image/recognize" \
  -H "Authorization: Bearer <token>" \
  -F "image=@/path/to/image.jpg"

# 7. RAG 文件上传
curl -X POST "http://localhost:8080/api/v1/file/rag/upload" \
  -H "Authorization: Bearer <token>" \
  -F "session_id=<session_id>" \
  -F "file=@/path/to/knowledge.pdf"

# 8. 语音合成
curl -X POST "http://localhost:8080/api/v1/voice/tts" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"text":"你好世界"}'

# 9. 管理员登录
curl -X POST "http://localhost:8080/api/v1/admin/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 10. 获取所有模型配置（管理员）
curl "http://localhost:8080/api/v1/admin/models" \
  -H "Authorization: Bearer <admin_token>"
```

## MCP 服务

MCP (Model Context Protocol) 服务提供外部工具集成能力：

```bash
# 启动所有 MCP 服务
go run common/mcp/main.go

# 启动指定服务
go run common/mcp/main.go --services=weather,search,tts,translate

# 指定端口
go run common/mcp/main.go --address=localhost:9090
```

### 可用服务

| 服务 | 功能 |
|------|------|
| weather | 天气查询 |
| search | 网络搜索 |
| translate | 文本翻译 |
| tts | 语音合成 |

## 数据模型

### User (用户表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | varchar(20) | 用户ID (U+6位数字) |
| nickname | varchar(50) | 昵称 |
| email | varchar(100) | 邮箱 (唯一) |
| password | varchar(255) | 密码 (bcrypt加密) |

### Session (会话表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| session_id | varchar(36) | 会话ID (UUID) |
| user_id | varchar(20) | 用户ID |
| title | varchar(100) | 会话标题 |
| rag_file_id | varchar(36) | RAG文件ID |
| rag_file_name | varchar(255) | RAG文件名 |

### Message (消息表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| session_id | varchar(36) | 会话ID |
| user_id | varchar(20) | 用户ID |
| content | text | 消息内容 |
| is_user | bool | 是否用户消息 |

### Admin (管理员表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| admin_id | varchar(20) | 管理员ID |
| username | varchar(50) | 用户名 (唯一) |
| password | varchar(255) | 密码 (bcrypt加密) |

### AIModelConfig (AI模型配置表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| config_id | varchar(36) | 配置ID (UUID) |
| name | varchar(100) | 显示名称 |
| model_type | varchar(50) | 模型类型 (openai_compatible/ollama) |
| base_url | varchar(255) | API地址 |
| model_name | varchar(100) | 模型名称 |
| api_key | varchar(255) | API密钥 |
| description | text | 描述 |
| is_enabled | bool | 是否启用 |
| is_default | bool | 是否默认 |
| sort_order | int | 排序权重 |

## 开发

### 构建
```bash
go build -o nexus-ai main.go
```

### 测试
```bash
go test ./...
```

## License

MIT
