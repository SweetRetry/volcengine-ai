package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/core"
	"volcengine-go-server/internal/provider"
	"volcengine-go-server/internal/repository"
	"volcengine-go-server/internal/service"
	"volcengine-go-server/pkg/logger"
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

	// 初始化日志系统
	logger.Init()

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
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(level)
	}

	// 初始化MongoDB数据库
	db, err := repository.NewMongoDB(cfg.Database.MongoURL)
	if err != nil {
		logrus.Fatal("连接MongoDB失败: ", err)
	}
	defer db.Close()

	// 初始化服务层
	taskService := service.NewTaskService(db)
	volcengineService := service.NewVolcengineService(cfg.AI, taskService)

	// 创建OpenAI服务（示例，如果需要的话）
	// openaiService := service.NewOpenAIService("your-openai-api-key", taskService)

	// 创建服务注册器
	serviceRegistry := core.NewServiceRegistry()

	// 创建并注册火山引擎任务分发器
	volcengineProvider := provider.NewVolcengineProvider(volcengineService, taskService)
	serviceRegistry.RegisterDispatcher(volcengineProvider)

	// 创建并注册OpenAI任务分发器（示例）
	// openaiProvider := provider.NewOpenAIProvider(openaiService, taskService)
	// serviceRegistry.RegisterDispatcher(openaiProvider)

	logrus.Infof("已注册AI服务分发器: %v", getRegisteredProviders(serviceRegistry))

	// 初始化队列（使用服务注册器）
	queueClient := core.NewTaskQueue(cfg.Redis.URL, taskService, serviceRegistry)

	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动日志管理器
	go logManager.Start(ctx)

	logrus.Info("任务处理中心启动中...")

	// 启动队列工作器
	go queueClient.StartWorker(ctx)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("正在关闭任务处理中心...")

	// 取消上下文，停止日志管理器和队列工作器
	cancel()

	// 等待一段时间让工作器完成当前任务
	time.Sleep(2 * time.Second)

	// 关闭队列客户端
	if err := queueClient.Close(); err != nil {
		logrus.Errorf("关闭队列客户端失败: %v", err)
	}

	logrus.Info("任务处理中心已退出")
}

// getRegisteredProviders 获取已注册的服务提供商列表
func getRegisteredProviders(registry *core.ServiceRegistry) []string {
	dispatchers := registry.GetAllDispatchers()
	providers := make([]string, 0, len(dispatchers))
	for name := range dispatchers {
		providers = append(providers, name)
	}
	return providers
}
