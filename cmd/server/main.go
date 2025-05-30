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

	"volcengine-go-server/api/handlers"
	"volcengine-go-server/api/middleware"
	"volcengine-go-server/api/routes"
	"volcengine-go-server/config"
	"volcengine-go-server/internal/core"
	"volcengine-go-server/internal/repository"
	"volcengine-go-server/internal/service"
	"volcengine-go-server/pkg/logger"
)

func main() {
	// 初始化日志系统（需要在最早期初始化）
	logger.Init()
	log := logger.GetLogger()

	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Warn("没有找到.env文件")
	}

	// 初始化配置
	cfg := config.New()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatal("配置验证失败: ", err)
	}

	// 创建日志管理器
	logManager := logger.NewLogManager()

	// 可以根据环境变量配置日志保留天数
	if keepDaysEnv := os.Getenv("LOG_KEEP_DAYS"); keepDaysEnv != "" {
		if keepDays, err := time.ParseDuration(keepDaysEnv + "h"); err == nil {
			logManager.SetKeepDays(int(keepDays.Hours() / 24))
		}
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.SetLevel("info")
	} else {
		logger.SetLevel(level.String())
	}

	// 初始化MongoDB数据库
	db, err := repository.NewMongoDB(cfg.Database.MongoURL)
	if err != nil {
		log.Fatal("连接MongoDB失败: ", err)
	}
	defer db.Close()

	// 初始化基础服务（API服务器只需要这些）
	userService := service.NewUserService(db)
	taskService := service.NewTaskService(db)

	// 创建空的服务注册器（API服务器不需要注册任何提供商）
	serviceRegistry := core.NewServiceRegistry()
	// 注意：API服务器不注册任何AI服务提供商，因为它不处理任务

	// 初始化队列客户端（只用于发送任务到队列）
	queueClient := core.NewTaskQueue(cfg.Redis.URL, taskService, serviceRegistry)

	// 初始化处理器
	aiHandler := handlers.NewAIHandler(taskService, queueClient)
	userHandler := handlers.NewUserHandler(userService)

	// 设置Gin模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由器
	r := gin.New()

	// 注册中间件
	r.Use(middleware.Logger())

	// 根据环境变量决定是否启用详细日志
	if os.Getenv("ENABLE_DETAILED_LOGGING") == "true" {
		log.Info("启用详细HTTP请求日志记录")
		r.Use(middleware.DetailedLogger())
	}

	r.Use(middleware.OptionsHandler())
	r.Use(middleware.Recovery())
	r.Use(middleware.RateLimiterMiddleware())

	// 设置路由
	routes.SetupRoutes(r, aiHandler, userHandler)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动日志管理器
	go logManager.Start(ctx)

	// 启动服务器
	go func() {
		log.Infof("API服务器启动在端口 %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("启动服务器失败: ", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("正在关闭服务器...")

	// 取消上下文，停止日志管理器
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("服务器强制关闭: ", err)
	}

	log.Info("服务器已退出")
}
