# 管理员模块 API 文档

> 基础路径: `/api/v1/admin`
> 版本: v1.2.0
> 更新日期: 2026-03-28

---

## 目录

- [概述](#概述)
- [接口列表](#接口列表)
  - [管理员登录](#1-管理员登录)
  - [获取管理员信息](#2-获取管理员信息)
  - [获取所有模型配置](#3-获取所有模型配置)
  - [获取单个模型配置](#4-获取单个模型配置)
  - [创建模型配置](#5-创建模型配置)
  - [更新模型配置](#6-更新模型配置)
  - [删除模型配置](#7-删除模型配置)
  - [设置默认模型](#8-设置默认模型)
  - [启用/禁用模型](#9-启用禁用模型)

---

## 概述

管理员模块提供管理员登录与 AI 模型配置管理功能。管理员可以动态配置 AI 模型，支持多种 LLM 提供商。

**认证方式**: JWT Bearer Token

**不需要认证的接口**:
- 管理员登录

**需要认证的接口**:
- 获取管理员信息
- 所有模型配置管理接口

**注意**: 管理员接口使用独立的 AdminJWT 中间件进行认证。

---

## 接口列表

### 1. 管理员登录

管理员登录并获取 Token。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/admin/login` |
| Method | `POST` |
| 认证 | 无需认证 |
| Content-Type | `application/json` |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 是 | 管理员用户名 |
| password | string | 是 | 密码 |

**请求示例:**

```json
{
    "username": "admin",
    "password": "admin123"
}
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "admin_id": "A000001",
        "username": "admin"
    }
}
```

**失败响应:**

```json
// 参数错误
{
    "code": 2001,
    "msg": "请求参数错误"
}

// 登录失败
{
    "code": 2011,
    "msg": "登录失败"
}
```

**说明:**
- 登录成功返回 JWT Token
- Token 有效期 24 小时

---

### 2. 获取管理员信息

获取当前登录管理员的信息。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/admin/info` |
| Method | `GET` |
| 认证 | **需要 AdminJWT 认证** |

**请求头:**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer Token |

**请求示例:**

```bash
GET /api/v1/admin/info
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "admin_id": "A000001",
        "username": "admin"
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

// Token 无效
{
    "code": 2006,
    "msg": "无效的 Token"
}
```

---

### 3. 获取所有模型配置

获取所有 AI 模型配置列表。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/admin/models` |
| Method | `GET` |
| 认证 | **需要 AdminJWT 认证** |

**请求头:**

```
Authorization: Bearer <token>
```

**请求示例:**

```bash
GET /api/v1/admin/models
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
                "base_url": "https://api.openai.com/v1",
                "model_name": "gpt-4o",
                "description": "OpenAI GPT-4o 模型",
                "is_enabled": true,
                "is_default": true,
                "sort_order": 1
            },
            {
                "config_id": "550e8400-e29b-41d4-a716-446655440001",
                "name": "DeepSeek-Chat",
                "model_type": "openai_compatible",
                "base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
                "model_name": "deepseek-chat",
                "description": "深度求索聊天模型",
                "is_enabled": true,
                "is_default": false,
                "sort_order": 2
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

### 4. 获取单个模型配置

获取指定 AI 模型配置的详细信息。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/admin/models/:config_id` |
| Method | `GET` |
| 认证 | **需要 AdminJWT 认证** |

**请求头:**

```
Authorization: Bearer <token>
```

**路径参数:**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| config_id | string | 是 | 模型配置 ID |

**请求示例:**

```bash
GET /api/v1/admin/models/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "config_id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "GPT-4o",
        "model_type": "openai_compatible",
        "base_url": "https://api.openai.com/v1",
        "model_name": "gpt-4o",
        "description": "OpenAI GPT-4o 模型",
        "is_enabled": true,
        "is_default": true,
        "sort_order": 1
    }
}
```

**失败响应:**

```json
// 记录不存在
{
    "code": 2009,
    "msg": "记录不存在"
}
```

---

### 5. 创建模型配置

创建新的 AI 模型配置。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/admin/models` |
| Method | `POST` |
| 认证 | **需要 AdminJWT 认证** |
| Content-Type | `application/json` |

**请求头:**

```
Authorization: Bearer <token>
```

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | 模型显示名称 |
| model_type | string | 是 | 模型类型：`openai_compatible` |
| base_url | string | 是 | API 请求地址 |
| model_name | string | 是 | 模型名称 |
| api_key | string | 否 | API 密钥 |
| description | string | 否 | 模型描述 |
| is_enabled | bool | 否 | 是否启用（默认 true） |
| is_default | bool | 否 | 是否为默认模型（默认 false） |
| sort_order | int | 否 | 排序权重（默认 0） |

**请求示例:**

```json
{
    "name": "GPT-4o",
    "model_type": "openai_compatible",
    "base_url": "https://api.openai.com/v1",
    "model_name": "gpt-4o",
    "api_key": "sk-...",
    "description": "OpenAI GPT-4o 模型",
    "is_enabled": true,
    "is_default": true,
    "sort_order": 1
}
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "config_id": "550e8400-e29b-41d4-a716-446655440000"
    }
}
```

**失败响应:**

```json
// 参数错误
{
    "code": 2001,
    "msg": "请求参数错误"
}
```

---

### 6. 更新模型配置

更新指定的 AI 模型配置。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/admin/models/:config_id` |
| Method | `PUT` |
| 认证 | **需要 AdminJWT 认证** |
| Content-Type | `application/json` |

**请求头:**

```
Authorization: Bearer <token>
```

**路径参数:**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| config_id | string | 是 | 模型配置 ID |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 否 | 模型显示名称 |
| model_type | string | 否 | 模型类型：`openai_compatible` |
| base_url | string | 否 | API 请求地址 |
| model_name | string | 否 | 模型名称 |
| api_key | string | 否 | API 密钥（空表示不更新） |
| description | string | 否 | 模型描述 |
| is_enabled | bool | 否 | 是否启用 |
| is_default | bool | 否 | 是否为默认模型 |
| sort_order | int | 否 | 排序权重 |

**请求示例:**

```json
{
    "name": "GPT-4o Turbo",
    "base_url": "https://api.openai.com/v1",
    "is_enabled": true
}
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
// 参数错误
{
    "code": 2001,
    "msg": "请求参数错误"
}
```

**说明:**
- 只提交需要更新的字段
- 未提交的字段保持不变
- `api_key` 为空字符串表示不更新该字段

---

### 7. 删除模型配置

删除指定的 AI 模型配置。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/admin/models/:config_id` |
| Method | `DELETE` |
| 认证 | **需要 AdminJWT 认证** |

**请求头:**

```
Authorization: Bearer <token>
```

**路径参数:**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| config_id | string | 是 | 模型配置 ID |

**请求示例:**

```bash
DELETE /api/v1/admin/models/550e8400-e29b-41d4-a716-446655440000
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
// 参数错误
{
    "code": 2001,
    "msg": "请求参数错误"
}
```

---

### 8. 设置默认模型

将指定模型设置为默认模型。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/admin/models/:config_id/default` |
| Method | `PUT` |
| 认证 | **需要 AdminJWT 认证** |

**请求头:**

```
Authorization: Bearer <token>
```

**路径参数:**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| config_id | string | 是 | 模型配置 ID |

**请求示例:**

```bash
PUT /api/v1/admin/models/550e8400-e29b-41d4-a716-446655440000/default
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

**说明:**
- 设置默认模型后，其他模型的 `is_default` 会自动设为 false
- 用户对话时如果不指定模型，会使用默认模型

---

### 9. 启用/禁用模型

启用或禁用指定的 AI 模型配置。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/admin/models/:config_id/toggle` |
| Method | `PUT` |
| 认证 | **需要 AdminJWT 认证** |
| Content-Type | `application/json` |

**请求头:**

```
Authorization: Bearer <token>
```

**路径参数:**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| config_id | string | 是 | 模型配置 ID |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| is_enabled | bool | 是 | 是否启用 |

**请求示例:**

```json
{
    "is_enabled": false
}
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": null
}
```

**说明:**
- 禁用的模型不会出现在用户可选模型列表中
- 可以批量禁用多个模型

---

## 模型类型说明

### openai_compatible

兼容 OpenAI API 格式的模型服务。

**配置示例:**

```json
{
    "name": "DeepSeek-Chat",
    "model_type": "openai_compatible",
    "base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
    "model_name": "deepseek-chat",
    "api_key": "sk-..."
}
```

**支持的提供商:**
- OpenAI
- 阿里云 DashScope (DeepSeek, Qwen 等)
- 智谱 AI
- 月之暗面 (Moonshot)
- 零一万物
- 其他兼容 OpenAI API 格式的服务

---

## 更新日志

### v1.2.0 (2026-03-28)

- 新增管理员模块
- 新增管理员登录接口
- 新增 AI 模型配置 CRUD 功能
- 支持动态添加/删除/启用/禁用模型
- 支持设置默认模型
- 仅支持 OpenAI 兼容格式的 API
