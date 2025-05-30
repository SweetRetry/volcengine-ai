# 🚀 Volcengine AI Server

基于火山方舟的企业级AI服务平台，采用现代化分层架构设计，支持多AI服务商、异步任务处理和高并发场景。

## ✨ 核心特性

### 🏗️ 现代化分层架构

- **Core层** - 核心基础设施，包含任务队列系统和AI任务分发器接口
- **Provider层** - 任务分发层，根据模型参数智能路由到具体Service
- **Service层** - 业务实现层，包含真实的AI API调用逻辑
- **Repository层** - 数据访问层，统一的数据库操作接口
- **职责分离** - 每层职责明确，易于维护和扩展

### 🔥 火山方舟多模型集成

- **豆包图像生成** 基于 `doubao-seedream-3-0-t2i-250415` 模型，支持高质量图像生成
- **即梦AI图像生成** 集成 `jimeng_high_aes_general_v21_L` 模型，专业级艺术创作
- **即梦AI视频生成** 基于 `jimeng_vgfm_t2v_l20` 模型，支持文本到视频生成
- **智能模型路由** Provider层根据不同模型自动选择最优处理策略
- **多格式支持** 支持URL和Base64两种图片返回格式
- **尺寸优化** 针对即梦AI官方建议的最佳尺寸配置进行优化

### ⚡ 核心基础设施 (Core)

