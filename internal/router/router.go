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

		// AI服务 - 火山引擎即梦AI图像生成
		ai := v1.Group("/ai")
		{
			// 火山引擎即梦AI - 异步图像生成
			ai.POST("/image/task", aiHandler.CreateVolcengineImageTask)         // 创建图像生成任务
			ai.GET("/image/result/:task_id", aiHandler.GetVolcengineTaskResult) // 查询任务结果
			ai.GET("/image/tasks", aiHandler.GetUserImageTasks)                 // 获取用户图像任务列表
			ai.DELETE("/image/task/:task_id", aiHandler.DeleteImageTask)        // 删除图像任务
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
