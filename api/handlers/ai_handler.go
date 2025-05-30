package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/models"
	"volcengine-go-server/internal/queue"
	"volcengine-go-server/internal/service"
)

type AIHandler struct {
	taskFactory *AITaskFactory
	taskService *service.TaskService
}

func NewAIHandler(taskService *service.TaskService, queueService *queue.RedisQueue) *AIHandler {
	return &AIHandler{
		taskFactory: NewAITaskFactory(taskService, queueService),
		taskService: taskService,
	}
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
