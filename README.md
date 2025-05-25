# 火山引擎即梦AI服务器

基于Go语言开发的AI服务器，集成了火山引擎即梦AI图像生成服务。**采用纯异步模式，通过任务ID管理，确保用户出图稳定性。**

## 功能特性

- ✅ 用户管理系统
- ✅ **异步任务处理** (纯异步模式)
- ✅ 火山引擎即梦AI图像生成 (异步)
- ✅ 任务ID管理机制
- ✅ Redis队列支持
- ✅ MongoDB数据存储
- ✅ 完整的API文档
- ✅ 参数校验和错误处理
- ✅ 轮询查询机制

## 快速开始

### 1. 环境要求

- Go 1.24+
- MongoDB
- Redis
- 火山引擎即梦AI API Key

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境变量

复制示例配置文件：
```bash
cp config.env.example .env
```

编辑 `.env` 文件，配置你的API Key：
```bash
# 火山引擎即梦AI配置
VOLCENGINE_API_KEY=你的API_KEY
VOLCENGINE_ENDPOINT=https://ark.cn-beijing.volces.com
VOLCENGINE_REGION=cn-beijing
```

### 4. 启动服务

开发模式（热重载）：
```bash
air
```

或直接运行：
```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## API接口

### 健康检查
```bash
GET /health
```

### 用户管理
```bash
POST   /api/v1/users          # 创建用户
GET    /api/v1/users/:id      # 获取用户
GET    /api/v1/users?email=   # 通过邮箱查询用户
PUT    /api/v1/users/:id      # 更新用户
DELETE /api/v1/users/:id      # 删除用户
```

### AI图像生成 (异步模式)
```bash
POST /api/v1/ai/image/task              # 创建图像生成任务
GET  /api/v1/ai/image/result/:task_id   # 查询任务结果
```

### 任务管理
```bash
POST   /api/v1/ai/tasks       # 创建异步任务
GET    /api/v1/tasks/:id      # 获取任务详情
GET    /api/v1/tasks/user/:user_id  # 获取用户任务列表
```

## 异步工作流程

### 1. 创建图像生成任务
```bash
curl -X POST http://localhost:8080/api/v1/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪在花园里玩耍",
    "user_id": "user_123",
    "model": "doubao-seedream-3.0-t2i",
    "size": "1024x1024",
    "quality": "standard"
  }'
```

**响应:**
```json
{
  "success": true,
  "data": {
    "task_id": "volcengine_img_1703123456789",
    "status": "pending",
    "message": "任务已创建，正在处理中"
  }
}
```

### 2. 查询任务结果
```bash
curl -X GET http://localhost:8080/api/v1/ai/image/result/volcengine_img_1703123456789
```

**处理中响应:**
```json
{
  "success": true,
  "data": {
    "task_id": "volcengine_img_1703123456789",
    "status": "processing",
    "message": "任务处理中，请稍后查询"
  }
}
```

**完成响应:**
```json
{
  "success": true,
  "data": {
    "task_id": "volcengine_img_1703123456789",
    "status": "completed",
    "result": {
      "image_url": "https://example.com/generated-image.jpg"
    }
  }
}
```

## 使用示例

### 创建用户
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "张三"
  }'
```

### 异步生成图像
```bash
# 1. 创建任务
TASK_ID=$(curl -s -X POST http://localhost:8080/api/v1/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪在花园里玩耍",
    "user_id": "user_123"
  }' | jq -r '.data.task_id')

echo "任务ID: $TASK_ID"

# 2. 轮询查询结果
while true; do
  RESULT=$(curl -s -X GET "http://localhost:8080/api/v1/ai/image/result/$TASK_ID")
  STATUS=$(echo $RESULT | jq -r '.data.status')
  
  if [ "$STATUS" = "completed" ]; then
    IMAGE_URL=$(echo $RESULT | jq -r '.data.result.image_url')
    echo "图像生成完成: $IMAGE_URL"
    break
  elif [ "$STATUS" = "failed" ]; then
    echo "任务失败"
    break
  else
    echo "任务处理中，等待3秒..."
    sleep 3
  fi
done
```

