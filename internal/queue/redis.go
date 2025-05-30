package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/service"
)

// AIServiceProvider AI服务提供商接口
type AIServiceProvider interface {
	// 获取提供商名称
	GetProviderName() string
	// 处理图像生成任务
	ProcessImageTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error
	// 处理文本生成任务
	ProcessTextTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error
	// 处理视频生成任务
	ProcessVideoTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error
}

// 服务注册器
type ServiceRegistry struct {
	providers map[string]AIServiceProvider
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		providers: make(map[string]AIServiceProvider),
	}
}

func (sr *ServiceRegistry) RegisterProvider(provider AIServiceProvider) {
	sr.providers[provider.GetProviderName()] = provider
}

func (sr *ServiceRegistry) GetProvider(name string) (AIServiceProvider, bool) {
	provider, exists := sr.providers[name]
	return provider, exists
}

func (sr *ServiceRegistry) GetAllProviders() map[string]AIServiceProvider {
	return sr.providers
}

type RedisQueue struct {
	client *asynq.Client
	server *asynq.Server
	opt    asynq.RedisConnOpt
	// 使用服务注册器替代具体的服务依赖
	serviceRegistry *ServiceRegistry
	taskService     *service.TaskService
}

// 任务类型常量
const (
	TypeTextGeneration  = "ai:text_generation"
	TypeImageGeneration = "ai:image_generation"
	TypeVideoGeneration = "ai:video_generation"
)

// 任务载荷结构
type AITaskPayload struct {
	TaskID   string                 `json:"task_id"`
	UserID   string                 `json:"user_id"`
	Type     string                 `json:"type"`
	Input    map[string]interface{} `json:"input"`
	Model    string                 `json:"model"`
	Provider string                 `json:"provider"`
}

func NewRedisQueue(
	redisURL string,
	taskService *service.TaskService,
	serviceRegistry *ServiceRegistry,
) *RedisQueue {
	// 解析Redis URL
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		logrus.Fatal("解析Redis URL失败: ", err)
	}

	client := asynq.NewClient(opt)

	server := asynq.NewServer(opt, asynq.Config{
		Concurrency: config.QueueConcurrency,
		Queues: map[string]int{
			"critical": config.QueueCriticalWeight,
			"default":  config.QueueDefaultWeight,
			"low":      config.QueueLowWeight,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logrus.Errorf("任务执行失败: %v, 错误: %v", task.Type(), err)
		}),
		Logger: logrus.New(),
	})

	return &RedisQueue{
		client:          client,
		server:          server,
		opt:             opt,
		serviceRegistry: serviceRegistry,
		taskService:     taskService,
	}
}

// 入队任务
func (r *RedisQueue) EnqueueTask(ctx context.Context, taskType string, payload *AITaskPayload, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, data, opts...)
	_, err = r.client.Enqueue(task)
	return err
}

// 入队延迟任务
func (r *RedisQueue) EnqueueDelayedTask(ctx context.Context, taskType string, payload *AITaskPayload, delay time.Duration, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, data, opts...)
	_, err = r.client.Enqueue(task, asynq.ProcessIn(delay))
	return err
}

// 入队定时任务
func (r *RedisQueue) EnqueueScheduledTask(ctx context.Context, taskType string, payload *AITaskPayload, processAt time.Time, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, data, opts...)
	_, err = r.client.Enqueue(task, asynq.ProcessAt(processAt))
	return err
}

// 启动队列工作器
func (r *RedisQueue) StartWorker(ctx context.Context) {
	mux := asynq.NewServeMux()

	// 注册任务处理器
	mux.HandleFunc(TypeTextGeneration, r.handleTextGeneration)
	mux.HandleFunc(TypeImageGeneration, r.handleImageGeneration)
	mux.HandleFunc(TypeVideoGeneration, r.handleVideoGeneration)

	logrus.Info("队列工作器启动中...")
	if err := r.server.Start(mux); err != nil {
		logrus.Fatal("启动队列工作器失败: ", err)
	}
}

// 停止队列工作器
func (r *RedisQueue) StopWorker() {
	r.server.Shutdown()
}

// 关闭队列客户端
func (r *RedisQueue) Close() error {
	return r.client.Close()
}