- **任务队列系统** 基于 [Asynq](https://github.com/hibiken/asynq) 的分布式任务队列
- **AI任务分发器** 统一的任务分发接口，支持插件化扩展
- **服务注册器** 动态注册和管理AI服务提供商
- **优先级队列** 支持 critical、default、low 三级优先级
- **任务重试** 自动重试机制，确保任务可靠执行
- **实时监控** 队列状态实时监控和管理

### 🎯 Provider-Service 分层模式

- **Provider层职责** - 任务分发和路由，根据模型参数决定调用哪个Service方法
- **Service层职责** - 具体的AI API调用和业务逻辑实现
- **接口解耦** - Provider只依赖Service接口，完全解耦
- **热插拔** - 运行时动态添加/移除服务商
- **容错机制** - 服务商故障自动切换

### 📝 智能日志管理系统

- **双输出模式** 同时输出到控制台和本地文件，支持实时查看和持久化存储
- **自动日志轮转** 按日期自动创建新日志文件，每天午夜自动切换
- **智能清理机制** 自动清理过期日志文件，可配置保留天数（默认7天）
- **结构化日志** 采用JSON格式，包含时间戳、级别、消息和结构化字段
- **灵活配置** 支持通过环境变量配置日志级别和保留策略

### 🔄 统一任务管理

- **统一Task模型** 一个模型处理所有类型任务（图像、视频、文本）
- **统一TaskService** 一个服务处理所有任务操作
- **极简API设计** 总共只有6个API接口（3个创建 + 3个统一）
- **线性扩展** API数量随任务类型线性增长（N+3），而非传统的平方增长（N×4）
- **状态追踪** 异步任务全生命周期状态管理
- **分页查询** 支持用户任务列表分页查询

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
                    │     AI Handler           │
                    │   (Unified Router)       │
                    └─────────────┬─────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
┌───────▼────────┐    ┌───────────▼──────────┐    ┌────────▼────────┐
│ Core Layer     │    │   Provider Layer     │    │  Service Layer  │
│ ├─TaskQueue    │    │ ├─VolcengineProvider │    │ ├─VolcengineService│
│ ├─Dispatcher   │    │ └─OpenAIProvider     │    │ └─OpenAIService  │
│ └─Registry     │    │                      │    │                 │
└────────────────┘    └──────────────────────┘    └─────────────────┘
        │                         │                         │
        └─────────────────────────┼─────────────────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
┌───────▼────────┐              ┌─▼─────────────────┐    ┌────────▼────────┐
│    Redis       │              │     MongoDB       │    │   Repository    │
│   (Queue)      │              │   (Database)      │    │     Layer       │
└────────────────┘              └───────────────────┘    └─────────────────┘
```

### 架构层次说明

#### Core Layer (核心基础设施层)

- **TaskQueue** - 任务队列系统，处理异步任务调度
- **AITaskDispatcher** - AI任务分发器接口定义
- **ServiceRegistry** - 服务注册器，管理Provider实例

#### Provider Layer (任务分发层)

- **VolcengineProvider** - 火山引擎任务分发器
- **OpenAIProvider** - OpenAI任务分发器
- **职责** - 根据模型参数决定调用哪个Service方法

#### Service Layer (业务实现层)

- **VolcengineService** - 火山引擎API具体实现
- **OpenAIService** - OpenAI API具体实现
- **TaskService** - 统一任务管理服务
- **职责** - 真实的AI API调用和业务逻辑

#### Repository Layer (数据访问层)

- **MongoDB** - 任务数据持久化
- **统一接口** - 标准化数据访问操作

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

### 核心基础设施

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

### ⚠️ 重要说明

**model和provider字段为必填参数**：所有任务创建接口都要求明确指定`model`和`provider`字段，不再提供默认值。这样设计的目的是：
- 确保用户明确知道使用的是哪个AI模型和服务提供商
- 避免因默认值变更导致的意外结果
- 提高API的明确性和可预测性
- 简化代码逻辑，减少配置复杂度

### 🎯 统一API设计

本系统采用统一的API设计模式，所有AI任务都遵循相同的请求结构和响应格式：

#### 任务创建接口

```bash
# 图像生成任务
POST /api/v1/ai/image/task

# 视频生成任务  
POST /api/v1/ai/video/task

# 文本生成任务
POST /api/v1/ai/text/task
```

#### 统一请求结构

```json
{
  "prompt": "任务描述文本",
  "user_id": "用户ID",
  "provider": "服务提供商名称",
  "model": "具体模型名称",
  
  // 可选参数（根据任务类型）
  "size": "图像尺寸",
  "aspect_ratio": "视频比例", 
  "max_tokens": 1000,
  "temperature": 0.7,
  "seed": -1
}
```

#### 统一响应结构

```json
{
  "success": true,
  "data": {
    "task_id": "任务ID",
    "status": "pending",
    "provider": "服务提供商",
    "model": "模型名称"
  },
  "message": "任务创建成功"
}
```

### 🔄 统一任务管理

```bash
# 查询任务结果（支持所有任务类型）
GET /api/v1/ai/task/result/{task_id}

# 删除任务（支持所有任务类型）
DELETE /api/v1/ai/task/{task_id}

# 获取用户任务列表（支持类型过滤）
GET /api/v1/ai/tasks?user_id={user_id}&type={type}&limit={limit}&offset={offset}
```

#### 任务查询响应

```json
{
  "success": true,
  "data": {
    "task_id": "任务ID",
    "type": "image|video|text",
    "status": "pending|processing|completed|failed",
    "created": "创建时间",
    "updated": "更新时间",
    
    // 结果字段（任务完成时）
    "image_url": "图像URL",
    "video_url": "视频URL", 
    "text_result": "文本结果"
  },
  "message": "任务完成"
}
```

### 📋 支持的模型和参数

#### 火山引擎模型

| 模型类型 | 模型名称 | 支持参数 |
|---------|---------|---------|
| 豆包图像 | `doubao-seedream-3-0-t2i-250415` | size: 1024x1024, 864x1152, 1152x864, 1280x720, 720x1280, 832x1248, 1248x832, 1512x648 |
| 即梦AI图像 | `jimeng_high_aes_general_v21_L` | size: 512x512, 512x384, 384x512, 512x341, 341x512, 512x288, 288x512 |
| 即梦AI视频 | `jimeng_vgfm_t2v_l20` | aspect_ratio: 16:9, 9:16, 1:1, 4:3, 3:4, 21:9; seed: 随机种子 |
| 豆包文本 | `doubao-pro-4k` | max_tokens: 最大令牌数; temperature: 温度参数 |

#### OpenAI模型（示例扩展）

| 模型类型 | 模型名称 | 支持参数 |
|---------|---------|---------|
| DALL-E图像 | `dall-e-3` | size: 1024x1024, 1024x1792, 1792x1024 |
| GPT文本 | `gpt-4` | max_tokens, temperature |
| Sora视频 | `sora` | aspect_ratio: 16:9, 9:16, 1:1 |

### 💡 使用示例

#### 创建图像生成任务

```bash
curl -X POST http://localhost:8080/api/v1/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪在花园里玩耍",
    "user_id": "user123",
    "provider": "volcengine",
    "model": "doubao-seedream-3-0-t2i-250415",
    "size": "1024x1024"
  }'
```

#### 查询任务状态

```bash
curl http://localhost:8080/api/v1/ai/task/result/task_id_here
```

#### 获取用户任务列表

```bash
# 获取所有任务
curl "http://localhost:8080/api/v1/ai/tasks?user_id=user123&limit=10&offset=0"

# 只获取图像任务
curl "http://localhost:8080/api/v1/ai/tasks?user_id=user123&type=image&limit=10&offset=0"
```

### 🔧 API设计优势

- **统一接口** - 所有任务类型使用相同的API模式
- **线性扩展** - 新增任务类型只需添加一个创建接口
- **类型无关** - 查询、删除、列表接口与任务类型无关
- **参数明确** - 必填的model和provider字段确保调用明确性
- **易于集成** - 客户端只需记住统一的调用模式

## 🏗️ 架构设计理念

### Provider-Service 分层模式

```go
// Provider层 - 任务分发
type VolcengineProvider struct {
    volcengineService *service.VolcengineService
    taskService       *service.TaskService
}

func (p *VolcengineProvider) DispatchImageTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
    switch model {
    case "jimeng_high_aes_general_v21_L":
        return p.volcengineService.GenerateImageByJimeng(ctx, taskID, input)
    case "doubao-seedream-3-0-t2i-250415":
        return p.volcengineService.GenerateImageByDoubao(ctx, taskID, input)
    default:
        return fmt.Errorf("不支持的模型: %s", model)
    }
}
```

```go
// Service层 - 具体实现
type VolcengineService struct {
    config      *config.AIConfig
    taskService *TaskService
}

