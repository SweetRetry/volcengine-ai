package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jimeng-go-server/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
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

// 获取用户任务列表
func (h *TaskHandler) GetUserTasks(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户ID不能为空",
		})
		return
	}

	// 获取分页参数
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	tasks, err := h.taskService.GetUserTasks(c.Request.Context(), userID, limit, offset)
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

// 取消任务
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务ID不能为空",
		})
		return
	}

	err := h.taskService.CancelTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "取消任务失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务已取消",
	})
}

// 重试任务
func (h *TaskHandler) RetryTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务ID不能为空",
		})
		return
	}

	err := h.taskService.RetryTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "重试任务失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务已重新提交",
	})
}

// 删除任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "任务ID不能为空",
		})
		return
	}

	err := h.taskService.DeleteTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "删除任务失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务已删除",
	})
}

// 获取队列统计信息
func (h *TaskHandler) GetQueueStats(c *gin.Context) {
	stats, err := h.taskService.GetQueueStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取队列统计失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
