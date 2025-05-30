package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/models"
	"volcengine-go-server/internal/repository"
)

// TaskService 统一任务服务 - 业务逻辑层
type TaskService struct {
	taskRepo repository.TaskRepository
}

// NewTaskService 创建任务服务
func NewTaskService(db repository.Database) *TaskService {
	return &TaskService{
		taskRepo: db.TaskRepository(),
	}
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(ctx context.Context, input *models.TaskInput) (*models.Task, error) {
	task := &models.Task{
		ID:       primitive.NewObjectID().Hex(),
		UserID:   input.UserID,
		Type:     input.Type,
		Prompt:   input.Prompt,
		Model:    input.Model,
		Provider: input.Provider,
		Status:   config.TaskStatusPending,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	// 根据任务类型设置特有字段和默认值
	switch input.Type {
	case models.TaskTypeImage:
		task.AspectRatio = input.AspectRatio
		task.N = input.N
		if task.N == 0 {
			task.N = 1
		}
	case models.TaskTypeVideo:
		task.Seed = input.Seed
		task.AspectRatio = input.AspectRatio
		// 设置默认值
		if task.Seed == 0 {
			task.Seed = config.DefaultVideoSeed
		}
		if task.AspectRatio == "" {
			task.AspectRatio = config.DefaultVideoAspectRatio
		}
	case models.TaskTypeText:
		task.MaxTokens = input.MaxTokens
		task.Temperature = input.Temperature
	}

	err := s.taskRepo.CreateTask(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask 获取任务
func (s *TaskService) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	return s.taskRepo.GetTaskByID(ctx, taskID)
}

// UpdateTaskStatus 更新任务状态
func (s *TaskService) UpdateTaskStatus(ctx context.Context, taskID, status string) error {
	return s.taskRepo.UpdateTaskStatus(ctx, taskID, status)
}

// UpdateTaskResult 更新任务结果
func (s *TaskService) UpdateTaskResult(ctx context.Context, taskID, resultURL string) error {
	// 首先获取任务以确定类型
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	return s.taskRepo.UpdateTaskResult(ctx, taskID, task, resultURL)
}

// UpdateTaskError 更新任务错误
func (s *TaskService) UpdateTaskError(ctx context.Context, taskID, errorMsg string) error {
	return s.taskRepo.UpdateTaskError(ctx, taskID, errorMsg)
}

// GetUserTasks 获取用户任务列表
func (s *TaskService) GetUserTasks(ctx context.Context, userID string, taskType string, limit, offset int) ([]*models.Task, error) {
	return s.taskRepo.GetTasksByUserID(ctx, userID, taskType, limit, offset)
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(ctx context.Context, taskID string) error {
	return s.taskRepo.DeleteTask(ctx, taskID)
}
