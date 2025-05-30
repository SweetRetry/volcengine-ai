# 🚀 Volcengine AI Server

基于火山方舟的企业级AI服务平台，采用现代化微服务架构，支持多AI服务商、异步任务处理和高并发场景。

## ✨ 核心特性

### 🔥 火山方舟多模型集成

- **豆包图像生成** 基于 `doubao-seedream-3.0-t2i-250415` 模型，支持高质量图像生成
- **即梦AI图像生成** 集成 `jimeng_high_aes_general_v21_L` 模型，专业级艺术创作
- **智能模型路由** 根据不同模型自动选择最优处理策略
- **多格式支持** 支持URL和Base64两种图片返回格式
- **尺寸优化** 针对即梦AI官方建议的最佳尺寸配置进行优化

### 🎯 即梦AI专业特性

- **官方推荐尺寸** 支持1:1、4:3、3:4、3:2、2:3、16:9、9:16等最佳比例
- **智能提示词扩写** 短提示词自动开启LLM扩写功能
- **AIGC超分技术** 自动开启超分辨率增强
- **双格式输出** 灵活支持图片URL和Base64数据返回
- **防御性编程** 完善的错误处理和日志记录

### ⚡ Redis 异步队列系统

- **高并发处理** 基于 [Asynq](https://github.com/hibiken/asynq) 的分布式任务队列
- **优先级队列** 支持 critical、default、low 三级优先级
- **任务重试** 自动重试机制，确保任务可靠执行
- **实时监控** 队列状态实时监控和管理
- **水平扩展** 支持多worker节点分布式处理

### 📝 智能日志管理系统

- **双输出模式** 同时输出到控制台和本地文件，支持实时查看和持久化存储
- **自动日志轮转** 按日期自动创建新日志文件，每天午夜自动切换
- **智能清理机制** 自动清理过期日志文件，可配置保留天数（默认7天）
- **结构化日志** 采用JSON格式，包含时间戳、级别、消息和结构化字段
- **灵活配置** 支持通过环境变量配置日志级别和保留策略
- **手动管理** 支持强制轮转和清理操作，便于运维管理

### 🏗️ 服务商注册者模式

- **插件化架构** 支持多AI服务商动态注册
- **统一接口** 标准化的AI服务提供商接口
- **模型路由** 智能根据模型类型选择处理策略
- **热插拔** 运行时动态添加/移除服务商
- **容错机制** 服务商故障自动切换

### 🔄 同步/异步双模式

- **异步任务** 适用于批量处理和长时间任务（推荐）
- **状态追踪** 异步任务全生命周期状态管理
- **分页查询** 支持用户任务列表分页查询
- **任务管理** 支持任务删除和状态更新

## 🏛️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Mobile App    │    │   Third Party   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      API Gateway         │
                    │    (Gin Framework)       │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │   AI Task Factory        │
                    │  (Model Router)          │
                    └─────────────┬─────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
┌───────▼────────┐    ┌───────────▼──────────┐    ┌────────▼────────┐
│ Volcengine AI  │    │    Redis Queue       │    │   MongoDB       │
│   Provider     │    │   (Asynq Worker)     │    │   Database      │
│ ├─豆包模型     │    │                      │    │                 │
│ └─即梦AI       │    │                      │    │                 │
└────────────────┘    └──────────────────────┘    └─────────────────┘
```

## 🛠️ 技术栈

### 后端核心

- **Go 1.21+** - 高性能后端语言
- **Gin** - 轻量级Web框架
- **MongoDB** - 文档数据库
- **Redis** - 缓存和队列存储

### AI服务集成

- **火山方舟SDK** - 官方Go SDK
- **火山引擎Visual SDK** - 即梦AI专用SDK
- **多模型支持** - 豆包、即梦AI图像生成

### 队列系统

- **Asynq** - 分布式任务队列
- **Redis Backend** - 队列数据存储和持久化
- **Worker Pool** - 并发任务处理

### 监控运维

- **Logrus** - 结构化日志
- **日志管理器** - 自动轮转和清理
- **参数验证** - Gin binding + validator
- **错误处理** - 统一错误响应机制

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MongoDB 4.4+
- Redis 6.0+
- 火山方舟API密钥 (ARK_API_KEY)
- 火山引擎Access Key (可选，用于即梦AI)

### 安装部署

1. **克隆项目**

```bash
git clone https://github.com/your-org/volcengine-go-server.git
cd volcengine-go-server
```

2. **安装依赖**

```bash
go mod download
```

3. **配置环境变量**

```bash
cp env.example .env
# 编辑 .env 文件，配置以下关键参数：
# PORT=8080
# ENVIRONMENT=development
# MONGO_URL=mongodb://localhost:27017/volcengine_db
# REDIS_URL=redis://localhost:6379
# ARK_API_KEY=your_ark_api_key_here
# VOLCENGINE_ACCESS_KEY=your_access_key (可选)
# VOLCENGINE_SECRET_KEY=your_secret_key (可选)
# AI_TIMEOUT=30s
# LOG_LEVEL=info
# LOG_KEEP_DAYS=7
```

4. **启动服务**

```bash
# 启动API服务器
make run
# 或者
go run cmd/server/main.go

# 启动队列工作器（任务处理中心）
make run-worker
# 或者
go run cmd/worker/main.go

# 构建所有服务
make build-all

# 查看所有可用命令
make help
```

### 开发模式（热重载）

安装Air工具（如果尚未安装）：
```bash
go install github.com/cosmtrek/air@latest
```

启动开发模式：
```bash
# 开发模式运行API服务器（热重载）
make dev

# 开发模式运行Worker服务（热重载）
make dev-worker

# 查看如何同时运行两个服务
make dev-all
```

在不同终端窗口中同时运行两个服务：
```bash
# 终端1 - API服务器
make dev

# 终端2 - Worker服务  
make dev-worker
```

### 快速开发启动

使用提供的开发启动脚本：
```bash
# 运行交互式开发启动脚本
./scripts/dev-start.sh
```

该脚本会：
- 自动检查并安装Air工具
- 检查并创建.env配置文件
- 提供交互式菜单选择启动模式

## 📖 API 文档

### 图像生成

#### 豆包模型图像生成

```bash
curl -X POST http://localhost:8080/api/v1/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪在花园里玩耍",
    "user_id": "user123",
    "provider": "volcengine",
    "model": "doubao-seedream-3.0-t2i-250415",
    "size": "1024x1024"
  }'
