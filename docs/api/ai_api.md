# AI 会话模块 API 文档

> 基础路径: `/api/v1/ai`
> 版本: v1.2.0
> 更新日期: 2026-03-28

---

## 目录

- [概述](#概述)
- [接口列表](#接口列表)
  - [获取会话列表](#1-获取会话列表)
  - [删除会话](#2-删除会话)
  - [获取聊天历史](#3-获取聊天历史)
  - [获取可用模型列表](#4-获取可用模型列表)
  - [创建会话并发送消息](#5-创建会话并发送消息)
  - [创建会话并流式发送消息](#6-创建会话并流式发送消息)
  - [发送消息（已存在会话）](#7-发送消息已存在会话)
  - [流式发送消息（已存在会话）](#8-流式发送消息已存在会话)

---

## 概述

AI 会话模块提供 AI 对话功能，支持普通模式和流式模式（SSE）。

**认证方式**: JWT Bearer Token（**所有接口都需要认证**）

**请求头:**

```
Authorization: Bearer <token>
```

**支持的模型类型:**
- 仅支持 `openai_compatible` 类型（兼容 OpenAI API 格式）
- 通过管理后台动态配置模型

---

## 接口列表

### 1. 获取会话列表

获取当前用户的所有聊天会话。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/ai/sessions` |
| Method | `GET` |
| 认证 | **需要 JWT 认证** |

**请求示例:**

```bash
GET /api/v1/ai/sessions
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "sessions": [
            {
                "id": "550e8400-e29b-41d4-a716-446655440000",
                "title": "会话标题",
                "created_at": "2026-03-24T10:00:00Z",
                "updated_at": "2026-03-24T10:30:00Z"
            }
        ]
    }
}
```

**失败响应:**

```json
// 未登录
{
    "code": 2007,
    "msg": "用户未登录"
}

// 服务繁忙
{
    "code": 4001,
    "msg": "服务繁忙"
}
```

---

### 2. 删除会话

删除指定会话及其聊天记录。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/ai/session` |
| Method | `DELETE` |
| 认证 | **需要 JWT 认证** |

**请求参数 (Query):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_id | string | 是 | 会话ID |

**请求示例:**

```bash
DELETE /api/v1/ai/session?session_id=550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": null
}
```

**失败响应:**

```json
// 未登录
{
    "code": 2007,
    "msg": "用户未登录"
}

// 参数错误
{
    "code": 2001,
    "msg": "请求参数错误"
}

// 记录不存在
{
    "code": 2009,
    "msg": "记录不存在"
}
```

---

### 3. 获取聊天历史

获取指定会话的聊天历史记录。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/ai/history` |
| Method | `GET` |
| 认证 | **需要 JWT 认证** |

**请求参数 (Query):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_id | string | 是 | 会话ID |

**请求示例:**

```bash
GET /api/v1/ai/history?session_id=550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "messages": [
            {
                "id": "msg-uuid-1",
                "content": "用户消息内容",
                "is_user": true,
                "created_at": "2026-03-24T10:00:00Z"
            },
            {
                "id": "msg-uuid-2",
                "content": "AI 回复内容",
                "is_user": false,
                "created_at": "2026-03-24T10:00:05Z"
            }
        ]
    }
}
```

**失败响应:**

```json
// 未登录
{
    "code": 2007,
    "msg": "用户未登录"
}

// 参数错误
{
    "code": 2001,
    "msg": "请求参数错误"
}

// 记录不存在
{
    "code": 2009,
    "msg": "记录不存在"
}
```

---

### 4. 获取可用模型列表

获取当前用户可用的 AI 模型配置列表。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/ai/models` |
| Method | `GET` |
| 认证 | **需要 JWT 认证** |

**请求示例:**

```bash
GET /api/v1/ai/models
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "models": [
            {
                "config_id": "550e8400-e29b-41d4-a716-446655440000",
                "name": "GPT-4o",
                "model_type": "openai_compatible",
                "is_default": true
            },
            {
                "config_id": "550e8400-e29b-41d4-a716-446655440001",
                "name": "DeepSeek-Chat",
                "model_type": "openai_compatible",
                "is_default": false
            }
        ]
    }
}
```

**失败响应:**

```json
// 未登录
{
    "code": 2007,
    "msg": "用户未登录"
}

// 服务繁忙
{
    "code": 4001,
    "msg": "服务繁忙"
}
```

---

### 5. 创建会话并发送消息

创建新会话并发送第一条消息，返回 AI 回复。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/ai/session/create` |
| Method | `POST` |
| 认证 | **需要 JWT 认证** |
| Content-Type | `application/json` |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_question | string | 是 | 用户消息内容 |
| config_id | string | 否 | 模型配置 ID，不传则使用默认模型 |
| use_agent | bool | 否 | 是否使用 Agent 模式（默认 false） |

**请求示例:**

```bash
POST /api/v1/ai/session/create
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
    "user_question": "你好，请介绍一下自己",
    "config_id": "550e8400-e29b-41d4-a716-446655440000",
    "use_agent": false
}
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "session_id": "550e8400-e29b-41d4-a716-446655440000",
        "ai_message": {
            "id": "msg-uuid",
            "content": "AI 回复内容",
            "created_at": "2026-03-24T10:00:05Z"
        }
    }
}
```

**失败响应:**

```json
// 未登录
{
    "code": 2007,
    "msg": "用户未登录"
}

// 参数错误
{
    "code": 2001,
    "msg": "请求参数错误"
}

// 模型不存在
{
    "code": 5001,
    "msg": "模型不存在"
}

// 模型运行失败
{
    "code": 5003,
    "msg": "模型运行失败"
}
```

---

### 6. 创建会话并流式发送消息

创建新会话并流式返回 AI 回复（SSE）。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/ai/session/create/stream` |
| Method | `POST` |
| 认证 | **需要 JWT 认证** |
| Content-Type | `application/json` |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_question | string | 是 | 用户消息内容 |
| config_id | string | 否 | 模型配置 ID，不传则使用默认模型 |
| use_agent | bool | 否 | 是否使用 Agent 模式（默认 false） |

**响应头:**

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**SSE 数据格式:**

```
data: {"session_id": "550e8400-e29b-41d4-a716-446655440000"}

data: AI 回复片段1

data: AI 回复片段2

data: [DONE]
```

**前端示例 (JavaScript):**

```javascript
fetch('/api/v1/ai/session/create/stream', {
    method: 'POST',
    headers: {
        'Authorization': 'Bearer <token>',
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
        user_question: '你好',
        config_id: '<config_id>',
        use_agent: false
    })
}).then(response => {
    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    function read() {
        reader.read().then(({ done, value }) => {
            if (done) return;
            const text = decoder.decode(value);
            // 解析 SSE 数据
            console.log(text);
            read();
        });
    }
    read();
});
```

---

### 7. 发送消息（已存在会话）

在已存在的会话中发送消息，返回 AI 回复。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/ai/chat` |
| Method | `POST` |
| 认证 | **需要 JWT 认证** |
| Content-Type | `application/json` |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_id | string | 是 | 会话ID |
| user_question | string | 是 | 用户消息内容 |
| config_id | string | 否 | 模型配置 ID，不传则使用默认模型 |
| use_agent | bool | 否 | 是否使用 Agent 模式（默认 false） |

**请求示例:**

```bash
POST /api/v1/ai/chat
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_question": "继续上一话题",
    "config_id": "550e8400-e29b-41d4-a716-446655440000",
    "use_agent": false
}
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "ai_message": {
            "id": "msg-uuid",
            "content": "AI 回复内容",
            "created_at": "2026-03-24T10:05:00Z"
        }
    }
}
```

**失败响应:**

```json
// 未登录
{
    "code": 2007,
    "msg": "用户未登录"
}

// 参数错误
{
    "code": 2001,
    "msg": "请求参数错误"
}

// 记录不存在
{
    "code": 2009,
    "msg": "记录不存在"
}

// 模型运行失败
{
    "code": 5003,
    "msg": "模型运行失败"
}
```

---

### 8. 流式发送消息（已存在会话）

在已存在的会话中流式发送消息（SSE）。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/ai/chat/stream` |
| Method | `POST` |
| 认证 | **需要 JWT 认证** |
| Content-Type | `application/json` |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_id | string | 是 | 会话ID |
| user_question | string | 是 | 用户消息内容 |
| config_id | string | 否 | 模型配置 ID，不传则使用默认模型 |
| use_agent | bool | 否 | 是否使用 Agent 模式（默认 false） |

**响应格式:** Server-Sent Events (SSE)，同接口 #6

---

## 更新日志

### v1.2.0 (2026-03-28)
- 新增 `config_id` 参数，支持指定具体模型配置
- 新增 `use_agent` 参数，支持 Agent 模式
- 新增获取可用模型列表接口
- 模型类型从固定改为动态配置
- 仅支持 OpenAI 兼容格式的 API

### v1.0.0 (2026-03-24)
- 初始版本发布
- 支持基本对话功能和流式响应
