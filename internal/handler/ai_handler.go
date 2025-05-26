package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jimeng-go-server/internal/queue"
	"jimeng-go-server/internal/service"
)

// 重构后的AI Handler - 更简洁
type AIHandler struct {
	taskFactory      *AITaskFactory
	imageTaskService *service.ImageTaskService
}

func NewAIHandler(imageTaskService *service.ImageTaskService, queueService *queue.RedisQueue) *AIHandler {
	return &AIHandler{
		taskFactory:      NewAITaskFactory(imageTaskService, queueService),
		imageTaskService: imageTaskService,
	}
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

// 创建图像任务 - 使用工厂模式
func (h *AIHandler) CreateImageTask(c *gin.Context) {
	h.taskFactory.CreateTask(c, TaskTypeImage)
}

// 创建文本任务 - 使用工厂模式
func (h *AIHandler) CreateTextTask(c *gin.Context) {
	h.taskFactory.CreateTask(c, TaskTypeText)
}

// 创建视频任务 - 使用工厂模式
func (h *AIHandler) CreateVideoTask(c *gin.Context) {
	h.taskFactory.CreateTask(c, TaskTypeVideo)
}

// 通用任务结果查询 - 可以查询任何类型的任务
func (h *AIHandler) GetImageTaskResult(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务ID不能为空",
		})
		return
	}

	// 目前只支持图像任务，未来可以扩展
	result, err := h.imageTaskService.GetImageTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "任务不存在",
			"message": err.Error(),
		})
		return
	}

	h.respondWithTaskResult(c, taskID, result)
}

// 统一的任务结果响应
func (h *AIHandler) respondWithTaskResult(c *gin.Context, taskID string, result *service.ImageTaskResult) {
	switch result.Status {
	case "completed":
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
	case "failed":
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "任务执行失败",
			"message": result.Error,
			"data": gin.H{
				"task_id": taskID,
				"status":  "failed",
				"created": result.Created,
			},
		})
	default:
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
}

// 获取用户任务列表 - 支持分页
func (h *AIHandler) GetUserImageTasks(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户ID不能为空",
		})
		return
	}

	// 解析分页参数
	limit, offset := h.parsePaginationParams(c)

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

// 删除任务
func (h *AIHandler) DeleteImageTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务ID不能为空",
		})
		return
	}

	if err := h.imageTaskService.DeleteImageTask(c.Request.Context(), taskID); err != nil {
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

// 解析分页参数的辅助方法
func (h *AIHandler) parsePaginationParams(c *gin.Context) (limit, offset int) {
	limit = 20 // 默认值
	offset = 0 // 默认值

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
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
