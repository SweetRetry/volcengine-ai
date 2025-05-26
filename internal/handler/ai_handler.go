package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jimeng-go-server/internal/queue"
	"jimeng-go-server/internal/service"
)

type AIHandler struct {
	imageTaskService *service.ImageTaskService
	queueService     *queue.RedisQueue
}

func NewAIHandler(imageTaskService *service.ImageTaskService, queueService *queue.RedisQueue) *AIHandler {
	return &AIHandler{
		imageTaskService: imageTaskService,
		queueService:     queueService,
	}
}

// parseIntParam 解析整数参数
func parseIntParam(s string) (int, error) {
	return strconv.Atoi(s)
}

// AI图像生成请求结构
type ImageGenerationRequest struct {
	Prompt   string `json:"prompt" binding:"required"`
	Model    string `json:"model"`
	N        int    `json:"n"`
	Size     string `json:"size"`
	Quality  string `json:"quality"`
	Style    string `json:"style"`
	UserID   string `json:"user_id" binding:"required"`
	Provider string `json:"provider"` // AI服务提供商：volcengine_jimeng, openai, etc.
}

// AI文本生成请求结构
type TextGenerationRequest struct {
	Prompt      string  `json:"prompt" binding:"required"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	UserID      string  `json:"user_id" binding:"required"`
	Provider    string  `json:"provider"` // AI服务提供商：volcengine_jimeng, openai, etc.
}

// AI视频生成请求结构
type VideoGenerationRequest struct {
	Prompt   string `json:"prompt" binding:"required"`
	Model    string `json:"model"`
	Duration int    `json:"duration"` // 视频时长（秒）
	Quality  string `json:"quality"`
	Style    string `json:"style"`
	UserID   string `json:"user_id" binding:"required"`
	Provider string `json:"provider"` // AI服务提供商：volcengine_jimeng, openai, etc.
}

// 创建AI图像生成任务
func (h *AIHandler) CreateImageTask(c *gin.Context) {
	var req ImageGenerationRequest
	if errors := ValidateRequest(c, &req); len(errors) > 0 {
		ResponseValidationError(c, errors)
		return
	}

	// 设置默认提供商和模型
	provider := req.Provider
	if provider == "" {
		provider = "volcengine_jimeng" // 默认使用火山引擎
	}

	model := req.Model
	if model == "" {
		// 根据提供商设置默认模型
		switch provider {
		case "volcengine_jimeng":
			model = "doubao-seedream-3.0-t2i"
		case "openai":
			model = "dall-e-3"
		default:
			model = "doubao-seedream-3.0-t2i"
		}
	}

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
	task, err := h.imageTaskService.CreateImageTask(c.Request.Context(), input)
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
		Type:     "image_generation",
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
	if err := h.queueService.EnqueueTask(c.Request.Context(), queue.TypeImageGeneration, payload); err != nil {
		// 如果入队失败，删除已创建的任务记录
		h.imageTaskService.DeleteImageTask(c.Request.Context(), task.ID)
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
			"status":   "pending",
			"provider": provider,
			"model":    model,
		},
		"message": "图像生成任务创建成功",
	})
}

// 查询AI图像生成任务结果
func (h *AIHandler) GetImageTaskResult(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务ID不能为空",
		})
		return
	}

	// 获取图像任务详情
	result, err := h.imageTaskService.GetImageTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "任务不存在",
			"message": err.Error(),
		})
		return
	}

	// 如果任务已经完成或失败，直接返回结果
	if result.Status == "completed" || result.Status == "failed" {
		if result.Status == "failed" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "任务执行失败",
				"message": result.Error,
				"data": gin.H{
					"task_id": taskID,
					"status":  "failed",
					"created": result.Created,
				},
			})
			return
		}

		// 任务完成
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"task_id":   taskID,
				"status":    "completed",
				"image_url": result.ImageURL,
				"created":   result.Created,
			},
			"message": "任务完成",
		})
		return
	}

	// 如果任务还在处理中，返回处理中状态
	// 现在使用Redis队列处理，任务状态由队列工作器更新
	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"data": gin.H{
			"task_id": taskID,
			"status":  "processing",
			"message": "任务处理中，请稍后查询",
			"created": result.Created,
		},
	})
}

// 获取用户的图像任务列表
func (h *AIHandler) GetUserImageTasks(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户ID不能为空",
		})
		return
	}

	// 分页参数
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := parseIntParam(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := parseIntParam(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	tasks, err := h.imageTaskService.GetUserImageTasks(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取任务列表失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks":  tasks,
			"limit":  limit,
			"offset": offset,
			"count":  len(tasks),
		},
	})
}

// 删除图像任务
func (h *AIHandler) DeleteImageTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务ID不能为空",
		})
		return
	}

	err := h.imageTaskService.DeleteImageTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除任务失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务删除成功",
	})
}

// 创建AI文本生成任务
func (h *AIHandler) CreateTextTask(c *gin.Context) {
	var req TextGenerationRequest
	if errors := ValidateRequest(c, &req); len(errors) > 0 {
		ResponseValidationError(c, errors)
		return
	}

	// TODO: 实现文本生成任务创建逻辑
	// 设置默认提供商和模型
	provider := req.Provider
	if provider == "" {
		provider = "volcengine_jimeng" // 默认使用火山引擎
	}

	model := req.Model
	if model == "" {
		// 根据提供商设置默认模型
		switch provider {
		case "volcengine_jimeng":
			model = "doubao-pro-4k"
		case "openai":
			model = "gpt-4"
		default:
			model = "doubao-pro-4k"
		}
	}

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

// 创建AI视频生成任务
func (h *AIHandler) CreateVideoTask(c *gin.Context) {
	var req VideoGenerationRequest
	if errors := ValidateRequest(c, &req); len(errors) > 0 {
		ResponseValidationError(c, errors)
		return
	}

	// TODO: 实现视频生成任务创建逻辑
	// 设置默认提供商和模型
	provider := req.Provider
	if provider == "" {
		provider = "volcengine_jimeng" // 默认使用火山引擎
	}

	model := req.Model
	if model == "" {
		// 根据提供商设置默认模型
		switch provider {
		case "volcengine_jimeng":
			model = "doubao-video-pro"
		case "openai":
			model = "sora"
		default:
			model = "doubao-video-pro"
		}
	}

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
