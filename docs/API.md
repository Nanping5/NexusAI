# NexusAi API 文档

> 基础路径: `/api/v1`
> 版本: v1.3.0
> 更新日期: 2026-03-28

---

## 目录

- [概述](#概述)
- [通用说明](#通用说明)
- [认证机制](#认证机制)
- [错误码说明](#错误码说明)
- [模块文档](#模块文档)
- [Redis 缓存架构](#redis-缓存架构)
- [上下文管理](#上下文管理)
- [接口限流](#接口限流)

---

## 概述

NexusAi 是一个基于 Go 语言和 Gin 框架的 AI 服务后端应用，提供用户认证、AI 对话、图像识别、RAG 检索增强、语音合成等核心功能。

**技术栈:**
- 编程语言: Go 1.25.4
- Web框架: Gin v1.12.0
- ORM: GORM v1.31.1
- 数据库: MySQL 8.0+
- 缓存: Redis 6.0+
- 向量数据库: Qdrant
- 消息队列: RabbitMQ 3.x
- 认证: JWT (golang-jwt/v4)
- AI框架: CloudWeGo Eino v0.8.5
- AI推理: 本地模型推理
- 语音服务: 阿里云 NLS
- MCP协议: mark3labs/mcp-go v0.45.0
- 配置管理: Viper (TOML)
- 日志: Zap v1.27.1

---

## 通用说明

### 请求格式

- Content-Type: `application/json` (POST/PUT 请求)
- 字符编码: UTF-8
- 文件上传: `multipart/form-data`

### 响应格式

所有接口统一返回 JSON 格式数据：

```json
{
    "code": 1000,
    "msg": "success",
    "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 业务状态码 |
| msg | string | 提示信息 |
| data | any | 返回数据（可选） |

---

## 认证机制

需要认证的接口需在请求头中携带 JWT Token：

```
Authorization: Bearer <token>
```

**Token 获取方式:**
- 用户注册接口返回
- 用户登录接口返回
- 管理员登录接口返回

**Token 有效期:** 24小时

**无需认证的接口:**
- `GET /api/v1/user/send_captcha` - 发送验证码
- `POST /api/v1/user/register` - 用户注册
- `POST /api/v1/user/login` - 用户登录
- `POST /api/v1/admin/login` - 管理员登录

---

## 错误码说明

### 通用错误码 (1xxx)

| 错误码 | 说明 |
|--------|------|
| 1000 | 成功 |

### 客户端错误 (2xxx)

| 错误码 | 说明 |
|--------|------|
| 2001 | 请求参数错误 |
| 2002 | 邮箱已存在 |
| 2003 | 用户不存在 |
| 2004 | 邮箱或密码错误 |
| 2005 | 两次密码不一致 |
| 2006 | 无效的Token |
| 2007 | 用户未登录 |
| 2008 | 验证码错误或已过期 |
| 2009 | 记录不存在 |
| 2010 | 密码不合法 |
| 2011 | 登录失败 |

### 权限错误 (3xxx)

| 错误码 | 说明 |
|--------|------|
| 3000 | 用户未登录或Token无效 |
| 3001 | 权限不足 |

### 服务端错误 (4xxx)

| 错误码 | 说明 |
|--------|------|
| 4001 | 服务繁忙 |

### AI 模型错误 (5xxx)

| 错误码 | 说明 |
|--------|------|
| 5001 | 模型不存在 |
| 5002 | 无法打开模型 |
| 5003 | 模型运行失败 |

### 其他服务错误 (6xxx)

| 错误码 | 说明 |
|--------|------|
| 6001 | 语音服务失败 |

---

## 模块文档

### [用户模块](api/user_api.md)

用户认证相关接口，包括注册、登录、信息管理。

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/user/send_captcha` | GET | ❌ | 发送验证码 |
| `/user/register` | POST | ❌ | 用户注册 |
| `/user/login` | POST | ❌ | 用户登录 |
| `/user/info` | GET | ✅ | 获取用户信息 |
| `/user/nickname` | PUT | ✅ | 更新昵称 |

📄 [详细文档](api/user_api.md)

---

### [AI 会话模块](api/ai_api.md)

AI 对话相关接口，支持普通模式和流式模式（SSE）。

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/ai/sessions` | GET | ✅ | 获取会话列表 |
| `/ai/session` | DELETE | ✅ | 删除会话 |
| `/ai/history` | GET | ✅ | 获取聊天历史 |
| `/ai/models` | GET | ✅ | 获取可用模型列表 |
| `/ai/session/create` | POST | ✅ | 创建会话并发送消息 |
| `/ai/session/create/stream` | POST | ✅ | 创建会话并流式发送消息 |
| `/ai/chat` | POST | ✅ | 发送消息（已存在会话） |
| `/ai/chat/stream` | POST | ✅ | 流式发送消息（已存在会话） |

📄 [详细文档](api/ai_api.md)

---

### [图像识别模块](api/image_api.md)

基于 MobileNetV2 的图像分类识别接口。

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/image/recognize` | POST | ✅ | 图像识别 |

📄 [详细文档](api/image_api.md)

---

### [RAG 文件模块](api/rag_api.md)

知识库文件上传与向量索引管理接口。

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/file/rag/upload` | POST | ✅ | 上传 RAG 文件 |

📄 [详细文档](api/rag_api.md)

---

### [语音合成模块](api/tts_api.md)

基于阿里云智能语音服务的文本转语音接口。

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/voice/tts` | POST | ✅ | 语音合成 |
| `/voice/tts/status` | GET | ✅ | 服务状态 |

📄 [详细文档](api/tts_api.md)

---

### [管理员模块](api/admin_api.md)

管理员登录与 AI 模型配置管理接口。

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/admin/login` | POST | ❌ | 管理员登录 |
| `/admin/info` | GET | ✅ (AdminJWT) | 获取管理员信息 |
| `/admin/models` | GET | ✅ (AdminJWT) | 获取所有模型配置 |
| `/admin/models/:config_id` | GET | ✅ (AdminJWT) | 获取单个模型配置 |
| `/admin/models` | POST | ✅ (AdminJWT) | 创建模型配置 |
| `/admin/models/:config_id` | PUT | ✅ (AdminJWT) | 更新模型配置 |
| `/admin/models/:config_id` | DELETE | ✅ (AdminJWT) | 删除模型配置 |
| `/admin/models/:config_id/default` | PUT | ✅ (AdminJWT) | 设置默认模型 |
| `/admin/models/:config_id/toggle` | PUT | ✅ (AdminJWT) | 启用/禁用模型 |

📄 [详细文档](api/admin_api.md)

---

## 快速开始

### 1. 注册账号

```bash
# 1. 发送验证码
curl "http://localhost:8080/api/v1/user/send_captcha?email=user@example.com"

# 2. 注册（使用收到的验证码）
curl -X POST "http://localhost:8080/api/v1/user/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"123456","captcha":"123456"}'
```

### 2. 登录获取 Token

```bash
curl -X POST "http://localhost:8080/api/v1/user/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"123456"}'
```

### 3. 使用 Token 调用 API

```bash
# 获取可用模型列表
curl "http://localhost:8080/api/v1/ai/models" \
  -H "Authorization: Bearer <token>"

# AI 对话
curl -X POST "http://localhost:8080/api/v1/ai/session/create" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_question":"你好","model_config_id":"<config_id>"}'

# 图像识别
curl -X POST "http://localhost:8080/api/v1/image/recognize" \
  -H "Authorization: Bearer <token>" \
  -F "image=@/path/to/image.jpg"

# RAG 文件上传
curl -X POST "http://localhost:8080/api/v1/file/rag/upload" \
  -H "Authorization: Bearer <token>" \
  -F "session_id=<session_id>" \
  -F "file=@/path/to/knowledge.pdf"

# 语音合成
curl -X POST "http://localhost:8080/api/v1/voice/tts" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"text":"你好世界"}'
```

---

## MCP 服务

MCP (Model Context Protocol) 服务提供外部工具集成能力，支持 AI Agent 调用。

### 可用服务

| 服务 | 功能 | 说明 |
|------|------|------|
| weather | 天气查询 | 查询指定城市的天气信息 |
| search | 网络搜索 | 搜索互联网信息 |
| translate | 文本翻译 | 多语言文本翻译 |
| tts | 语音合成 | 文本转语音 |

### 启动 MCP 服务

```bash
# 启动所有服务（默认端口 8081）
go run common/mcp/main.go

# 启动指定服务
go run common/mcp/main.go --services=weather,search

# 指定端口
go run common/mcp/main.go --address=localhost:9090
```

---

## 更新日志

### v1.3.0 (2026-03-28)
- **Redis 缓存扩展**：
  - 模型配置重构为纯 Redis 缓存
  - 新增会话列表缓存
  - 新增消息历史缓存
  - 新增用户信息缓存
  - 新增在线用户追踪
- **上下文管理**：
  - 新增滑动窗口上下文裁剪
  - 新增 Token 统计功能
  - 新增可配置的上下文策略
- **接口限流**：
  - 新增基于 Redis 的接口限流中间件
  - AI 创建会话接口：每分钟最多 10 次
  - AI 聊天接口：每分钟最多 30 次
- **Bug 修复**：
  - 修复模型配置缓存 APIKey 丢失问题

### v1.2.0 (2026-03-28)
- 新增管理员模块：管理员登录、模型配置 CRUD
- 新增动态模型配置功能
- 新增获取可用模型列表接口
- AI 对话支持通过 model_config_id 指定模型
- RAG 向量存储从 Redis 迁移到 Qdrant

### v1.1.0 (2026-03-27)
- 新增 MCP 服务集成：天气、搜索、翻译、TTS
- 新增 Agent 模式支持（ReAct）
- AI 框架升级为 CloudWeGo Eino

### v1.0.1 (2026-03-26)
- 新增语音合成模块：TTS 文本转语音功能

### v1.0.0 (2026-03-25)
- 初始版本发布
- 用户模块：注册、登录、信息管理
- AI 会话模块：对话、流式响应、历史记录
- 图像识别模块：基于 MobileNetV2 的图像分类
- RAG 文件模块：知识库文件上传与向量索引

---

## Redis 缓存架构

### Key 设计规范

所有 Redis Key 遵循统一命名规范：`nexus:{type}:{id}`

| Key 模式 | 说明 | TTL |
|----------|------|-----|
| `nexus:session:{session_id}` | 用户登录会话 | 7 天 |
| `nexus:user:{user_id}` | 用户信息缓存 | 30 分钟 |
| `nexus:sessions:{user_id}` | 用户会话列表 | 5 分钟 |
| `nexus:history:{session_id}` | 聊天历史缓存 | 1 小时 |
| `nexus:model:{config_id}` | 模型配置缓存 | 5 分钟 |
| `nexus:online` | 在线用户集合 | 5 分钟 |
| `nexus:limit:{user_id}:{api}` | 接口限流计数 | 1 分钟 |
| `nexus:captcha:{email}` | 验证码 | 3 分钟 |

### 缓存策略

#### 模型配置缓存
- 纯 Redis 缓存，无内存缓存
- 包含 APIKey 等敏感信息
- 管理后台修改后自动失效

#### 会话列表缓存
- 用户获取会话列表时优先读取 Redis
- 创建/删除会话时自动失效
- 支持从数据库回源

#### 消息历史缓存
- 三级查询：内存 → Redis → 数据库
- 追加式写入，支持历史恢复

#### 用户信息缓存
- 读取时优先 Redis
- 更新昵称后自动失效

---

## 上下文管理

### 配置说明

```toml
[ai_config]
max_context_messages = 20     # 最大保留消息轮次
max_context_tokens = 16000    # 最大上下文 Token 数
context_strategy = "sliding_window"  # 上下文策略
```

### 策略说明

| 策略 | 说明 |
|------|------|
| `sliding_window` | 滑动窗口：保留最近 N 轮对话 |
| `summary` | 摘要压缩：对早期对话生成摘要（开发中） |

### Token 统计

系统会实时统计每次对话的 Token 使用情况：
- **输入 Token**：包含系统提示词和历史消息
- **输出 Token**：AI 响应内容

可通过日志查看 Token 消耗：
```
token stats  input=1234  output=567
```

---

## 接口限流

### 限流规则

| 接口 | 限制 | 时间窗口 |
|------|------|----------|
| `/ai/session/create` | 10 次 | 1 分钟 |
| `/ai/session/create/stream` | 10 次 | 1 分钟 |
| `/ai/chat` | 30 次 | 1 分钟 |
| `/ai/chat/stream` | 30 次 | 1 分钟 |

### 限流响应

当请求超过限制时，返回 HTTP 429：

```json
{
    "code": 429,
    "msg": "请求过于频繁，请稍后再试"
}
```
