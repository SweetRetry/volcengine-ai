package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"jimeng-go-server/internal/config"
	"jimeng-go-server/internal/database"
	"jimeng-go-server/internal/handler"
	"jimeng-go-server/internal/middleware"
	"jimeng-go-server/internal/queue"
	"jimeng-go-server/internal/router"
	"jimeng-go-server/internal/service"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		logrus.Warn("没有找到.env文件")
	}

	// 初始化配置
	cfg := config.New()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		logrus.Fatal("配置验证失败: ", err)
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(level)
	}

	// 初始化MongoDB数据库
	db, err := database.NewMongoDB(cfg.Database.MongoURL)
	if err != nil {
		logrus.Fatal("连接MongoDB失败: ", err)
	}
	defer db.Close()

	// 初始化服务
	volcengineAIService := service.NewVolcengineAIService(cfg.AI)
	userService := service.NewUserService(db)
	imageTaskService := service.NewImageTaskService(db)

	// 创建服务注册器
	serviceRegistry := queue.NewServiceRegistry()

	// 创建并注册火山引擎AI服务提供商
	volcengineProvider := service.NewVolcengineAIProvider(volcengineAIService, imageTaskService)
	serviceRegistry.RegisterProvider(volcengineProvider)

	// 创建并注册OpenAI服务提供商（示例）
	// openaiProvider := service.NewOpenAIProvider(cfg.OpenAI.APIKey)
	// serviceRegistry.RegisterProvider(openaiProvider)

	// 初始化队列（使用服务注册器）
	queueClient := queue.NewRedisQueue(cfg.Redis.URL, imageTaskService, serviceRegistry)

	// 初始化处理器
	aiHandler := handler.NewAIHandler(imageTaskService, queueClient)
	userHandler := handler.NewUserHandler(userService)

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由器
	r := gin.New()

	// 注册中间件
	r.Use(middleware.Logger())
	r.Use(middleware.OptionsHandler())
	r.Use(middleware.Recovery())
	r.Use(middleware.RateLimiterMiddleware())

	// 设置路由
	router.SetupRoutes(r, aiHandler, userHandler)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// 启动队列工作器
	go queueClient.StartWorker(context.Background())

	// 启动服务器
	go func() {
		logrus.Infof("服务器启动在端口 %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("启动服务器失败: ", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("服务器强制关闭: ", err)
	}

	logrus.Info("服务器已退出")
}
