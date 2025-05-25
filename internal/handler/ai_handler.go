package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jimeng-go-server/internal/queue"
	"jimeng-go-server/internal/service"
)

type AIHandler struct {
	imageTaskService    *service.ImageTaskService
	volcengineAIService *service.VolcengineAIService
	queueService        *queue.RedisQueue
}

func NewAIHandler(imageTaskService *service.ImageTaskService, volcengineAIService *service.VolcengineAIService, queueService *queue.RedisQueue) *AIHandler {
	return &AIHandler{
		imageTaskService:    imageTaskService,
		volcengineAIService: volcengineAIService,
		queueService:        queueService,
	}
}

// parseIntParam 解析整数参数
func parseIntParam(s string) (int, error) {
	return strconv.Atoi(s)
}

// 火山引擎即梦AI图像生成请求结构
type VolcengineImageRequest struct {
	Prompt  string `json:"prompt" binding:"required"`
	Model   string `json:"model"`
	N       int    `json:"n"`
	Size    string `json:"size"`
	Quality string `json:"quality"`
	Style   string `json:"style"`
	UserID  string `json:"user_id" binding:"required"`
}

// 创建火山引擎即梦AI异步图像生成任务
func (h *AIHandler) CreateVolcengineImageTask(c *gin.Context) {
	var req VolcengineImageRequest
	if errors := ValidateRequest(c, &req); len(errors) > 0 {
		ResponseValidationError(c, errors)
		return
	}

	// 设置默认模型
	model := req.Model
	if model == "" {
		model = "doubao-seedream-3.0-t2i"
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
		Provider: "volcengine_jimeng",
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
			"provider": "volcengine_jimeng",
		},
		"message": "图像生成任务创建成功",
	})
}

// 查询火山引擎即梦AI任务结果
func (h *AIHandler) GetVolcengineTaskResult(c *gin.Context) {
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
