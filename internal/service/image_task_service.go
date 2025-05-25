package service

import (
	"context"
	"encoding/json"
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
	Prompt         string `json:"prompt"`
	UserID         string `json:"user_id"`
	Model          string `json:"model,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Style          string `json:"style,omitempty"`
	N              int    `json:"n,omitempty"`
	ExternalTaskID string `json:"external_task_id,omitempty"` // 火山引擎等外部服务的任务ID
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
func (s *ImageTaskService) CreateImageTask(ctx context.Context, input *ImageTaskInput) (*database.Task, error) {
	// 序列化输入参数
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("序列化输入参数失败: %w", err)
	}

	// 创建任务记录
	task := &database.Task{
		UserID:         input.UserID,
		Type:           "image_generation",
		Status:         "pending",
		Input:          string(inputJSON),
		ExternalTaskID: input.ExternalTaskID, // 存储外部任务ID
	}

	if err := s.db.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("创建图像任务失败: %w", err)
	}

	return task, nil
}

// GetImageTask 获取图像任务详情
func (s *ImageTaskService) GetImageTask(ctx context.Context, taskID string) (*ImageTaskResult, error) {
	task, err := s.db.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	if task.Type != "image_generation" {
		return nil, fmt.Errorf("任务类型不匹配，期望image_generation，实际%s", task.Type)
	}

	result := &ImageTaskResult{
		TaskID:  task.ID,
		Status:  task.Status,
		Created: task.CreatedAt,
	}

	// 如果任务完成，解析输出结果
	if task.Status == "completed" && task.Output != "" {
		var output map[string]interface{}
		if err := json.Unmarshal([]byte(task.Output), &output); err == nil {
			if imageURL, ok := output["image_url"].(string); ok {
				result.ImageURL = imageURL
			}
		}
	}

	// 如果任务失败，设置错误信息
	if task.Status == "failed" {
		result.Error = task.ErrorMsg
	}

	return result, nil
}

// GetUserImageTasks 获取用户的图像任务列表
func (s *ImageTaskService) GetUserImageTasks(ctx context.Context, userID string, limit, offset int) ([]*ImageTaskResult, error) {
	tasks, err := s.db.GetTasksByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("获取用户任务列表失败: %w", err)
	}

	var results []*ImageTaskResult
	for _, task := range tasks {
		// 只返回图像生成任务
		if task.Type != "image_generation" {
			continue
		}

		result := &ImageTaskResult{
			TaskID:  task.ID,
			Status:  task.Status,
			Created: task.CreatedAt,
		}

		// 解析输出结果
		if task.Status == "completed" && task.Output != "" {
			var output map[string]interface{}
			if err := json.Unmarshal([]byte(task.Output), &output); err == nil {
				if imageURL, ok := output["image_url"].(string); ok {
					result.ImageURL = imageURL
				}
			}
		}

		if task.Status == "failed" {
			result.Error = task.ErrorMsg
		}

		results = append(results, result)
	}

	return results, nil
}

// UpdateImageTaskStatus 更新图像任务状态
func (s *ImageTaskService) UpdateImageTaskStatus(ctx context.Context, taskID, status string, imageURL, errorMsg string) error {
	task, err := s.db.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if task.Type != "image_generation" {
		return fmt.Errorf("任务类型不匹配")
	}

	task.Status = status
	task.ErrorMsg = errorMsg

	// 如果任务完成，设置输出结果
	if status == "completed" && imageURL != "" {
		output := map[string]interface{}{
			"image_url": imageURL,
		}
		outputJSON, _ := json.Marshal(output)
		task.Output = string(outputJSON)
	}

	if status == "completed" || status == "failed" {
		now := time.Now()
		task.CompletedAt = &now
	}

	if err := s.db.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}

	return nil
}

// GetImageTaskInput 获取图像任务的输入参数
func (s *ImageTaskService) GetImageTaskInput(ctx context.Context, taskID string) (*ImageTaskInput, error) {
	task, err := s.db.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	if task.Type != "image_generation" {
		return nil, fmt.Errorf("任务类型不匹配")
	}

	var input ImageTaskInput
	if err := json.Unmarshal([]byte(task.Input), &input); err != nil {
		return nil, fmt.Errorf("解析任务输入失败: %w", err)
	}

	return &input, nil
}

// DeleteImageTask 删除图像任务
func (s *ImageTaskService) DeleteImageTask(ctx context.Context, taskID string) error {
	task, err := s.db.GetTaskByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if task.Type != "image_generation" {
		return fmt.Errorf("任务类型不匹配")
	}

	if err := s.db.DeleteTask(ctx, taskID); err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}

	return nil
}