```

#### 即梦AI图像生成

```bash
curl -X POST http://localhost:8080/api/v1/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "过曝，强对比，夜晚，雪地里，巨大的黄色浴缸，小狗泡澡带墨镜",
    "user_id": "user123",
    "provider": "volcengine",
    "model": "jimeng_high_aes_general_v21_L",
    "size": "16:9"
  }'
```

#### 查询任务状态

```bash
curl http://localhost:8080/api/v1/ai/image/result/{task_id}
```

### 支持的图像尺寸

#### 豆包模型支持尺寸

- `1024x1024` (1:1) - 默认
- `864x1152` (3:4)
- `1152x864` (4:3)
- `1280x720` (16:9)
- `720x1280` (9:16)
- `832x1248` (2:3)
- `1248x832` (3:2)
- `1512x648` (21:9)

#### 即梦AI推荐尺寸（官方优化）

- `512x512` (1:1) - 最佳效果
- `512x384` (4:3)
- `384x512` (3:4)
- `512x341` (3:2)
- `341x512` (2:3)
- `512x288` (16:9)
- `288x512` (9:16)

### 任务管理

```bash
# 获取用户任务列表
curl "http://localhost:8080/api/v1/ai/image/tasks?user_id=user123&limit=10&offset=0"

# 删除任务
curl -X DELETE http://localhost:8080/api/v1/ai/image/task/{task_id}
```

### 用户管理

```bash
# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "张三"
  }'

# 获取用户信息
curl http://localhost:8080/api/v1/users/{user_id}
```

## 📝 日志系统

### 日志文件结构

```
logs/
├── app-2024-01-15.log    # 今天的日志
├── app-2024-01-14.log    # 昨天的日志
├── app-2024-01-13.log    # 前天的日志
└── ...                   # 更早的日志（根据保留策略）
```

### 日志格式

采用JSON格式，包含以下字段：

```json
{
  "time": "2024-01-15 10:30:45",
  "level": "info",
  "msg": "用户登录",
  "user_id": "12345",
  "ip": "192.168.1.1"
}
```

### 日志级别

- `debug`: 调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息

### 日志管理命令

```bash
# 查看今天的日志
make show-logs

# 查看实时日志
tail -f logs/app-$(date +%Y-%m-%d).log

# 搜索错误日志
grep "ERROR" logs/app-*.log

# 清理所有日志文件
make clean-logs

# 查看日志文件大小
du -h logs/
```

### 自动化功能

- **自动轮转**: 每天午夜自动创建新的日志文件
- **自动清理**: 自动删除超过保留期限的日志文件（默认7天）
- **双输出**: 同时输出到控制台和文件，便于开发和生产环境使用

## 🔧 配置说明

### 模型配置常量

```go
// 火山引擎豆包模型
VolcengineImageModel = "doubao-seedream-3.0-t2i-250415"

// 火山引擎即梦AI模型
VolcengineJimengImageModel = "jimeng_high_aes_general_v21_L"

