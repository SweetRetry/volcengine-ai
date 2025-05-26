# AI服务提供商架构

## 概述

本系统采用**服务注册器模式**来管理多个AI服务提供商，避免了在RedisQueue初始化时传入过多参数的问题，提供了更优雅和可扩展的架构。

## 架构设计

### 核心组件

1. **AIServiceProvider接口** - 定义了所有AI服务提供商必须实现的方法
2. **ServiceRegistry** - 服务注册器，管理所有已注册的AI服务提供商
3. **具体提供商实现** - 如VolcengineAIProvider、OpenAIProvider等

### 接口定义

```go
type AIServiceProvider interface {
    // 获取提供商名称
    GetProviderName() string
    // 处理图像生成任务
    ProcessImageTask(ctx context.Context, taskID string, input map[string]interface{}) error
    // 处理文本生成任务
    ProcessTextTask(ctx context.Context, taskID string, input map[string]interface{}) error
    // 处理视频生成任务
    ProcessVideoTask(ctx context.Context, taskID string, input map[string]interface{}) error
}
```

## 使用方式

### 1. 创建新的AI服务提供商

```go
// 实现AIServiceProvider接口
type MyAIProvider struct {
    apiKey string
    client *MyAIClient
}

func NewMyAIProvider(apiKey string) *MyAIProvider {
    return &MyAIProvider{
        apiKey: apiKey,
        client: NewMyAIClient(apiKey),
    }
}

func (m *MyAIProvider) GetProviderName() string {
    return "my_ai_service"
}

func (m *MyAIProvider) ProcessImageTask(ctx context.Context, taskID string, input map[string]interface{}) error {
    // 实现具体的图像生成逻辑
    return nil
}

// 实现其他方法...
```

### 2. 注册服务提供商

```go
// 在main.go中注册
serviceRegistry := queue.NewServiceRegistry()

// 注册火山引擎提供商
volcengineProvider := service.NewVolcengineAIProvider(volcengineAIService, imageTaskService)
serviceRegistry.RegisterProvider(volcengineProvider)

// 注册OpenAI提供商
openaiProvider := service.NewOpenAIProvider(cfg.OpenAI.APIKey)
serviceRegistry.RegisterProvider(openaiProvider)

// 注册自定义提供商
myProvider := service.NewMyAIProvider(cfg.MyAI.APIKey)
serviceRegistry.RegisterProvider(myProvider)
```

### 3. 使用不同提供商

客户端可以通过`provider`字段指定使用哪个AI服务提供商：

```json
{
    "prompt": "生成一张美丽的风景图",
    "user_id": "user123",
    "provider": "volcengine_jimeng"  // 或 "openai", "my_ai_service"
}
```

## 架构优势

### 🎯 **解决的问题**

1. **参数过多问题**：避免了RedisQueue初始化时需要传入大量AI服务依赖
2. **紧耦合问题**：RedisQueue不再直接依赖具体的AI服务实现
3. **扩展性问题**：添加新的AI服务提供商变得非常简单

### ✨ **架构优势**

1. **高度解耦**：队列系统与具体AI服务实现完全分离
2. **易于扩展**：添加新提供商只需实现接口并注册
3. **统一管理**：所有AI服务提供商通过注册器统一管理
4. **运行时选择**：可以根据请求动态选择不同的AI服务提供商

### 🔄 **工作流程**

```
1. 客户端请求 → AI Handler
2. AI Handler → 创建任务记录 + 入队
3. Redis队列工作器 → 根据provider获取对应的服务提供商
4. 服务提供商 → 调用具体的AI API
5. 服务提供商 → 更新任务状态
```

## 当前支持的提供商

| 提供商名称 | 标识符 | 支持的任务类型 | 状态 |
|-----------|--------|---------------|------|
| 火山引擎即梦 | `volcengine_jimeng` | 图像生成 | ✅ 已实现 |
| OpenAI | `openai` | 图像、文本、视频 | 🚧 示例实现 |

## 扩展指南

### 添加新的AI服务提供商

1. **创建提供商类**：实现`AIServiceProvider`接口
2. **实现具体方法**：根据AI服务的API实现各种任务处理
3. **注册提供商**：在main.go中注册新的提供商
4. **更新配置**：添加相应的配置项（API密钥等）

### 示例：添加百度文心一言

```go
// 1. 创建提供商
type BaiduProvider struct {
    apiKey    string
    secretKey string
}

func (b *BaiduProvider) GetProviderName() string {
    return "baidu_wenxin"
}

// 2. 注册提供商
baiduProvider := service.NewBaiduProvider(cfg.Baidu.APIKey, cfg.Baidu.SecretKey)
serviceRegistry.RegisterProvider(baiduProvider)

// 3. 使用
{
    "prompt": "写一首诗",
    "provider": "baidu_wenxin"
}
```

## 最佳实践

1. **错误处理**：每个提供商应该妥善处理API调用失败的情况
2. **超时控制**：设置合理的超时时间，避免任务无限等待
3. **日志记录**：详细记录任务处理过程，便于调试
4. **配置管理**：将API密钥等敏感信息放在配置文件中
5. **测试覆盖**：为每个提供商编写单元测试

这种架构设计使得系统具有很高的灵活性和可扩展性，可以轻松适应不断变化的AI服务生态。 