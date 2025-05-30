package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"volcengine-go-server/api/handlers"
)

func SetupRoutes(
	r *gin.Engine,
	aiHandler *handlers.AIHandler,
	userHandler *handlers.UserHandler,
) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "服务正常运行",
		})
	})

	// API版本分组
	v1 := r.Group("/api/v1")
	{
		// 用户管理
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.GET("", userHandler.GetUserByEmail) // ?email=xxx
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// AI服务
		ai := v1.Group("/ai")
		{
			// AI任务创建 - 类型特定接口
			ai.POST("/image/task", aiHandler.CreateImageTask) // 创建图像生成任务
			ai.POST("/text/task", aiHandler.CreateTextTask)   // 创建文本生成任务 (TODO: 待实现)
			ai.POST("/video/task", aiHandler.CreateVideoTask) // 创建视频生成任务

			// 统一任务管理 - 通用接口
			ai.GET("/task/result/:task_id", aiHandler.GetTaskResult) // 查询任务结果（通用）
			ai.DELETE("/task/:task_id", aiHandler.DeleteTask)        // 删除任务（通用）
			ai.GET("/tasks", aiHandler.GetUserTasks)                 // 获取用户任务列表（通用，支持类型过滤）
		}
	}

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "接口不存在",
			"path":  c.Request.URL.Path,
		})
	})
}
