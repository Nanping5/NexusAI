# 用户模块 API 文档

> 基础路径: `/api/v1/user`
> 版本: v1.0.0
> 更新日期: 2026-03-28

---

## 目录

- [概述](#概述)
- [接口列表](#接口列表)
  - [发送验证码](#1-发送验证码)
  - [用户注册](#2-用户注册)
  - [用户登录](#3-用户登录)
  - [获取用户信息](#4-获取用户信息)
  - [更新昵称](#5-更新昵称)

---

## 概述

用户模块提供用户认证相关功能，包括注册、登录、信息获取等。

**认证方式**: JWT Bearer Token

**不需要认证的接口**:
- 发送验证码
- 用户注册
- 用户登录

**需要认证的接口**:
- 获取用户信息
- 更新昵称

---

## 接口列表

### 1. 发送验证码

发送注册验证码到指定邮箱。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/user/send_captcha` |
| Method | `GET` |
| 认证 | 无需认证 |

**请求参数 (Query):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| email | string | 是 | 邮箱地址，需符合邮箱格式 |

**请求示例:**

```bash
GET /api/v1/user/send_captcha?email=user@example.com
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "验证码已发送，请注意查收",
    "data": null
}
```

**失败响应:**

```json
// 参数错误
{
    "code": 2001,
    "msg": "请求参数错误",
    "data": null
}

// 服务繁忙（发送频繁）
{
    "code": 4001,
    "msg": "发送太频繁，请 60 秒后再试",
    "data": null
}
```

**说明:**
- 验证码有效期: 3分钟
- 发送频率限制: 60秒冷却时间
- 验证码存储在 Redis 中

---

### 2. 用户注册

使用邮箱和验证码注册新用户。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/user/register` |
| Method | `POST` |
| 认证 | 无需认证 |
| Content-Type | `application/json` |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| email | string | 是 | 邮箱地址，需符合邮箱格式 |
| password | string | 是 | 密码，至少6位 |
| captcha | string | 是 | 验证码，6位数字 |

**请求示例:**

```json
{
    "email": "user@example.com",
    "password": "123456",
    "captcha": "123456"
}
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
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

// 验证码错误
{
    "code": 2008,
    "msg": "验证码错误或已过期"
}

// 邮箱已存在
{
    "code": 2002,
    "msg": "邮箱已存在"
}

// 密码不合法
{
    "code": 2010,
    "msg": "密码不合法"
}
```

**说明:**
- 注册成功后自动生成 JWT Token
- 用户ID格式: U + 6位数字 (如: U000001)
- 密码使用 bcrypt 加密存储
- 昵称自动生成（格式：用户+6位随机数）

---

### 3. 用户登录

使用邮箱和密码登录。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/user/login` |
| Method | `POST` |
| 认证 | 无需认证 |
| Content-Type | `application/json` |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| email | string | 是 | 邮箱地址 |
| password | string | 是 | 密码 |

**请求示例:**

```json
{
    "email": "user@example.com",
    "password": "123456"
}
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
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

// 邮箱或密码错误
{
    "code": 2004,
    "msg": "邮箱或密码错误"
}

// 用户不存在
{
    "code": 2003,
    "msg": "用户不存在"
}
```

**说明:**
- 登录成功返回 JWT Token
- Token 有效期 24小时

---

### 4. 获取用户信息

获取当前登录用户的信息。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/user/info` |
| Method | `GET` |
| 认证 | **需要 JWT 认证** |

**请求头:**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer Token |

**请求示例:**

```bash
GET /api/v1/user/info
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "user_id": "U000001",
        "email": "user@example.com",
        "nickname": "用户123456"
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

// Token无效
{
    "code": 2006,
    "msg": "无效的Token"
}
```

---

### 5. 更新昵称

更新当前登录用户的昵称。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/user/nickname` |
| Method | `PUT` |
| 认证 | **需要 JWT 认证** |
| Content-Type | `application/json` |

**请求头:**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer Token |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| nickname | string | 是 | 新昵称 |

**请求示例:**

```bash
PUT /api/v1/user/nickname
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
    "nickname": "新昵称"
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
```
