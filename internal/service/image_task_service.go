package service

import (
	"context"
	"fmt"
	"time"

	"jimeng-go-server/internal/database"
)

// ImageTaskService 专门处理图像生成任务的服务
type ImageTaskService struct {
	db database.Database
}

// ImageTaskInput 图像生成任务的输入参数
type ImageTaskInput struct {
	Prompt  string `json:"prompt"`
	UserID  string `json:"user_id"`
	Model   string `json:"model,omitempty"`
	Size    string `json:"size,omitempty"`
	Quality string `json:"quality,omitempty"`
	Style   string `json:"style,omitempty"`
	N       int    `json:"n,omitempty"`
}

// ImageTaskResult 图像生成任务的结果
type ImageTaskResult struct {
	TaskID   string    `json:"task_id"`
	Status   string    `json:"status"`
	ImageURL string    `json:"image_url,omitempty"`
	Error    string    `json:"error,omitempty"`
	Created  time.Time `json:"created"`
}

func NewImageTaskService(db database.Database) *ImageTaskService {
	return &ImageTaskService{
		db: db,
	}
}

// CreateImageTask 创建图像生成任务
func (s *ImageTaskService) CreateImageTask(ctx context.Context, input *ImageTaskInput) (*database.ImageTask, error) {
	// 创建图像任务记录
	task := &database.ImageTask{
		UserID:  input.UserID,
		Prompt:  input.Prompt,
		Model:   input.Model,
		Size:    input.Size,
		Quality: input.Quality,
		Style:   input.Style,
		N:       input.N,
		Status:  "pending",
	}

	if err := s.db.CreateImageTask(ctx, task); err != nil {
		return nil, fmt.Errorf("创建图像任务失败: %w", err)
	}

	return task, nil
}

// GetImageTask 获取图像任务详情
func (s *ImageTaskService) GetImageTask(ctx context.Context, taskID string) (*ImageTaskResult, error) {
	task, err := s.db.GetImageTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	result := &ImageTaskResult{
		TaskID:   task.ID,
		Status:   task.Status,
		ImageURL: task.ImageURL,
		Error:    task.Error,
		Created:  task.Created,
	}

	return result, nil
}

// GetUserImageTasks 获取用户的图像任务列表
func (s *ImageTaskService) GetUserImageTasks(ctx context.Context, userID string, limit, offset int) ([]*ImageTaskResult, error) {
	tasks, err := s.db.GetImageTasksByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("获取用户任务列表失败: %w", err)
	}

	results := make([]*ImageTaskResult, len(tasks))
	for i, task := range tasks {
		results[i] = &ImageTaskResult{
			TaskID:   task.ID,
			Status:   task.Status,
			ImageURL: task.ImageURL,
			Error:    task.Error,
			Created:  task.Created,
		}
	}

	return results, nil
}

// UpdateImageTaskStatus 更新图像任务状态
func (s *ImageTaskService) UpdateImageTaskStatus(ctx context.Context, taskID, status string, imageURL, errorMsg string) error {
	if err := s.db.UpdateImageTaskStatus(ctx, taskID, status, imageURL, errorMsg); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}
	return nil
}

// GetImageTaskInput 获取图像任务的输入参数
func (s *ImageTaskService) GetImageTaskInput(ctx context.Context, taskID string) (*ImageTaskInput, error) {
	task, err := s.db.GetImageTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	input := &ImageTaskInput{
		Prompt:  task.Prompt,
		UserID:  task.UserID,
		Model:   task.Model,
		Size:    task.Size,
		Quality: task.Quality,
		Style:   task.Style,
		N:       task.N,
	}

	return input, nil
}

// DeleteImageTask 删除图像任务
func (s *ImageTaskService) DeleteImageTask(ctx context.Context, taskID string) error {
	if err := s.db.DeleteImageTask(ctx, taskID); err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}
	return nil
}
