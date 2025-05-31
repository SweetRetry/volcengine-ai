# API 使用示例

## 即梦AI图生视频 (Image to Video)

### 基本用法

```bash
curl -X POST http://localhost:8080/api/v1/ai/video/task \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "provider": "volcengine",
    "model": "jimeng_vgfm_i2v_l20",
    "image_urls": [
      "https://example.com/image1.jpg"
    ],
    "aspect_ratio": "16:9"
  }'
```

### 带提示词的图生视频

```bash
curl -X POST http://localhost:8080/api/v1/ai/video/task \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "provider": "volcengine",
    "model": "jimeng_vgfm_i2v_l20",
    "image_urls": [
      "https://example.com/image1.jpg"
    ],
    "prompt": "让图片中的人物挥手",
    "aspect_ratio": "16:9",
    "seed": 12345
  }'
```

### 自动检测图片尺寸比例

如果不提供 `aspect_ratio` 参数，系统会自动检测第一张图片的尺寸并匹配最接近的标准比例：

```bash
curl -X POST http://localhost:8080/api/v1/ai/video/task \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "provider": "volcengine",
    "model": "jimeng_vgfm_i2v_l20",
    "image_urls": [
      "https://example.com/image1.jpg"
    ],
    "prompt": "让图片动起来"
  }'
```

### 响应示例

```json
{
  "success": true,
  "data": {
    "task_id": "60f7b1234567890abcdef123",
    "status": "pending",
    "provider": "volcengine",
    "model": "jimeng_vgfm_i2v_l20",
    "seed": 12345,
    "aspect_ratio": "16:9",
    "image_count": 1,
    "task_type": "image_to_video"
  },
  "message": "视频生成任务创建成功"
}
```

## 即梦AI文生视频 (Text to Video)

### 基本用法

```bash
curl -X POST http://localhost:8080/api/v1/ai/video/task \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "provider": "volcengine",
    "model": "jimeng_vgfm_t2v_l20",
    "prompt": "一只可爱的小猫在花园里玩耍",
    "aspect_ratio": "16:9",
    "seed": 12345
  }'
```

### 响应示例

```json
{
  "success": true,
  "data": {
    "task_id": "60f7b1234567890abcdef124",
    "status": "pending",
    "provider": "volcengine",
    "model": "jimeng_vgfm_t2v_l20",
    "seed": 12345,
    "aspect_ratio": "16:9",
    "task_type": "text_to_video"
  },
  "message": "视频生成任务创建成功"
}
```

## 参数说明

### 必填参数

- `user_id`: 用户ID
- `provider`: AI服务提供商，固定为 "volcengine"
- `model`: AI模型名称
  - `jimeng_vgfm_i2v_l20`: 图生视频模型
  - `jimeng_vgfm_t2v_l20`: 文生视频模型

### 图生视频特有参数

- `image_urls`: 图片链接数组（必填）
- `prompt`: 视频生成提示词（可选，150字符以内）

### 文生视频特有参数

- `prompt`: 视频生成提示词（必填，150字符以内）

### 通用可选参数

- `aspect_ratio`: 视频尺寸比例，支持的值：
  - `16:9` (默认)
  - `9:16`
  - `1:1`
  - `4:3`
  - `3:4`
  - `21:9`
  - `9:21`
- `seed`: 随机种子，范围 [-1, 2^64-1]，-1表示随机生成

## 查询任务结果

```bash
curl -X GET http://localhost:8080/api/v1/ai/tasks/{task_id}
```

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "task_id": "60f7b1234567890abcdef123",
    "type": "video",
    "status": "completed",
    "video_url": "https://example.com/generated_video.mp4",
    "seed": 12345,
    "aspect_ratio": "16:9",
    "created": "2023-07-20T10:30:00Z",
    "updated": "2023-07-20T10:35:00Z"
  },
  "message": "任务完成"
}
```

## 错误处理

### 参数验证错误

```json
{
  "error": "图生视频任务缺少image_urls参数",
  "message": "请提供至少一个图片链接"
}
```

### 不支持的比例

```json
{
  "error": "不支持的aspect_ratio: 2:1，支持的比例: 16:9, 4:3, 1:1, 3:4, 9:16, 21:9, 9:21"
}
```

### 提示词过长

```json
{
  "error": "prompt长度超过150字符限制，当前长度: 200"
}
```

## 注意事项

1. **图片格式支持**: 支持 JPEG、PNG、WebP 格式
2. **图片尺寸检测**: 系统会自动检测第一张图片尺寸并匹配最接近的标准比例
3. **多图片处理**: 虽然支持传入多个图片URL，但尺寸检测只基于第一张图片
4. **任务处理时间**: 视频生成通常需要1-3分钟，请耐心等待
5. **并发限制**: 系统支持多个任务并发处理
6. **结果保存**: 生成的视频会保存在云存储中，链接长期有效 