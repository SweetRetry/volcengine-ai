# 🚀 Volcengine AI Server

基于火山方舟的企业级AI服务平台，采用现代化微服务架构，支持多AI服务商、异步任务处理和高并发场景。

## ✨ 核心特性

### 🔥 火山方舟集成
- **原生支持** 火山引擎豆包系列模型
- **图像生成** 基于 `doubao-seedream-3.0-t2i` 模型
- **文本生成** 支持 `doubao-pro-4k` 模型  
- **视频生成** 集成 `doubao-video-pro` 模型
- **高性能** 直接调用火山方舟API，无中间层损耗

### ⚡ Redis 异步队列系统
- **高并发处理** 基于 [Asynq](https://github.com/hibiken/asynq) 的分布式任务队列
- **优先级队列** 支持 critical、default、low 三级优先级
- **任务重试** 自动重试机制，确保任务可靠执行
- **实时监控** 队列状态实时监控和管理
- **水平扩展** 支持多worker节点分布式处理

### 🏗️ 服务商注册者模式
- **插件化架构** 支持多AI服务商动态注册
- **统一接口** 标准化的AI服务提供商接口
- **热插拔** 运行时动态添加/移除服务商
- **负载均衡** 智能路由到最优服务商
- **容错机制** 服务商故障自动切换

### 🔄 同步/异步双模式
- **同步模式** 适用于实时交互场景
- **异步模式** 适用于批量处理和长时间任务
- **灵活切换** 根据业务需求自由选择处理模式
- **状态追踪** 异步任务全生命周期状态管理

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
                    │   Service Registry       │
                    │  (Provider Pattern)      │
                    └─────────────┬─────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
┌───────▼────────┐    ┌───────────▼──────────┐    ┌────────▼────────┐
│ Volcengine AI  │    │    Redis Queue       │    │   MongoDB       │
│   Provider     │    │   (Asynq Worker)     │    │   Database      │
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
- **多模型支持** - 图像、文本、视频生成

### 队列系统
- **Asynq** - 分布式任务队列
- **Redis Streams** - 消息流处理
- **Worker Pool** - 并发任务处理

### 监控运维
- **Logrus** - 结构化日志
- **Prometheus** - 指标监控（规划中）
- **Docker** - 容器化部署

## 🚀 快速开始

### 环境要求
- Go 1.21+
- MongoDB 4.4+
- Redis 6.0+
- 火山方舟API密钥 (ARK_API_KEY)

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
# MONGO_URL=mongodb://localhost:27017/xxx
# REDIS_URL=redis://localhost:6379
# ARK_API_KEY=your_ark_api_key_here
# AI_TIMEOUT=30s
```

4. **启动服务**
```bash
# 启动API服务器
go run cmd/server/main.go

# 启动队列工作器
go run cmd/worker/main.go
```

### Docker 部署

```bash
# 构建镜像
docker build -t volcengine-ai-server .

# 使用 Docker Compose 启动完整环境
docker-compose up -d
```

## 📖 API 文档

### 图像生成

#### 异步生成（推荐）
```bash
# 创建图像生成任务
curl -X POST http://localhost:8080/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪在花园里玩耍",
    "user_id": "user123",
    "provider": "volcengine",
    "model": "doubao-seedream-3.0-t2i",
    "size": "1024x1024"
  }'

# 查询任务状态
curl http://localhost:8080/ai/image/result/{task_id}
```

### 任务管理

```bash
# 获取用户任务列表
curl "http://localhost:8080/ai/image/tasks?user_id=user123&limit=10"

# 删除任务
curl -X DELETE http://localhost:8080/ai/image/task/{task_id}
```

## 🔧 配置说明

### 服务商配置
```go
// 注册火山引擎服务商
arkAPIKey := os.Getenv("ARK_API_KEY")
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

### 数据库配置
```go
// MongoDB连接
mongoURL := os.Getenv("MONGO_URL")
db, err := database.NewMongoDB(mongoURL)
if err != nil {
    log.Fatal("数据库连接失败:", err)
}
```

## 📊 性能特性

### 并发处理能力
- **队列并发度**: 10个worker（可配置）
- **API并发**: 支持数千并发请求
- **任务吞吐**: 每秒处理100+图像生成任务

### 可扩展性
- **水平扩展**: 支持多实例部署
- **队列分片**: Redis集群支持
- **数据库分片**: MongoDB分片集群

### 可靠性保障
- **任务重试**: 自动重试失败任务
- **数据持久化**: MongoDB + Redis持久化
- **服务降级**: 服务商故障自动切换

## 🔍 监控运维

### 队列监控
```bash
# 查看队列状态
./scripts/check_queue_status.sh

# 清理队列
./scripts/clear_redis_queue.sh
```

### 性能测试
```bash
# 全链路测试
./scripts/test_volcengine_full.sh

# 压力测试
./scripts/load_test.sh
```

### 日志查看
```bash
# 实时日志
tail -f logs/app.log

# 错误日志
grep "ERROR" logs/app.log
```

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
- [Asynq](https://github.com/hibiken/asynq) - 优秀的Go任务队列库
- [Gin](https://github.com/gin-gonic/gin) - 高性能Web框架
- [MongoDB](https://www.mongodb.com/) - 灵活的文档数据库

## 📞 联系我们

- 项目主页: [GitHub Repository](https://github.com/your-org/volcengine-go-server)
- 问题反馈: [Issues](https://github.com/your-org/volcengine-go-server/issues)
- 邮箱: your-email@example.com

---

⭐ 如果这个项目对你有帮助，请给我们一个星标！ 