// 即梦AI推荐尺寸
JimengImageSize1x1  = "512x512"  // 1:1 比例
JimengImageSize16x9 = "512x288"  // 16:9 比例
```

### 服务商注册

```go
// 注册火山引擎服务商
registry := queue.NewServiceRegistry()
volcengineProvider := service.NewVolcengineAIProvider(
    volcengineService,
    imageTaskService,
)
registry.RegisterProvider(volcengineProvider)
```

### 队列配置

```go
// Redis队列配置
redisURL := os.Getenv("REDIS_URL")
queue := queue.NewRedisQueue(
    redisURL,
    imageTaskService,
    serviceRegistry,
)

// 启动工作器
go queue.StartWorker(ctx)
```

### 日志系统配置

```go
// 日志管理器配置
logManager := logger.NewLogManager()
logManager.SetKeepDays(7)                    // 保留7天
logManager.SetRotateInterval(24 * time.Hour) // 24小时轮转一次
logManager.SetCleanInterval(24 * time.Hour)  // 24小时清理一次

// 启动日志管理器
ctx := context.Background()
go logManager.Start(ctx)

// 手动操作
logManager.ForceRotate() // 强制轮转
logManager.ForceClean()  // 强制清理
```

### 环境变量配置

```bash
# 日志级别 (debug, info, warn, error)
LOG_LEVEL=info

# 日志保留天数
LOG_KEEP_DAYS=7
```

## 📊 性能特性

### 模型性能对比

| 模型     | 生成时间 | 图像质量 | 适用场景           |
| -------- | -------- | -------- | ------------------ |
| 豆包模型 | 3-5秒    | 高质量   | 通用图像生成       |
| 即梦AI   | 6-10秒   | 专业级   | 艺术创作、专业设计 |

### 并发处理能力

- **队列并发度**: 10个worker（可配置）
- **API并发**: 支持数千并发请求
- **任务吞吐**: 每秒处理100+图像生成任务

### 可扩展性

- **水平扩展**: 支持多实例部署
- **队列分片**: Redis集群支持
- **数据库分片**: MongoDB分片集群

## 🔍 监控运维

### 日志管理

```bash
# 查看今天的日志
make show-logs

# 查看实时日志
tail -f logs/app-$(date +%Y-%m-%d).log

# 清理日志文件
make clean-logs

# 搜索错误日志
grep "ERROR" logs/app-*.log

# 查看日志文件列表
ls -la logs/
```

### 队列监控

```bash
# 查看队列状态
make redis-queue-status

# 清理队列
make redis-queue-clear

# 强制清理队列（无需确认）
make redis-queue-clear-force
```

## 🆕 最新更新

### v1.3.0 - 智能日志管理系统

- ✅ **日志管理系统**: 完整的日志轮转和清理功能
- ✅ **双输出模式**: 同时输出到控制台和文件
- ✅ **自动轮转**: 按日期自动创建新日志文件
- ✅ **智能清理**: 自动清理过期日志，可配置保留天数
- ✅ **Makefile优化**: 新增构建和运行任务处理中心的命令
- ✅ **运维工具**: 增强日志查看和管理命令

### v1.2.0 - 即梦AI集成与多模型支持

- ✅ **即梦AI集成**: 完整支持即梦AI图像生成模型
- ✅ **多模型路由**: 智能根据模型选择处理策略
- ✅ **尺寸优化**: 基于官方建议的最佳尺寸配置
- ✅ **双格式支持**: URL和Base64两种返回格式
- ✅ **参数验证优化**: 简化验证逻辑，移除冗余标签
- ✅ **常量重构**: 统一模型和尺寸常量管理

### v1.1.0 - 架构重构与性能优化

- ✅ **服务商模式**: 插件化AI服务提供商架构
- ✅ **任务工厂**: 统一的AI任务创建和管理
- ✅ **队列系统**: 基于Redis的异步任务处理
- ✅ **数据库优化**: MongoDB索引和查询优化
- ✅ **错误处理**: 统一的错误响应机制

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: 添加某个特性'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

### 提交规范

- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

- [火山方舟](https://www.volcengine.com/product/ark) - 提供强大的AI能力
- [火山引擎即梦AI](https://www.volcengine.com/product/jimeng) - 专业级图像生成
- [Asynq](https://github.com/hibiken/asynq) - 优秀的Go任务队列库
- [Gin](https://github.com/gin-gonic/gin) - 高性能Web框架
- [MongoDB](https://www.mongodb.com/) - 灵活的文档数据库

⭐ 如果这个项目对你有帮助，请给我们一个星标！

## 📚 相关文档

- [日志系统说明](docs/日志系统说明.md) - 详细的日志管理系统使用指南

## 📞 联系我们

- 项目主页: [GitHub Repository](https://github.com/your-org/volcengine-go-server)
- 问题反馈: [Issues](https://github.com/your-org/volcengine-go-server/issues)
- 邮箱: <zimin.zhang2000@gmail.com>

---
