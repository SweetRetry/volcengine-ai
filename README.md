# 即梦AI服务器 (Jimeng AI Server)

基于Go语言开发的AI图像生成服务器，集成了火山引擎即梦AI图像生成服务。**采用纯异步模式，通过任务ID管理，确保用户出图稳定性。**

## 🚀 功能特性

- ✅ **异步任务处理** - 纯异步模式，保证出图稳定性
- ✅ **火山引擎即梦AI集成** - 支持多种图像生成模型
- ✅ **任务ID管理机制** - 唯一任务标识，支持状态追踪
- ✅ **用户管理系统** - 完整的用户CRUD操作
- ✅ **Redis队列支持** - 高性能异步任务队列
- ✅ **MongoDB数据存储** - 可靠的数据持久化
- ✅ **RESTful API设计** - 标准化的API接口
- ✅ **参数校验和错误处理** - 完善的输入验证和异常处理
- ✅ **轮询查询机制** - 支持任务状态实时查询
- ✅ **热重载开发** - 使用Air实现开发时热重载

## 📋 环境要求

- **Go**: 1.24+
- **MongoDB**: 4.4+
- **Redis**: 6.0+
- **火山引擎即梦AI**: API Key

## 🛠️ 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd jimeng-go-server
```

### 2. 安装依赖

```bash
make install
# 或者
go mod tidy
```

### 3. 配置环境变量

复制示例配置文件：
```bash
cp env.example .env
```

编辑 `.env` 文件，配置你的API Key：
```bash
# 火山引擎即梦AI配置
VOLCENGINE_ACCESS_KEY=你的ACCESS_KEY
VOLCENGINE_SECRET_KEY=你的SECRET_KEY

# 数据库配置
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=jimeng_ai

# Redis配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# 服务器配置
SERVER_PORT=8080
SERVER_HOST=localhost
```

### 4. 启动服务

**开发模式（推荐）：**
```bash
make dev
# 或者
air
```

**生产模式：**
```bash
make run
# 或者
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## 🗂️ 队列管理

项目使用 Redis 作为任务队列，提供了便捷的队列管理命令：

### 查看队列状态
```bash
make redis-queue-status
# 或者
./scripts/clear_redis_queue.sh --show-only
```

### 清理队列数据
```bash
# 交互式清理（需要确认）
make redis-queue-clear

# 强制清理（无需确认）
make redis-queue-clear-force
```

### 队列数据说明

#### 🏗️ 架构相关
- **`asynq:servers`**: 活跃的服务器实例列表，用于服务发现和健康检查
- **`asynq:workers`**: 活跃的工作器列表，用于负载均衡和监控
- **`asynq:queues`**: 所有已知的队列名称集合
- **`asynq:servers:{server_id}`**: 特定服务器实例的详细配置信息

#### 📋 队列任务
- **`asynq:{queue}:pending`**: 等待处理的任务队列
- **`asynq:{queue}:active`**: 正在处理的任务
- **`asynq:{queue}:retry`**: 失败后等待重试的任务
- **`asynq:{queue}:archived`**: 重试次数耗尽或跳过重试的失败任务
- **`asynq:{queue}:completed`**: 已完成的任务（可选）

#### 📊 统计数据
- **`asynq:{queue}:processed`**: 已处理任务总数计数器
- **`asynq:{queue}:failed`**: 失败任务总数计数器
- **`asynq:{queue}:processed:{date}`**: 按日期统计的已处理任务数
- **`asynq:{queue}:failed:{date}`**: 按日期统计的失败任务数

#### 🔧 任务数据
- **`asynq:{queue}:t:{task_id}`**: 存储任务的详细数据和元信息

⚠️ **注意**: 清理队列数据会删除所有未完成的任务和监控信息，请谨慎操作！

## 📚 API接口文档

### 健康检查
```http
GET /health
```

### 用户管理
```http
POST   /api/v1/users          # 创建用户
GET    /api/v1/users/:id      # 获取用户信息
GET    /api/v1/users?email=   # 通过邮箱查询用户
PUT    /api/v1/users/:id      # 更新用户信息
DELETE /api/v1/users/:id      # 删除用户
```

### AI图像生成 (异步模式)
```http
POST /api/v1/ai/image/task              # 创建图像生成任务
GET  /api/v1/ai/image/result/:task_id   # 查询任务结果
```

### 任务管理
```http
POST   /api/v1/ai/tasks              # 创建异步任务
GET    /api/v1/tasks/:id             # 获取任务详情
GET    /api/v1/tasks/user/:user_id   # 获取用户任务列表
```

## 🔄 异步工作流程

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

**响应示例:**
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

## 💡 使用示例

### 创建用户
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "张三"
  }'