func (s *VolcengineService) GenerateImageByJimeng(ctx context.Context, taskID string, input map[string]interface{}) error {
    // 具体的即梦AI API调用逻辑
    // ...
}

func (s *VolcengineService) GenerateImageByDoubao(ctx context.Context, taskID string, input map[string]interface{}) error {
    // 具体的豆包API调用逻辑
    // ...
}
```

### 核心基础设施

```go
// Core层 - 任务队列系统
type TaskQueue struct {
    client          *asynq.Client
    server          *asynq.Server
    serviceRegistry *ServiceRegistry
    taskService     *service.TaskService
}

// Core层 - 服务注册器
type ServiceRegistry struct {
    dispatchers map[string]AITaskDispatcher
}

func (sr *ServiceRegistry) RegisterDispatcher(dispatcher AITaskDispatcher) {
    sr.dispatchers[dispatcher.GetProviderName()] = dispatcher
}
```

### 扩展新服务商

添加新的AI服务商只需要：

1. **实现Service层**

```go
type NewAIService struct {
    apiKey      string
    taskService *TaskService
}

func (s *NewAIService) GenerateImage(ctx context.Context, taskID string, input map[string]interface{}) error {
    // 新服务商的API调用逻辑
}
```

2. **实现Provider层**

```go
type NewAIProvider struct {
    newAIService *NewAIService
    taskService  *TaskService
}

func (p *NewAIProvider) DispatchImageTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
    return p.newAIService.GenerateImage(ctx, taskID, input)
}
```

3. **注册到系统**

```go
newAIService := service.NewNewAIService(apiKey, taskService)
newAIProvider := provider.NewNewAIProvider(newAIService, taskService)
serviceRegistry.RegisterDispatcher(newAIProvider)
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

### v2.0.0 - 现代化分层架构重构

- ✅ **Core基础设施层** - 将任务队列系统移至core目录，作为核心基础设施
- ✅ **Provider-Service分层** - 完全分离任务分发和具体实现职责
- ✅ **接口驱动设计** - 基于AITaskDispatcher接口的插件化架构
- ✅ **目录结构优化** - 更清晰的分层目录结构，职责明确
- ✅ **TaskQueue重命名** - 从RedisQueue重命名为TaskQueue，更通用的命名
- ✅ **架构文档完善** - 详细的架构设计理念和扩展指南

### v1.6.0 - 极简API设计

- ✅ **移除兼容性接口** - 去掉所有类型特定的任务列表接口，采用最简洁的设计
- ✅ **统一任务列表** - 只保留一个`GET /ai/tasks`接口，通过`type`参数过滤
- ✅ **极简路由** - 总共只有6个API接口（3个创建 + 3个统一）
- ✅ **线性扩展** - API数量随任务类型线性增长（N+3），而非传统的平方增长（N×4）
- ✅ **model字段必填** - 移除默认模型机制，要求用户明确指定AI模型
- ✅ **provider字段必填** - 移除默认提供商机制，要求用户明确指定AI服务提供商

### v1.5.0 - 统一任务管理架构重构

- ✅ **统一Task模型** - 合并ImageTask和VideoTask为统一的Task模型
- ✅ **统一TaskService** - 一个服务处理所有类型的任务（图像、视频、文本）
- ✅ **简化API设计** - 统一的任务查询和删除接口，无需区分任务类型
- ✅ **优化数据库设计** - 单一tasks集合存储所有任务，减少复杂性
- ✅ **扩展性增强** - 新增任务类型时只需线性增长API数量

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

## 📞 联系我们

- 项目主页: [GitHub Repository](https://github.com/your-org/volcengine-go-server)
- 问题反馈: [Issues](https://github.com/your-org/volcengine-go-server/issues)
- 邮箱: <zimin.zhang2000@gmail.com>

---

## 🐳 Docker 部署

### 单独构建和运行

```bash
# 构建API服务器镜像
make docker-build-server

# 构建Worker服务镜像
make docker-build-worker

# 构建所有镜像
make docker-build-all

# 运行API服务器容器（暴露8080端口）
make docker-run-server

# 运行Worker服务容器（不暴露端口，后台任务处理）
make docker-run-worker
```

### 使用Docker Compose

```bash
# 启动完整服务栈（推荐）
make docker-compose-up

# 停止所有服务
make docker-compose-down

# 启动包含监控服务的完整栈
docker-compose --profile monitoring up -d
```

### 服务架构说明

- **API服务器**: 暴露8080端口，提供HTTP API服务
- **Worker服务**: 不暴露端口，通过Redis队列处理后台任务
- **MongoDB**: 数据持久化存储
- **Redis**: 缓存和任务队列
- **监控服务**: 可选的Prometheus + Grafana监控栈
