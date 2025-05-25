package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type RedisQueue struct {
	client *asynq.Client
	server *asynq.Server
	opt    asynq.RedisConnOpt
	// 添加服务依赖，用于更新任务状态
	imageTaskService interface {
		UpdateImageTaskStatus(ctx context.Context, taskID, status string, imageURL, errorMsg string) error
	}
}

// 任务类型常量
const (
	TypeTextGeneration  = "ai:text_generation"
	TypeImageGeneration = "ai:image_generation"
	TypeTranslation     = "ai:translation"
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

func NewRedisQueue(redisURL string, imageTaskService interface {
	UpdateImageTaskStatus(ctx context.Context, taskID, status string, imageURL, errorMsg string) error
},
) *RedisQueue {
	// 解析Redis URL
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		logrus.Fatal("解析Redis URL失败: ", err)
	}

	client := asynq.NewClient(opt)

	server := asynq.NewServer(opt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logrus.Errorf("任务执行失败: %v, 错误: %v", task.Type(), err)
		}),
		Logger: logrus.New(),
	})

	return &RedisQueue{
		client:           client,
		server:           server,
		opt:              opt,
		imageTaskService: imageTaskService,
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
	mux.HandleFunc(TypeTranslation, r.handleTranslation)

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

	logrus.Infof("处理文本生成任务: %s, 用户: %s", payload.TaskID, payload.UserID)

	// 这里会在service层实现具体的AI调用逻辑
	// 现在只是记录日志
	time.Sleep(2 * time.Second) // 模拟AI处理时间

	logrus.Infof("文本生成任务完成: %s", payload.TaskID)
	return nil
}

// 图像生成任务处理器
func (r *RedisQueue) handleImageGeneration(ctx context.Context, task *asynq.Task) error {
	var payload AITaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	logrus.Infof("处理图像生成任务: %s, 用户: %s", payload.TaskID, payload.UserID)

	// 这里应该调用实际的AI服务
	// 为了演示，我们模拟处理过程
	time.Sleep(5 * time.Second)

	// 模拟生成的图像URL
	imageURL := fmt.Sprintf("https://example.com/generated-image-%s.jpg", payload.TaskID)

	logrus.Infof("图像生成任务完成: %s, 图像URL: %s", payload.TaskID, imageURL)

	// 更新数据库中的任务状态
	if r.imageTaskService != nil {
		if err := r.imageTaskService.UpdateImageTaskStatus(ctx, payload.TaskID, "completed", imageURL, ""); err != nil {
			logrus.Errorf("更新任务状态失败: %v", err)
			return err
		}
		logrus.Infof("任务状态已更新为完成: %s", payload.TaskID)
	}

	return nil
}

// 翻译任务处理器
func (r *RedisQueue) handleTranslation(ctx context.Context, task *asynq.Task) error {
	var payload AITaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	logrus.Infof("处理翻译任务: %s, 用户: %s", payload.TaskID, payload.UserID)

	// 模拟翻译处理时间
	time.Sleep(1 * time.Second)

	logrus.Infof("翻译任务完成: %s", payload.TaskID)
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