// 文本生成任务处理器
func (r *RedisQueue) handleTextGeneration(ctx context.Context, task *asynq.Task) error {
	var payload AITaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	logrus.Infof("处理文本生成任务: %s, 用户: %s, 提供商: %s", payload.TaskID, payload.UserID, payload.Provider)

	// 获取对应的AI服务提供商
	provider, exists := r.serviceRegistry.GetProvider(payload.Provider)
	if !exists {
		errorMsg := fmt.Sprintf("未找到AI服务提供商: %s", payload.Provider)
		logrus.Errorf(errorMsg)
		// 文本任务暂无数据库状态管理，返回SkipRetry错误让任务被正确归档
		return fmt.Errorf("未找到AI服务提供商: %s: %w", payload.Provider, asynq.SkipRetry)
	}

	// 调用提供商的文本生成处理方法
	if err := provider.ProcessTextTask(ctx, payload.TaskID, payload.Model, payload.Input); err != nil {
		logrus.Errorf("文本生成任务处理失败: %v", err)
		// 文本任务暂无数据库状态管理，返回SkipRetry错误让任务被正确归档
		return fmt.Errorf("文本生成任务处理失败: %v: %w", err, asynq.SkipRetry)
	}

	logrus.Infof("文本生成任务完成: %s", payload.TaskID)
	return nil
}

// 图像生成任务处理器
func (r *RedisQueue) handleImageGeneration(ctx context.Context, task *asynq.Task) error {
	var payload AITaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	logrus.Infof("处理图像生成任务: %s, 用户: %s, 提供商: %s", payload.TaskID, payload.UserID, payload.Provider)

	// 获取对应的AI服务提供商
	provider, exists := r.serviceRegistry.GetProvider(payload.Provider)
	if !exists {
		errorMsg := fmt.Sprintf("未找到AI服务提供商: %s", payload.Provider)
		logrus.Errorf(errorMsg)
		r.taskService.UpdateTaskError(ctx, payload.TaskID, errorMsg)
		// 返回SkipRetry错误，让asynq将任务标记为archived而不是processed
		return fmt.Errorf("未找到AI服务提供商: %s: %w", payload.Provider, asynq.SkipRetry)
	}

	// 调用提供商的图像生成处理方法
	if err := provider.ProcessImageTask(ctx, payload.TaskID, payload.Model, payload.Input); err != nil {
		logrus.Errorf("图像生成任务处理失败: %v", err)
		r.taskService.UpdateTaskError(ctx, payload.TaskID, err.Error())
		// 返回SkipRetry错误，让asynq将任务标记为archived而不是processed
		return fmt.Errorf("图像生成任务处理失败: %v: %w", err, asynq.SkipRetry)
	}

	logrus.Infof("图像生成任务完成: %s", payload.TaskID)
	return nil
}

// 视频生成任务处理器
func (r *RedisQueue) handleVideoGeneration(ctx context.Context, task *asynq.Task) error {
	var payload AITaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	logrus.Infof("处理视频生成任务: %s, 用户: %s, 提供商: %s", payload.TaskID, payload.UserID, payload.Provider)

	// 获取对应的AI服务提供商
	provider, exists := r.serviceRegistry.GetProvider(payload.Provider)
	if !exists {
		errorMsg := fmt.Sprintf("未找到AI服务提供商: %s", payload.Provider)
		logrus.Errorf(errorMsg)
		r.taskService.UpdateTaskError(ctx, payload.TaskID, errorMsg)
		return fmt.Errorf("未找到AI服务提供商: %s: %w", payload.Provider, asynq.SkipRetry)
	}

	// 调用提供商的视频生成处理方法
	if err := provider.ProcessVideoTask(ctx, payload.TaskID, payload.Model, payload.Input); err != nil {
		logrus.Errorf("视频生成任务处理失败: %v", err)
		r.taskService.UpdateTaskError(ctx, payload.TaskID, err.Error())
		return fmt.Errorf("视频生成任务处理失败: %v: %w", err, asynq.SkipRetry)
	}

	logrus.Infof("视频生成任务完成: %s", payload.TaskID)
	return nil
}

// 获取队列统计信息
func (r *RedisQueue) GetQueueStats(ctx context.Context) (*QueueStats, error) {
	inspector := asynq.NewInspector(r.opt)
	defer inspector.Close()

	stats, err := inspector.GetQueueInfo("default")
	if err != nil {
		return nil, err
	}

	return &QueueStats{
		Pending:   stats.Pending,
		Active:    stats.Active,
		Scheduled: stats.Scheduled,
		Retry:     stats.Retry,
		Archived:  stats.Archived,
		Completed: stats.Completed,
		Processed: stats.Processed,
		Failed:    stats.Failed,
		Timestamp: time.Now(),
	}, nil
}

// 队列统计信息结构
type QueueStats struct {
	Pending   int       `json:"pending"`
	Active    int       `json:"active"`
	Scheduled int       `json:"scheduled"`
	Retry     int       `json:"retry"`
	Archived  int       `json:"archived"`
	Completed int       `json:"completed"`
	Processed int       `json:"processed"`
	Failed    int       `json:"failed"`
	Timestamp time.Time `json:"timestamp"`
}
