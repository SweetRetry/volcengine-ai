package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/core"
	"volcengine-go-server/internal/models"
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

type AIHandler struct {
	taskService  *service.TaskService
	queueService *core.TaskQueue
}

func NewAIHandler(taskService *service.TaskService, queueService *core.TaskQueue) *AIHandler {
	return &AIHandler{
		taskService:  taskService,
		queueService: queueService,
	}
}

// 创建图像任务
func (h *AIHandler) CreateImageTask(c *gin.Context) {
	h.createTask(c, TaskTypeImage)
}

// 创建文本任务
func (h *AIHandler) CreateTextTask(c *gin.Context) {
	h.createTask(c, TaskTypeText)
}

// 创建视频任务
func (h *AIHandler) CreateVideoTask(c *gin.Context) {
	h.createTask(c, TaskTypeVideo)
}

// 创建AI任务的通用方法
func (h *AIHandler) createTask(c *gin.Context, taskType AITaskType) {
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
		h.handleImageTaskCreation(c, &req, provider, model)
	case TaskTypeText:
		h.handleTextTaskCreation(c, &req, provider, model)
	case TaskTypeVideo:
		h.handleVideoTaskCreation(c, &req, provider, model)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不支持的任务类型",
		})
	}
}

// 处理图像任务创建的具体实现
func (h *AIHandler) handleImageTaskCreation(c *gin.Context, req *AITaskRequest, provider, model string) {
	// 创建图像任务输入
	input := &models.TaskInput{
		Prompt:   req.Prompt,
		UserID:   req.UserID,
		Type:     "image",
		Model:    model,
		Provider: provider,
		Size:     req.Size,
	}

	// 在任务系统中创建记录
	task, err := h.taskService.CreateTask(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建图像任务记录失败",
			"message": err.Error(),
		})
		return
	}

	// 构建队列任务载荷
	payload := &core.AITaskPayload{
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
	if err := h.queueService.EnqueueTask(c.Request.Context(), core.TypeImageGeneration, payload); err != nil {
		// 如果入队失败，删除已创建的任务记录
		h.taskService.DeleteTask(c.Request.Context(), task.ID)
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

// 处理文本任务创建的具体实现
func (h *AIHandler) handleTextTaskCreation(c *gin.Context, req *AITaskRequest, provider, model string) {
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

// 处理视频任务创建的具体实现
func (h *AIHandler) handleVideoTaskCreation(c *gin.Context, req *AITaskRequest, provider, model string) {
	// 创建视频任务输入
	input := &models.TaskInput{
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
	task, err := h.taskService.CreateTask(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建视频任务记录失败",
			"message": err.Error(),
		})
		return
	}

	// 构建队列任务载荷
	payload := &core.AITaskPayload{
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
	if err := h.queueService.EnqueueTask(c.Request.Context(), core.TypeVideoGeneration, payload); err != nil {
		// 如果入队失败，删除已创建的任务记录
		h.taskService.DeleteTask(c.Request.Context(), task.ID)
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

// 统一任务结果查询
func (h *AIHandler) GetTaskResult(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务ID不能为空",
		})
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "任务不存在",
			"message": err.Error(),
		})
		return
	}

	h.respondWithTaskResult(c, task)
}

// 统一任务删除
func (h *AIHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务ID不能为空",
		})
		return
	}

	if err := h.taskService.DeleteTask(c.Request.Context(), taskID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "任务不存在",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务删除成功",
	})
}

// 统一任务列表查询
func (h *AIHandler) GetUserTasks(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户ID不能为空",
		})
		return
	}

	// 获取任务类型过滤参数（可选）
	taskType := c.Query("type")

	// 解析分页参数
	limit, offset := h.parsePaginationParams(c)

	tasks, err := h.taskService.GetUserTasks(c.Request.Context(), userID, taskType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取任务列表失败",
			"message": err.Error(),
		})
		return
	}

	responseData := gin.H{
		"tasks":  tasks,
		"limit":  limit,
		"offset": offset,
		"count":  len(tasks),
	}

	// 如果指定了类型过滤，在响应中包含类型信息
	if taskType != "" {
		responseData["type"] = taskType
	} else {
		responseData["type"] = "all"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
	})
}

// 统一的任务结果响应
func (h *AIHandler) respondWithTaskResult(c *gin.Context, task *models.Task) {
	responseData := gin.H{
		"task_id": task.ID,
		"type":    task.Type,
		"status":  task.Status,
		"created": task.Created,
		"updated": task.Updated,
	}

	// 根据任务类型添加特定字段
	switch task.Type {
	case models.TaskTypeImage:
		if task.ImageURL != "" {
			responseData["image_url"] = task.ImageURL
		}
		responseData["size"] = task.Size
	case models.TaskTypeVideo:
		if task.VideoURL != "" {
			responseData["video_url"] = task.VideoURL
		}
		responseData["req_key"] = task.ReqKey
		responseData["seed"] = task.Seed
		responseData["aspect_ratio"] = task.AspectRatio
	case models.TaskTypeText:
		if task.TextResult != "" {
			responseData["text_result"] = task.TextResult
		}
		responseData["max_tokens"] = task.MaxTokens
		responseData["temperature"] = task.Temperature
	}

	switch task.Status {
	case config.TaskStatusCompleted:
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    responseData,
			"message": "任务完成",
		})
	case config.TaskStatusFailed:
		responseData["error"] = task.Error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "任务执行失败",
			"message": task.Error,
			"data":    responseData,
		})
	default:
		c.JSON(http.StatusAccepted, gin.H{
			"success": true,
			"data":    responseData,
			"message": "任务处理中，请稍后查询",
		})
	}
}

// 解析分页参数的辅助方法
func (h *AIHandler) parsePaginationParams(c *gin.Context) (limit, offset int) {
	limit = config.DefaultPageLimit
	offset = config.DefaultPageOffset

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= config.MaxPageLimit {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