## 项目结构

```
volcengine-ai-server/
├── cmd/server/           # 主程序入口
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── handler/         # HTTP处理器
│   ├── middleware/      # 中间件
│   ├── queue/          # 队列管理
│   ├── router/         # 路由配置
│   └── service/        # 业务逻辑
├── docs/               # API文档
├── test_volcengine_async_api.sh  # 异步API测试脚本
└── README.md
```

## 火山引擎即梦AI集成

本项目集成了火山引擎的即梦AI图像生成服务，**采用纯异步模式**，支持：

- 🎨 文本到图像生成 (异步)
- 🔧 多种模型选择
- ⚙️ 灵活的参数配置
- 📊 任务状态管理
- 🔄 轮询查询机制
- 🛡️ 完整的错误处理
- 📈 出图稳定性保障

### 异步模式优势

1. **稳定性**: 通过任务ID管理，避免网络中断导致的结果丢失
2. **可追踪**: 每个任务都有唯一ID，便于状态查询和问题排查
3. **用户体验**: 支持轮询查询，用户可以实时了解任务进度
4. **系统健壮性**: 异步处理避免长时间阻塞，提高系统并发能力

详细的API文档请查看：[火山引擎即梦AI异步接口文档](docs/volcengine_ai_api.md)

## 开发指南

### 添加新的AI服务

1. 在 `internal/service/` 目录下创建新的服务文件
2. 实现相应的接口方法
3. 在 `internal/handler/` 中添加HTTP处理器
4. 在 `internal/router/` 中注册路由
5. 更新配置文件和文档

### 测试

运行异步API测试脚本：
```bash
./test_volcengine_async_api.sh
```

测试脚本包含：
- 健康检查
- 用户创建
- 异步任务创建
- 结果轮询查询
- 参数校验
- 错误处理

### 部署

1. 构建二进制文件：
```bash
go build -o volcengine-ai-server cmd/server/main.go
```

2. 配置生产环境变量
3. 启动服务：
```bash
./volcengine-ai-server
```

## 最佳实践

### 客户端轮询策略

```javascript
async function pollTaskResult(taskId, maxAttempts = 30) {
    for (let i = 0; i < maxAttempts; i++) {
        const response = await fetch(`/api/v1/ai/image/result/${taskId}`);
        const result = await response.json();
        
        if (result.data.status === 'completed') {
            return result.data.result.image_url;
        } else if (result.data.status === 'failed') {
            throw new Error('任务失败');
        }
        
        // 建议3-5秒轮询间隔
        await new Promise(resolve => setTimeout(resolve, 3000));
    }
    
    throw new Error('任务超时');
}
```

### 错误处理

```javascript
async function createImageWithRetry(prompt, userId, maxRetries = 3) {
    for (let i = 0; i < maxRetries; i++) {
        try {
            const taskId = await createImageTask(prompt, userId);
            return await pollTaskResult(taskId);
        } catch (error) {
            if (i === maxRetries - 1) throw error;
            
            const delay = Math.pow(2, i) * 1000; // 指数退避
            await new Promise(resolve => setTimeout(resolve, delay));
        }
    }
}
```

## 更新日志

### v2.0.0 (当前版本)
- 🔄 **重大变更**: 改为纯异步模式
- ✅ 新增任务ID管理机制
- ✅ 优化用户出图稳定性
- ✅ 完善错误处理和状态管理
- ✅ 新增异步API测试脚本
- ❌ 移除同步图像生成接口

### v1.0.0
- ✅ 基础同步图像生成功能
- ✅ 火山引擎API集成
- ✅ 用户管理系统
- ✅ 任务队列支持

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交 Issue 或联系开发团队。 