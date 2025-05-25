package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"jimeng-go-server/internal/database"
	"jimeng-go-server/internal/queue"
)

type TaskService struct {
	db    database.Database
	queue *queue.RedisQueue
}

func NewTaskService(db database.Database, queue *queue.RedisQueue) *TaskService {
	return &TaskService{
		db:    db,
		queue: queue,
	}
}

// 创建AI任务
func (s *TaskService) CreateAITask(ctx context.Context, userID, taskType string, input map[string]interface{}, model, provider string) (*database.Task, error) {
	// 创建任务记录
	task := &database.Task{
		UserID: userID,
		Type:   taskType,
		Status: "pending",
		Input:  fmt.Sprintf("%v", input), // 简单序列化，实际项目中应使用JSON
	}

	if err := s.db.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	// 将任务加入队列
	payload := &queue.AITaskPayload{
		TaskID:   task.ID,
		UserID:   userID,
		Type:     taskType,
		Input:    input,
		Model:    model,
		Provider: provider,
	}

	var queueType string
	switch taskType {
	case "text_generation":
		queueType = queue.TypeTextGeneration
	case "image_generation":
		queueType = queue.TypeImageGeneration
	case "translation":
		queueType = queue.TypeTranslation
	default:
		queueType = queue.TypeTextGeneration
	}

	// 根据任务类型设置不同的优先级
	opts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	if taskType == "image_generation" {
		opts = append(opts, asynq.Queue("default"))
	} else {
		opts = append(opts, asynq.Queue("critical"))
	}

	if err := s.queue.EnqueueTask(ctx, queueType, payload, opts...); err != nil {
		// 如果入队失败，更新任务状态为失败
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("入队失败: %v", err)
		s.db.UpdateTask(ctx, task)
		return nil, fmt.Errorf("任务入队失败: %w", err)
	}

	return task, nil
}

// 创建延迟任务
func (s *TaskService) CreateDelayedTask(ctx context.Context, userID, taskType string, input map[string]interface{}, model, provider string, delay time.Duration) (*database.Task, error) {
	// 创建任务记录
	task := &database.Task{
		UserID: userID,
		Type:   taskType,
		Status: "scheduled",
		Input:  fmt.Sprintf("%v", input),
	}

	if err := s.db.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("创建延迟任务失败: %w", err)
	}

	// 将任务加入延迟队列
	payload := &queue.AITaskPayload{
		TaskID:   task.ID,
		UserID:   userID,
		Type:     taskType,
		Input:    input,
		Model:    model,
		Provider: provider,
	}

	var queueType string
	switch taskType {
	case "text_generation":
		queueType = queue.TypeTextGeneration
	case "image_generation":
		queueType = queue.TypeImageGeneration
	case "translation":
		queueType = queue.TypeTranslation
	default:
		queueType = queue.TypeTextGeneration
	}

	opts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
		asynq.Queue("default"),
	}

	if err := s.queue.EnqueueDelayedTask(ctx, queueType, payload, delay, opts...); err != nil {
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("延迟任务入队失败: %v", err)
		s.db.UpdateTask(ctx, task)
		return nil, fmt.Errorf("延迟任务入队失败: %w", err)
	}

	return task, nil
}

// 获取任务详情
func (s *TaskService) GetTask(ctx context.Context, taskID string) (*database.Task, error) {
	task, err := s.db.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}
	return task, nil
}

// 获取用户的任务列表
func (s *TaskService) GetUserTasks(ctx context.Context, userID string, limit, offset int) ([]*database.Task, error) {
	tasks, err := s.db.GetTasksByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("获取用户任务列表失败: %w", err)
	}
	return tasks, nil
}

// 更新任务状态
func (s *TaskService) UpdateTaskStatus(ctx context.Context, taskID, status string, output, errorMsg string) error {
	task, err := s.db.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}

	task.Status = status
	task.Output = output
	task.ErrorMsg = errorMsg

	if status == "completed" || status == "failed" {
		now := time.Now()
		task.CompletedAt = &now
	}

	if err := s.db.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	return nil
}

// 取消任务
func (s *TaskService) CancelTask(ctx context.Context, taskID string) error {
	task, err := s.db.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if task.Status == "completed" || task.Status == "failed" || task.Status == "cancelled" {
		return fmt.Errorf("任务已完成或已取消，无法取消")
	}

	task.Status = "cancelled"
	now := time.Now()
	task.CompletedAt = &now

	if err := s.db.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("取消任务失败: %w", err)
	}

	return nil
}

// 重试失败的任务
func (s *TaskService) RetryTask(ctx context.Context, taskID string) error {
	task, err := s.db.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if task.Status != "failed" {
		return fmt.Errorf("只能重试失败的任务")
	}

	// 重置任务状态
	task.Status = "pending"
	task.ErrorMsg = ""
	task.CompletedAt = nil

	if err := s.db.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("重置任务状态失败: %w", err)
	}

	// 重新入队
	payload := &queue.AITaskPayload{
		TaskID: task.ID,
		UserID: task.UserID,
		Type:   task.Type,
		// Input需要重新解析，这里简化处理
	}

	var queueType string
	switch task.Type {
	case "text_generation":
		queueType = queue.TypeTextGeneration
	case "image_generation":
		queueType = queue.TypeImageGeneration
	case "translation":
		queueType = queue.TypeTranslation
	default:
		queueType = queue.TypeTextGeneration
	}

	opts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
		asynq.Queue("default"),
	}

	if err := s.queue.EnqueueTask(ctx, queueType, payload, opts...); err != nil {
		return fmt.Errorf("重新入队失败: %w", err)
	}

	return nil
}

// 获取队列统计信息
func (s *TaskService) GetQueueStats(ctx context.Context) (*queue.QueueStats, error) {
	stats, err := s.queue.GetQueueStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取队列统计失败: %w", err)
	}
	return stats, nil
}

// 删除任务
func (s *TaskService) DeleteTask(ctx context.Context, taskID string) error {
	if err := s.db.DeleteTask(ctx, taskID); err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}
	return nil
} 