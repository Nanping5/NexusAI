# 语音合成模块 API 文档

> 基础路径: `/api/v1/voice`
> 版本: v1.0.0
> 更新日期: 2026-03-28

---

## 目录

- [概述](#概述)
- [接口列表](#接口列表)
  - [语音合成](#1-语音合成)
  - [服务状态](#2-服务状态)

---

## 概述

语音合成模块提供文本转语音（TTS）功能，将文本转换为自然流畅的语音。

**认证方式**: JWT Bearer Token（**所有接口都需要认证**）

**请求头:**

```
Authorization: Bearer <token>
```

**服务配置:**

| 配置项 | 值 |
|--------|-----|
| 服务商 | 阿里云智能语音服务 |
| 发音人 | 知小夏 (zhixiaoxia) |
| 音频格式 | WAV |
| 采样率 | 16000 Hz |

---

## 接口列表

### 1. 语音合成

将文本转换为语音文件。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/voice/tts` |
| Method | `POST` |
| 认证 | **需要 JWT 认证** |
| Content-Type | `application/json` |

**请求体 (JSON):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| text | string | 是 | 待合成的文本内容（最大 5000 字符） |

**请求示例:**

```bash
POST /api/v1/voice/tts
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
    "text": "你好世界，欢迎来到NexusAI"
}
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "audio_url": "/tmp/tts_1711463820123456789.wav",
        "format": "wav"
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

// 语音服务失败
{
    "code": 6001,
    "msg": "语音服务失败"
}
```

---

### 2. 服务状态

获取语音合成服务的当前状态。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/voice/tts/status` |
| Method | `GET` |
| 认证 | **需要 JWT 认证** |

**请求示例:**

```bash
GET /api/v1/voice/tts/status
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "format": "wav",
        "sample_rate": 16000,
        "voice": "zhixiaoxia",
        "status": "ok"
    }
}
```

---

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 1000 | 成功 |
| 2001 | 请求参数错误 |
| 2007 | 用户未登录 |
| 4001 | 服务繁忙 |
| 6001 | 语音服务失败 |

---

## 使用示例

### JavaScript (Fetch API)

```javascript
// 语音合成
async function synthesizeSpeech(text) {
    const response = await fetch('/api/v1/voice/tts', {
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('token'),
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ text })
    });

    const result = await response.json();
    if (result.code === 1000) {
        // 播放音频
        const audio = new Audio(result.data.audio_url);
        audio.play();
    }
}

// 使用示例
synthesizeSpeech('你好，这是语音合成测试');
```

### Python (requests)

```python
import requests

def synthesize_speech(token, text):
    url = 'http://localhost:8080/api/v1/voice/tts'
    headers = {
        'Authorization': f'Bearer {token}',
        'Content-Type': 'application/json'
    }
    data = {'text': text}

    response = requests.post(url, json=data, headers=headers)
    return response.json()

# 使用示例
result = synthesize_speech('your_token', '你好世界')
print(result)
```

---

## 配置说明

### config.toml 配置

```toml
[voice_service_config]
access_key_id = "your-access-key-id"
access_key_secret = "your-access-key-secret"
app_key = "your-app-key"
```

### 阿里云语音服务配置

1. 登录阿里云控制台
2. 开通智能语音服务
3. 创建项目获取 AppKey
4. 创建 AccessKey 获取 AccessKey ID 和 AccessKey Secret

---

## 注意事项

1. **文本长度限制**: 单次请求文本最大 5000 字符
2. **音频格式**: 当前仅支持 WAV 格式输出
3. **发音人**: 当前使用固定发音人「知小夏」，暂不支持切换
4. **文件存储**: 合成的音频文件存储在服务器临时目录，建议及时下载使用
5. **服务依赖**: 需要配置阿里云智能语音服务相关凭证
