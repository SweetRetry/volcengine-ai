package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jimeng-go-server/internal/config"
	"jimeng-go-server/internal/queue"
	"jimeng-go-server/internal/service"
)

// 通用AI任务请求结构
type AITaskRequest struct {
	Prompt   string `json:"prompt" binding:"required"`
	Model    string `json:"model"`
	UserID   string `json:"user_id" binding:"required"`
	Provider string `json:"provider"`

	// 图像生成特有字段
	N       int    `json:"n,omitempty"`
	Size    string `json:"size,omitempty"`
	Quality string `json:"quality,omitempty"`
	Style   string `json:"style,omitempty"`

	// 文本生成特有字段
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`

	// 视频生成特有字段
	Duration int `json:"duration,omitempty"`
}

// AI任务类型
type AITaskType string

const (
	TaskTypeImage AITaskType = "image"
	TaskTypeText  AITaskType = "text"
	TaskTypeVideo AITaskType = "video"
)

// AI任务工厂
type AITaskFactory struct {
	imageTaskService *service.ImageTaskService
	queueService     *queue.RedisQueue
}

func NewAITaskFactory(imageTaskService *service.ImageTaskService, queueService *queue.RedisQueue) *AITaskFactory {
	return &AITaskFactory{
		imageTaskService: imageTaskService,
		queueService:     queueService,
	}
}

// 获取默认提供商
func (f *AITaskFactory) getDefaultProvider(req *AITaskRequest) string {
	if req.Provider != "" {
		return req.Provider
	}
	return config.DefaultAIProvider
}

// 获取默认模型
func (f *AITaskFactory) getDefaultModel(provider string, taskType AITaskType) string {
	if provider == "" {
		provider = config.DefaultAIProvider
	}

	switch provider {
	case "volcengine":
		switch taskType {
		case TaskTypeImage:
			return config.VolcengineImageModel
		case TaskTypeText:
			return config.VolcengineTextModel
		case TaskTypeVideo:
			return config.VolcengineVideoModel
		}
	case "openai":
		switch taskType {
		case TaskTypeImage:
			return config.OpenAIImageModel
		case TaskTypeText:
			return config.OpenAITextModel
		case TaskTypeVideo:
			return config.OpenAIVideoModel
		}
	}

	// 默认返回火山引擎图像模型
	return config.VolcengineImageModel
}

// 创建AI任务的通用方法
func (f *AITaskFactory) CreateTask(c *gin.Context, taskType AITaskType) {
	var req AITaskRequest
	if errors := ValidateRequest(c, &req); len(errors) > 0 {
		ResponseValidationError(c, errors)
		return
	}

	provider := f.getDefaultProvider(&req)
	model := req.Model
	if model == "" {
		model = f.getDefaultModel(provider, taskType)
	}

	switch taskType {
	case TaskTypeImage:
		f.createImageTask(c, &req, provider, model)
	case TaskTypeText:
		f.createTextTask(c, &req, provider, model)
	case TaskTypeVideo:
		f.createVideoTask(c, &req, provider, model)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不支持的任务类型",
		})
	}
}

// 创建图像任务
func (f *AITaskFactory) createImageTask(c *gin.Context, req *AITaskRequest, provider, model string) {
	// 创建图像任务输入
	input := &service.ImageTaskInput{
		Prompt:  req.Prompt,
		UserID:  req.UserID,
		Model:   model,
		Size:    req.Size,
		Quality: req.Quality,
		Style:   req.Style,
		N:       req.N,
	}

	// 在任务系统中创建记录
	task, err := f.imageTaskService.CreateImageTask(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建图像任务记录失败",
			"message": err.Error(),
		})
		return
	}

	// 构建队列任务载荷
	payload := &queue.AITaskPayload{
		TaskID:   task.ID,
		UserID:   req.UserID,
		Type:     string(TaskTypeImage) + "_generation",
		Provider: provider,
		Model:    model,
		Input: map[string]interface{}{
			"prompt":  req.Prompt,
			"size":    req.Size,
			"quality": req.Quality,
			"style":   req.Style,
			"n":       req.N,
		},
	}

	// 将任务放入Redis队列
	if err := f.queueService.EnqueueTask(c.Request.Context(), queue.TypeImageGeneration, payload); err != nil {
		// 如果入队失败，删除已创建的任务记录
		f.imageTaskService.DeleteImageTask(c.Request.Context(), task.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "任务入队失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":  task.ID,
			"status":   config.TaskStatusPending,
			"provider": provider,
			"model":    model,
		},
		"message": "图像生成任务创建成功",
	})
}

// 创建文本任务
func (f *AITaskFactory) createTextTask(c *gin.Context, req *AITaskRequest, provider, model string) {
	// TODO: 实现文本任务创建逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "文本生成功能暂未实现",
		"message": "该功能正在开发中，敬请期待",
		"data": gin.H{
			"provider": provider,
			"model":    model,
			"prompt":   req.Prompt,
		},
	})
}

// 创建视频任务
func (f *AITaskFactory) createVideoTask(c *gin.Context, req *AITaskRequest, provider, model string) {
	// TODO: 实现视频任务创建逻辑
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "视频生成功能暂未实现",
		"message": "该功能正在开发中，敬请期待",
		"data": gin.H{
			"provider": provider,
			"model":    model,
			"prompt":   req.Prompt,
			"duration": req.Duration,
		},
	})
}
