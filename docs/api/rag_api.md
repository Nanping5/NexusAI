# RAG 文件模块 API 文档

> 基础路径: `/api/v1/file`
> 版本: v1.2.0
> 更新日期: 2026-03-28

---

## 目录

- [概述](#概述)
- [接口列表](#接口列表)
  - [上传 RAG 文件](#1-上传-rag-文件)
- [工作流程说明](#工作流程说明)

---

## 概述

RAG（Retrieval-Augmented Generation）文件模块提供知识库文件上传功能，支持将文档内容向量化存储到 Qdrant 向量数据库，用于 AI 对话时的知识检索。

**重要变更（v1.2.0）**：RAG 文件现在与 **Session（会话）** 绑定，每个会话可以独立上传一个知识库文件，实现会话级别的知识隔离。向量存储从 Redis 迁移到 Qdrant。

**认证方式**: JWT Bearer Token（**所有接口都需要认证**）

**请求头:**

```
Authorization: Bearer <token>
```

**支持的文件格式:**
- PDF
- TXT
- DOC/DOCX
- 其他文本格式文件

**存储机制:**
- 文件存储在服务器本地 `uploads/{userId}/{sessionID}/` 目录
- 文件内容向量化后存储在 Qdrant 中
- 每个会话只保留一个最新的文件

---

## 接口列表

### 1. 上传 RAG 文件

上传知识库文件并创建向量索引（绑定到指定会话）。

| 项目 | 内容 |
|------|------|
| URL | `/api/v1/file/rag/upload` |
| Method | `POST` |
| 认证 | **需要 JWT 认证** |
| Content-Type | `multipart/form-data` |

**请求参数 (Form-Data):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| file | file | 是 | 知识库文件 |
| session_id | string | 是 | 会话 ID |

**请求示例:**

```bash
POST /api/v1/file/rag/upload
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="session_id"

sess_abc123def456
--boundary
Content-Disposition: form-data; name="file"; filename="knowledge.pdf"
Content-Type: application/pdf

<二进制文件数据>
--boundary--
```

**cURL 示例:**

```bash
curl -X POST "http://localhost:8080/api/v1/file/rag/upload" \
  -H "Authorization: Bearer <token>" \
  -F "session_id=sess_abc123def456" \
  -F "file=@/path/to/knowledge.pdf"
```

**成功响应:**

```json
{
    "code": 1000,
    "msg": "success",
    "data": {
        "file_path": "uploads/U000001/sess_abc123def456/uuid.pdf"
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

// 参数错误（未上传文件或未提供 session_id）
{
    "code": 2001,
    "msg": "请求参数错误"
}

// 服务繁忙（文件保存失败或索引创建失败）
{
    "code": 4001,
    "msg": "服务繁忙"
}
```

**说明:**
- 上传新文件时会自动删除该会话的旧文件和旧索引
- 文件保存成功后会自动创建向量索引
- 每个会话只保留一个最新的知识库文件
- 删除会话时会自动清理对应的 RAG 文件和索引

---

## 工作流程说明

### 文件上传流程

```
1. 验证文件格式和大小
        ↓
2. 创建会话目录 uploads/{userId}/{sessionID}/
        ↓
3. 生成 UUID 文件名并保存文件
        ↓
4. 读取文件内容并分块处理
        ↓
5. 使用 Embedding 模型生成向量
        ↓
6. 将向量存入 Qdrant（带 session_id payload）
        ↓
7. 删除旧文件和旧索引（如有）
        ↓
8. 返回文件路径
```

### Qdrant 向量存储结构

**集合名称**: `nexus_rag`（可配置）

**向量参数:**
- 向量维度: 1024（可配置）
- 距离度量: Cosine

**Payload 结构:**
```json
{
    "content": "文档内容片段",
    "session_id": "会话ID",
    "source": "文件名",
    "chunk": "分块索引"
}
```

### RAG 检索流程

在 AI 对话时，系统会：

1. 根据会话 ID 查找对应的知识库文件
2. 将用户问题转换为向量
3. 在 Qdrant 中进行向量相似度搜索（按 session_id 过滤）
4. 获取最相关的文档内容
5. 将文档内容作为上下文传递给 AI 模型

---

## 技术细节

### 向量生成

- **Embedding 模型**: 可配置，通过 `config.toml` 的 `rag_embedding_model` 指定
- **向量维度**: 可配置，通过 `config.toml` 的 `rag_dimension` 指定
- **相似度算法**: 余弦相似度（Cosine）

### 文件存储

- 文件存储路径: `uploads/{userId}/{sessionID}/{uuid}.{ext}`
- `userId`: 用户ID，如 `U000001`
- `sessionID`: 会话ID，如 `sess_abc123def456`
- `uuid`: 随机生成的 UUID
- `ext`: 原始文件扩展名

### 文档分块

- 最大分块大小: 8000 字符
- 自动按段落和句子边界分块
- 保持内容完整性

### 错误处理

上传过程中的错误处理：

| 阶段 | 失败处理 |
|------|----------|
| 文件验证 | 直接返回错误 |
| 目录创建 | 直接返回错误 |
| 文件保存 | 清理已创建的文件 |
| 向量生成 | 删除已保存的文件 |

---

## 配置说明

### 环境变量

| 变量名 | 说明 |
|--------|------|
| `RAG_OPENAI_API_KEY` | OpenAI API Key（用于 Embedding 服务） |

### config.toml 配置

```toml
[rag_config]
rag_dimension = 1024
rag_chat_model_name = "deepseek-chat"
rag_doc_dir = "./Rag_docs"
rag_base_url = "https://api.openai.com/v1"
rag_embedding_model = "text-embedding-v3"

[qdrant_config]
host = "localhost"
port = 6333
collection_name = "nexus_rag"
```

---

## 数据库变更

### Session 表新增字段

```sql
ALTER TABLE sessions ADD COLUMN rag_file_id VARCHAR(36);
ALTER TABLE sessions ADD COLUMN rag_file_name VARCHAR(255);
```

---

## 更新日志

### v1.2.0 (2026-03-28)
- 向量存储从 Redis 迁移到 Qdrant
- 优化文档分块处理逻辑
- 更新配置项命名规范

### v1.1.0 (2026-03-25)
- RAG 文件与会话绑定，支持会话级别知识隔离
- 新增 Session 表字段：rag_file_id, rag_file_name

### v1.0.0 (2026-03-24)
- 初始版本发布
- 支持 PDF/TXT/DOCX 文件上传
- Redis 向量索引存储
