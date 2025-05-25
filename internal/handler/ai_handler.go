package handler

import (
	"net/http"
	"strconv"

	"jimeng-go-server/internal/service"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	imageTaskService    *service.ImageTaskService
	volcengineAIService *service.VolcengineAIService
}

func NewAIHandler(imageTaskService *service.ImageTaskService, volcengineAIService *service.VolcengineAIService) *AIHandler {
	return &AIHandler{
		imageTaskService:    imageTaskService,
		volcengineAIService: volcengineAIService,
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
			"error":   "创建图像任务失败",
			"message": err.Error(),
		})
		return
	}

	// 构建火山引擎API选项
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

	// 调用火山引擎API创建异步任务
	taskResponse, err := h.volcengineAIService.CreateImageTask(c.Request.Context(), req.Prompt, options)
	if err != nil {
		// 如果火山引擎API调用失败，更新任务状态
		h.imageTaskService.UpdateImageTaskStatus(c.Request.Context(), task.ID, "failed", "", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建图像生成任务失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":          task.ID,
			"status":           "pending",
			"provider":         "volcengine_jimeng",
			"external_task_id": taskResponse.TaskID,
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

	// 如果任务还在处理中
	if result.Status == "pending" || result.Status == "processing" {
		c.JSON(http.StatusAccepted, gin.H{
			"success": true,
			"data": gin.H{
				"task_id": taskID,
				"status":  result.Status,
				"message": "任务处理中，请稍后查询",
				"created": result.Created,
			},
		})
		return
	}

	// 如果任务失败
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
