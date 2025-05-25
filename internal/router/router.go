package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jimeng-go-server/internal/handler"
)

func SetupRoutes(
	r *gin.Engine,
	aiHandler *handler.AIHandler,
	userHandler *handler.UserHandler,
	taskHandler *handler.TaskHandler,
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

		// AI服务 - 纯异步模式
		ai := v1.Group("/ai")
		{
			// 通用异步任务
			ai.POST("/tasks", aiHandler.CreateAsyncTask)
			ai.GET("/tasks/:id", aiHandler.GetTaskStatus)

			// 火山引擎即梦AI - 异步图像生成
			ai.POST("/image/task", aiHandler.CreateVolcengineImageTask)         // 创建图像生成任务
			ai.GET("/image/result/:task_id", aiHandler.GetVolcengineTaskResult) // 查询任务结果
		}

		// 任务管理
		tasks := v1.Group("/tasks")
		{
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.GET("/user/:user_id", taskHandler.GetUserTasks)
			tasks.POST("/:id/cancel", taskHandler.CancelTask)
			tasks.POST("/:id/retry", taskHandler.RetryTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
		}

		// 队列管理
		queue := v1.Group("/queue")
		{
			queue.GET("/stats", taskHandler.GetQueueStats)
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