```

### 异步生成图像（完整流程）
```bash
#!/bin/bash

# 1. 创建任务
echo "创建图像生成任务..."
TASK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪在花园里玩耍",
    "user_id": "user_123",
    "model": "doubao-seedream-3.0-t2i",
    "size": "1024x1024"
  }')

TASK_ID=$(echo $TASK_RESPONSE | jq -r '.data.task_id')
echo "任务ID: $TASK_ID"

# 2. 轮询查询结果
echo "开始轮询查询结果..."
MAX_ATTEMPTS=30
ATTEMPT=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
  RESULT=$(curl -s -X GET "http://localhost:8080/api/v1/ai/image/result/$TASK_ID")
  STATUS=$(echo $RESULT | jq -r '.data.status')
  
  echo "第 $((ATTEMPT + 1)) 次查询，状态: $STATUS"
  
  if [ "$STATUS" = "completed" ]; then
    IMAGE_URL=$(echo $RESULT | jq -r '.data.result.image_url')
    echo "✅ 图像生成完成!"
    echo "🖼️  图像URL: $IMAGE_URL"
    break
  elif [ "$STATUS" = "failed" ]; then
    ERROR_MSG=$(echo $RESULT | jq -r '.data.message')
    echo "❌ 任务失败: $ERROR_MSG"
    break
  else
    echo "⏳ 任务处理中，等待3秒后重试..."
    sleep 3
  fi
  
  ATTEMPT=$((ATTEMPT + 1))
done

if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
  echo "⏰ 任务查询超时"
fi
```

## 📁 项目结构

```
jimeng-go-server/
├── cmd/
│   └── server/              # 主程序入口
│       └── main.go
├── internal/
│   ├── config/             # 配置管理
│   ├── database/           # 数据库连接
│   ├── handler/            # HTTP处理器
│   ├── middleware/         # 中间件
│   ├── queue/             # 队列管理
│   ├── router/            # 路由配置
│   └── service/           # 业务逻辑
├── docs/                  # 文档目录
│   ├── AI_PROVIDER_ARCHITECTURE.md
│   ├── volcengine_ai_api.md
│   └── 热重载使用说明.md
├── scripts/               # 脚本文件
├── bin/                   # 编译输出
├── .air.toml             # Air热重载配置
├── Makefile              # 构建脚本
├── go.mod                # Go模块文件
├── go.sum                # Go依赖锁定
├── .env.example           # 环境变量示例
└── README.md             # 项目说明
```

## 🎨 火山引擎即梦AI集成

本项目深度集成了火山引擎的即梦AI图像生成服务，**采用纯异步模式**，具备以下特性：

### 支持功能
- 🎨 **文本到图像生成** - 支持中英文提示词
- 🔧 **多种模型选择** - doubao-seedream-3.0-t2i等
- ⚙️ **灵活参数配置** - 尺寸、质量、风格等
- 📊 **任务状态管理** - pending/processing/completed/failed
- 🔄 **轮询查询机制** - 实时状态更新
- 🛡️ **完整错误处理** - 详细的错误信息和重试机制
- 📈 **出图稳定性保障** - 任务ID确保结果不丢失

### 异步模式优势

1. **🔒 稳定性**: 通过任务ID管理，避免网络中断导致的结果丢失
2. **🔍 可追踪**: 每个任务都有唯一ID，便于状态查询和问题排查
3. **👥 用户体验**: 支持轮询查询，用户可以实时了解任务进度
4. **🚀 系统健壮性**: 异步处理避免长时间阻塞，提高系统并发能力
5. **📊 可扩展性**: 支持队列机制，便于水平扩展

详细的API文档请查看：[火山引擎即梦AI异步接口文档](docs/volcengine_ai_api.md)

## 🔧 开发指南

### 常用命令

```bash
# 安装依赖
make install

# 开发模式运行（热重载）
make dev

# 构建应用
make build

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# 查看所有可用命令
make help
```

### 添加新的AI服务提供商

1. 在 `internal/service/` 目录下创建新的服务文件
2. 实现相应的接口方法
3. 在 `internal/handler/` 中添加HTTP处理器
4. 在 `internal/router/` 中注册路由
5. 更新配置文件和文档

参考：[AI服务提供商架构文档](docs/AI_PROVIDER_ARCHITECTURE.md)

### 测试

运行异步API测试脚本：
```bash
./test_task_association.sh
```

测试脚本包含：
- ✅ 健康检查
- ✅ 用户创建和管理
- ✅ 异步任务创建
- ✅ 结果轮询查询
- ✅ 参数校验
- ✅ 错误处理测试

## 🚀 部署指南

### 本地部署

1. **构建二进制文件：**
```bash
make build
```

2. **配置生产环境变量**
3. **启动服务：**
```bash
./jimeng-server
```

### Docker部署

```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run
```

### 生产环境部署

```bash
# 构建Linux版本
make build-linux

