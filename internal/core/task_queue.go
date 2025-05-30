package core

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

// TaskQueue 任务队列系统
type TaskQueue struct {
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

// NewTaskQueue 创建新的任务队列
func NewTaskQueue(
	redisURL string,
	taskService *service.TaskService,
	serviceRegistry *ServiceRegistry,
) *TaskQueue {
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

	return &TaskQueue{
		client:          client,
		server:          server,
		opt:             opt,
		serviceRegistry: serviceRegistry,
		taskService:     taskService,
	}
}

// 入队任务
func (r *TaskQueue) EnqueueTask(ctx context.Context, taskType string, payload *AITaskPayload, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, data, opts...)
	_, err = r.client.Enqueue(task)
	return err
}

// 入队延迟任务
func (r *TaskQueue) EnqueueDelayedTask(ctx context.Context, taskType string, payload *AITaskPayload, delay time.Duration, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, data, opts...)
	_, err = r.client.Enqueue(task, asynq.ProcessIn(delay))
	return err
}

// 入队定时任务
func (r *TaskQueue) EnqueueScheduledTask(ctx context.Context, taskType string, payload *AITaskPayload, processAt time.Time, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(taskType, data, opts...)
	_, err = r.client.Enqueue(task, asynq.ProcessAt(processAt))
	return err
}

// 启动队列工作器
func (r *TaskQueue) StartWorker(ctx context.Context) {
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
func (r *TaskQueue) StopWorker() {
	r.server.Shutdown()
}

// 关闭队列客户端
func (r *TaskQueue) Close() error {
	return r.client.Close()
}

// 文本生成任务处理器
func (r *TaskQueue) handleTextGeneration(ctx context.Context, task *asynq.Task) error {
	var payload AITaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	logrus.Infof("处理文本生成任务: %s, 用户: %s, 提供商: %s", payload.TaskID, payload.UserID, payload.Provider)

	// 获取对应的AI任务分发器
	dispatcher, exists := r.serviceRegistry.GetDispatcher(payload.Provider)
	if !exists {
		errorMsg := fmt.Sprintf("未找到AI任务分发器: %s", payload.Provider)
		logrus.Errorf(errorMsg)
		// 文本任务暂无数据库状态管理，返回SkipRetry错误让任务被正确归档
		return fmt.Errorf("未找到AI任务分发器: %s: %w", payload.Provider, asynq.SkipRetry)
	}

	// 调用分发器的文本生成分发方法
	if err := dispatcher.DispatchTextTask(ctx, payload.TaskID, payload.Model, payload.Input); err != nil {
		logrus.Errorf("文本生成任务分发失败: %v", err)
		// 文本任务暂无数据库状态管理，返回SkipRetry错误让任务被正确归档
		return fmt.Errorf("文本生成任务分发失败: %v: %w", err, asynq.SkipRetry)
	}

	logrus.Infof("文本生成任务完成: %s", payload.TaskID)
	return nil
}

// 图像生成任务处理器
func (r *TaskQueue) handleImageGeneration(ctx context.Context, task *asynq.Task) error {
	var payload AITaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	logrus.Infof("处理图像生成任务: %s, 用户: %s, 提供商: %s", payload.TaskID, payload.UserID, payload.Provider)

	// 获取对应的AI任务分发器
	dispatcher, exists := r.serviceRegistry.GetDispatcher(payload.Provider)
	if !exists {
		errorMsg := fmt.Sprintf("未找到AI任务分发器: %s", payload.Provider)
		logrus.Errorf(errorMsg)
		r.taskService.UpdateTaskError(ctx, payload.TaskID, errorMsg)
		return fmt.Errorf("未找到AI任务分发器: %s: %w", payload.Provider, asynq.SkipRetry)
	}

	// 调用分发器的图像生成分发方法
	if err := dispatcher.DispatchImageTask(ctx, payload.TaskID, payload.Model, payload.Input); err != nil {
		logrus.Errorf("图像生成任务分发失败: %v", err)
		return err // 让任务重试
	}

	logrus.Infof("图像生成任务完成: %s", payload.TaskID)
	return nil
}

// 视频生成任务处理器
func (r *TaskQueue) handleVideoGeneration(ctx context.Context, task *asynq.Task) error {
	var payload AITaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	logrus.Infof("处理视频生成任务: %s, 用户: %s, 提供商: %s", payload.TaskID, payload.UserID, payload.Provider)

	// 获取对应的AI任务分发器
	dispatcher, exists := r.serviceRegistry.GetDispatcher(payload.Provider)
	if !exists {
		errorMsg := fmt.Sprintf("未找到AI任务分发器: %s", payload.Provider)
		logrus.Errorf(errorMsg)
		r.taskService.UpdateTaskError(ctx, payload.TaskID, errorMsg)
		return fmt.Errorf("未找到AI任务分发器: %s: %w", payload.Provider, asynq.SkipRetry)
	}

	// 调用分发器的视频生成分发方法
	if err := dispatcher.DispatchVideoTask(ctx, payload.TaskID, payload.Model, payload.Input); err != nil {
		logrus.Errorf("视频生成任务分发失败: %v", err)
		return err // 让任务重试
	}

	logrus.Infof("视频生成任务完成: %s", payload.TaskID)
	return nil
}

// 获取队列统计信息
func (r *TaskQueue) GetQueueStats(ctx context.Context) (*QueueStats, error) {
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
