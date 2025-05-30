package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/queue"
	"volcengine-go-server/internal/service"
)

// 通用AI任务请求结构
type AITaskRequest struct {
	Prompt   string `json:"prompt" binding:"required"`
	Model    string `json:"model" binding:"required"` // 设为必填字段
	UserID   string `json:"user_id" binding:"required"`
	Provider string `json:"provider" binding:"required"` // 设为必填字段

	// 图像生成特有字段
	Size string `json:"size,omitempty"`

	// 文本生成特有字段
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`

	// 视频生成特有字段
	Duration    int    `json:"duration,omitempty"`
	ReqKey      string `json:"req_key,omitempty"`      // 服务标识
	Seed        int64  `json:"seed,omitempty"`         // 随机种子
	AspectRatio string `json:"aspect_ratio,omitempty"` // 视频尺寸比例
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
	taskService  *service.TaskService
	queueService *queue.RedisQueue
}

func NewAITaskFactory(taskService *service.TaskService, queueService *queue.RedisQueue) *AITaskFactory {
	return &AITaskFactory{
		taskService:  taskService,
		queueService: queueService,
	}
}

// 创建AI任务的通用方法
func (f *AITaskFactory) CreateTask(c *gin.Context, taskType AITaskType) {
	var req AITaskRequest
	if errors := ValidateRequest(c, &req); len(errors) > 0 {
		ResponseValidationError(c, errors)
		return
	}

	// 验证model字段是否为空
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "model字段不能为空",
			"message": "请指定要使用的AI模型",
		})
		return
	}

	// 验证provider字段是否为空
	if req.Provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "provider字段不能为空",
			"message": "请指定要使用的AI服务提供商",
		})
		return
	}

	provider := req.Provider
	model := req.Model

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
	input := &service.TaskInput{
		Prompt:   req.Prompt,
		UserID:   req.UserID,
		Type:     "image",
		Model:    model,
		Provider: provider,
		Size:     req.Size,
	}

	// 在任务系统中创建记录
	task, err := f.taskService.CreateTask(c.Request.Context(), input)
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
			"prompt": req.Prompt,
			"size":   req.Size,
		},
	}

	// 将任务放入Redis队列
	if err := f.queueService.EnqueueTask(c.Request.Context(), queue.TypeImageGeneration, payload); err != nil {
		// 如果入队失败，删除已创建的任务记录
		f.taskService.DeleteTask(c.Request.Context(), task.ID)
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
	// 创建视频任务输入
	input := &service.TaskInput{
		Prompt:      req.Prompt,
		UserID:      req.UserID,
		Type:        "video",
		Model:       model,
		Provider:    provider,
		ReqKey:      req.ReqKey,
		Seed:        req.Seed,
		AspectRatio: req.AspectRatio,
	}

	// 在任务系统中创建记录
	task, err := f.taskService.CreateTask(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建视频任务记录失败",
			"message": err.Error(),
		})
		return
	}

	// 构建队列任务载荷
	payload := &queue.AITaskPayload{
		TaskID:   task.ID,
		UserID:   req.UserID,
		Type:     string(TaskTypeVideo) + "_generation",
		Provider: provider,
		Model:    model,
		Input: map[string]interface{}{
			"prompt":       req.Prompt,
			"req_key":      task.ReqKey,
			"seed":         task.Seed,
			"aspect_ratio": task.AspectRatio,
		},
	}

	// 将任务放入Redis队列
	if err := f.queueService.EnqueueTask(c.Request.Context(), queue.TypeVideoGeneration, payload); err != nil {
		// 如果入队失败，删除已创建的任务记录
		f.taskService.DeleteTask(c.Request.Context(), task.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "任务入队失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":      task.ID,
			"status":       config.TaskStatusPending,
			"provider":     provider,
			"model":        model,
			"req_key":      task.ReqKey,
			"seed":         task.Seed,
			"aspect_ratio": task.AspectRatio,
		},
		"message": "视频生成任务创建成功",
	})
}