# 部署到服务器
make deploy
```

## 💡 最佳实践

### 客户端轮询策略

```javascript
/**
 * 轮询查询任务结果
 * @param {string} taskId - 任务ID
 * @param {number} maxAttempts - 最大尝试次数
 * @param {number} interval - 轮询间隔（毫秒）
 */
async function pollTaskResult(taskId, maxAttempts = 30, interval = 3000) {
    for (let i = 0; i < maxAttempts; i++) {
        try {
            const response = await fetch(`/api/v1/ai/image/result/${taskId}`);
            const result = await response.json();
            
            if (result.data.status === 'completed') {
                return result.data.result.image_url;
            } else if (result.data.status === 'failed') {
                throw new Error(`任务失败: ${result.data.message}`);
            }
            
            // 建议3-5秒轮询间隔
            await new Promise(resolve => setTimeout(resolve, interval));
        } catch (error) {
            console.error(`轮询第${i + 1}次失败:`, error);
            if (i === maxAttempts - 1) throw error;
        }
    }
    
    throw new Error('任务查询超时');
}
```

### 错误处理和重试机制

```javascript
/**
 * 带重试机制的图像生成
 * @param {string} prompt - 提示词
 * @param {string} userId - 用户ID
 * @param {number} maxRetries - 最大重试次数
 */
async function createImageWithRetry(prompt, userId, maxRetries = 3) {
    for (let i = 0; i < maxRetries; i++) {
        try {
            // 创建任务
            const taskResponse = await fetch('/api/v1/ai/image/task', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ prompt, user_id: userId })
            });
            
            const taskResult = await taskResponse.json();
            if (!taskResult.success) {
                throw new Error(taskResult.message);
            }
            
            // 轮询结果
            return await pollTaskResult(taskResult.data.task_id);
            
        } catch (error) {
            console.error(`第${i + 1}次尝试失败:`, error);
            
            if (i === maxRetries - 1) throw error;
            
            // 指数退避策略
            const delay = Math.pow(2, i) * 1000;
            await new Promise(resolve => setTimeout(resolve, delay));
        }
    }
}
```

## 📈 更新日志

### v2.1.0 (当前版本)
- 🔄 **配置优化**: 简化火山引擎API配置，移除冗余参数
- ✅ **架构改进**: 新增AI服务提供商架构支持
- ✅ **任务关联**: 实现外部任务ID关联功能
- ✅ **代码重构**: 优化代码结构，提升可维护性
- ✅ **文档完善**: 更新API文档和架构说明

### v2.0.0
- 🔄 **重大变更**: 改为纯异步模式
- ✅ **任务管理**: 新增任务ID管理机制
- ✅ **稳定性**: 优化用户出图稳定性
- ✅ **错误处理**: 完善错误处理和状态管理
- ✅ **测试脚本**: 新增异步API测试脚本
- ❌ **接口移除**: 移除同步图像生成接口

### v1.0.0
- ✅ **基础功能**: 基础同步图像生成功能
- ✅ **API集成**: 火山引擎API集成
- ✅ **用户系统**: 用户管理系统
- ✅ **队列支持**: 任务队列支持

## 🤝 贡献指南

我们欢迎所有形式的贡献！请遵循以下步骤：

1. **Fork** 项目到你的GitHub账户
2. **创建功能分支** (`git checkout -b feature/AmazingFeature`)
3. **提交更改** (`git commit -m 'feat: 添加某个很棒的功能'`)
4. **推送到分支** (`git push origin feature/AmazingFeature`)
5. **创建Pull Request**

### 提交信息规范

请使用以下格式的提交信息：
- `feat: 新功能`
- `fix: 修复bug`
- `docs: 文档更新`
- `style: 代码格式调整`
- `refactor: 代码重构`
- `test: 测试相关`
- `chore: 构建过程或辅助工具的变动`

## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。

## 📞 联系方式

- **Issues**: [GitHub Issues](https://github.com/your-repo/jimeng-go-server/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-repo/jimeng-go-server/discussions)
- **Email**: your-email@example.com

## 🙏 致谢

感谢以下开源项目和服务：

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web框架
- [MongoDB](https://www.mongodb.com/) - 数据库
- [Redis](https://redis.io/) - 缓存和队列
- [火山引擎即梦AI](https://www.volcengine.com/) - AI图像生成服务
- [Air](https://github.com/cosmtrek/air) - 热重载工具

---

**⭐ 如果这个项目对你有帮助，请给我们一个Star！** 