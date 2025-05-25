# 火山引擎即梦AI图像生成API文档

## 概述

本API提供基于火山引擎即梦AI的图像生成服务，采用纯异步模式，通过taskId确保用户出图稳定性。

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: 暂无（开发阶段）
- **数据格式**: JSON
- **字符编码**: UTF-8

## API接口

### 1. 创建图像生成任务

**接口地址**: `POST /ai/image/task`

**请求参数**:
```json
{
  "prompt": "一只可爱的小猫咪，坐在花园里",
  "user_id": "user123",
  "model": "doubao-seedream-3.0-t2i",
  "size": "1024x1024",
  "quality": "standard",
  "style": "anime",
  "n": 1
}
```

**参数说明**:
- `prompt` (string, 必填): 图像描述文本
- `user_id` (string, 必填): 用户ID
- `model` (string, 可选): 模型名称，默认为 "doubao-seedream-3.0-t2i"
- `size` (string, 可选): 图像尺寸，如 "1024x1024"
- `quality` (string, 可选): 图像质量，如 "standard", "hd"
- `style` (string, 可选): 图像风格，如 "anime", "realistic"
- `n` (int, 可选): 生成图像数量，默认为1

**响应示例**:
```json
{
  "success": true,
  "data": {
    "task_id": "67890abcdef",
    "status": "pending",
    "provider": "volcengine_jimeng",
    "external_task_id": "volcengine_img_1234567890"
  },
  "message": "图像生成任务创建成功"
}
```

### 2. 查询任务结果

**接口地址**: `GET /ai/image/result/{task_id}`

**路径参数**:
- `task_id`: 任务ID

**响应示例**:

**处理中**:
```json
{
  "success": true,
  "data": {
    "task_id": "67890abcdef",
    "status": "pending",
    "message": "任务处理中，请稍后查询",
    "created": "2023-12-01T10:00:00Z"
  }
}
```

**任务完成**:
```json
{
  "success": true,
  "data": {
    "task_id": "67890abcdef",
    "status": "completed",
    "image_url": "https://example.com/generated-image.jpg",
    "created": "2023-12-01T10:00:00Z"
  },
  "message": "任务完成"
}
```

**任务失败**:
```json
{
  "error": "任务执行失败",
  "message": "具体错误信息",
  "data": {
    "task_id": "67890abcdef",
    "status": "failed",
    "created": "2023-12-01T10:00:00Z"
  }
}
```

### 3. 获取用户图像任务列表

**接口地址**: `GET /ai/image/tasks`

**查询参数**:
- `user_id` (string, 必填): 用户ID
- `limit` (int, 可选): 每页数量，默认20，最大100
- `offset` (int, 可选): 偏移量，默认0

**请求示例**:
```
GET /ai/image/tasks?user_id=user123&limit=10&offset=0
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "task_id": "67890abcdef",
        "status": "completed",
        "image_url": "https://example.com/image1.jpg",
        "created": "2023-12-01T10:00:00Z"
      },
      {
        "task_id": "67890abcdeg",
        "status": "pending",
        "created": "2023-12-01T09:30:00Z"
      }
    ],
    "limit": 10,
    "offset": 0,
    "count": 2
  }
}
```

### 4. 删除图像任务

**接口地址**: `DELETE /ai/image/task/{task_id}`

**路径参数**:
- `task_id`: 任务ID

**响应示例**:
```json
{
  "success": true,
  "message": "任务删除成功"
}
```

## 任务状态说明

- `pending`: 任务已创建，等待处理
- `processing`: 任务处理中
- `completed`: 任务完成
- `failed`: 任务失败

## 错误码说明

- `400`: 请求参数错误
- `404`: 任务不存在
- `500`: 服务器内部错误

## 使用示例

### 完整的图像生成流程

1. **创建任务**:
```bash
curl -X POST http://localhost:8080/api/v1/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪，坐在花园里",
    "user_id": "user123",
    "size": "1024x1024",
    "quality": "hd"
  }'
```

2. **轮询查询结果**:
```bash
curl http://localhost:8080/api/v1/ai/image/result/67890abcdef
```

3. **获取用户任务列表**:
```bash
curl "http://localhost:8080/api/v1/ai/image/tasks?user_id=user123&limit=10"
```

## 最佳实践

1. **轮询间隔**: 建议每3-5秒查询一次任务状态
2. **超时处理**: 如果任务超过5分钟仍未完成，可能需要重新提交
3. **错误重试**: 对于网络错误，建议实现指数退避重试机制
4. **资源清理**: 及时删除不需要的任务记录

## 注意事项

- 图像生成通常需要30秒到2分钟时间
- 生成的图像URL有效期为24小时
- 每个用户同时最多可有10个pending状态的任务
- prompt文本建议控制在500字符以内 