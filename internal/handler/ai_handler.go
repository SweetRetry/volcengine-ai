package handler

import (
	"net/http"
	"time"

	"jimeng-go-server/internal/service"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	taskService         *service.TaskService
	volcengineAIService *service.VolcengineAIService
}

func NewAIHandler(taskService *service.TaskService, volcengineAIService *service.VolcengineAIService) *AIHandler {
	return &AIHandler{
		taskService:         taskService,
		volcengineAIService: volcengineAIService,
	}
}

// 异步任务请求结构
type AsyncTaskRequest struct {
	Type     string                 `json:"type" binding:"required"`
	Model    string                 `json:"model" binding:"required"`
	Input    map[string]interface{} `json:"input" binding:"required"`
	Provider string                 `json:"provider"`
	UserID   string                 `json:"user_id" binding:"required"`
	Delay    int                    `json:"delay"`
}

// 火山引擎即梦AI图像生成请求结构
type VolcengineImageRequest struct {
	Prompt  string                 `json:"prompt" binding:"required"`
	Model   string                 `json:"model"`
	N       int                    `json:"n"`
	Size    string                 `json:"size"`
	Quality string                 `json:"quality"`
	Style   string                 `json:"style"`
	UserID  string                 `json:"user_id" binding:"required"`
	Options map[string]interface{} `json:"options"`
}

// 创建异步任务
func (h *AIHandler) CreateAsyncTask(c *gin.Context) {
	var req AsyncTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	provider := req.Provider
	if provider == "" {
		provider = "volcengine"
	}

	var task interface{}
	var err error

	if req.Delay > 0 {
		delay := time.Duration(req.Delay) * time.Second
		task, err = h.taskService.CreateDelayedTask(
			c.Request.Context(),
			req.UserID,
			req.Type,
			req.Input,
			req.Model,
			provider,
			delay,
		)
	} else {
		task, err = h.taskService.CreateAITask(
			c.Request.Context(),
			req.UserID,
			req.Type,
			req.Input,
			req.Model,
			provider,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建任务失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    task,
		"message": "任务创建成功",
	})
}

// 获取任务状态
func (h *AIHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

// 创建火山引擎即梦AI异步图像生成任务
func (h *AIHandler) CreateVolcengineImageTask(c *gin.Context) {
	var req VolcengineImageRequest
	if errors := ValidateRequest(c, &req); len(errors) > 0 {
		ResponseValidationError(c, errors)
		return
	}

	// 构建选项参数
	options := make(map[string]interface{})
	if req.Model != "" {
		options["model"] = req.Model
	}
	if req.N > 0 {
		options["n"] = req.N
	}
	if req.Size != "" {
		options["size"] = req.Size
	}
	if req.Quality != "" {
		options["quality"] = req.Quality
	}
	if req.Style != "" {
		options["style"] = req.Style
	}

	// 合并自定义选项
	for k, v := range req.Options {
		options[k] = v
	}

	// 创建异步任务
	taskResponse, err := h.volcengineAIService.CreateImageTask(c.Request.Context(), req.Prompt, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建图像生成任务失败",
			"message": err.Error(),
		})
		return
	}

	// 同时在任务系统中创建记录
	input := map[string]interface{}{
		"prompt":  req.Prompt,
		"options": options,
	}

	model := req.Model
	if model == "" {
		model = "doubao-seedream-3.0-t2i"
	}

	task, err := h.taskService.CreateAITask(
		c.Request.Context(),
		req.UserID,
		"image_generation",
		input,
		model,
		"volcengine_jimeng",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建任务记录失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":        taskResponse.TaskID,
			"status":         taskResponse.Status,
			"message":        taskResponse.Message,
			"provider":       taskResponse.Provider,
			"system_task_id": task,
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

	result, err := h.volcengineAIService.GetTaskResult(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询任务结果失败",
			"message": err.Error(),
		})
		return
	}

	// 如果任务还在处理中
	if result.Error == "任务处理中，请稍后查询" {
		c.JSON(http.StatusAccepted, gin.H{
			"success": true,
			"data": gin.H{
				"task_id": taskID,
				"status":  "processing",
				"message": "任务处理中，请稍后查询",
			},
		})
		return
	}

	// 如果有错误
	if result.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "任务执行失败",
			"message": result.Error,
			"data": gin.H{
				"task_id": taskID,
				"status":  "failed",
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
			"result":    result,
			"image_url": result.ImageURL,
		},
		"message": "任务完成",
	})
}
