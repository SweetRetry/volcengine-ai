package volcengine

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/volcengine/volc-sdk-golang/service/visual"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"

	"volcengine-go-server/config"
	"volcengine-go-server/pkg/logger"
)

// TaskService 接口定义，避免循环依赖
type TaskService interface {
	UpdateTaskError(ctx context.Context, taskID string, errorMsg string) error
	UpdateTaskResult(ctx context.Context, taskID string, result string) error
}

// VolcengineService 火山引擎AI服务 - Service层，负责具体的API调用实现
type VolcengineService struct {
	config       config.AIConfig
	client       *arkruntime.Client
	logger       *logrus.Logger
	visualClient *visual.Visual
	taskService  TaskService
}

// NewVolcengineService 创建火山引擎AI服务实例
func NewVolcengineService(cfg config.AIConfig, taskService TaskService) *VolcengineService {
	// 设置API Key到环境变量
	if cfg.VolcengineAPIKey != "" {
		os.Setenv("ARK_API_KEY", cfg.VolcengineAPIKey)
	}

	// 创建火山方舟客户端
	client := arkruntime.NewClientWithApiKey(
		os.Getenv("ARK_API_KEY"),
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)

	visualClient := visual.NewInstance()
	visualClient.Client.SetAccessKey(cfg.VolcengineAccessKey)
	visualClient.Client.SetSecretKey(cfg.VolcengineSecretKey)

	return &VolcengineService{
		config:       cfg,
		client:       client,
		visualClient: visualClient,
		logger:       logger.GetLogger(), // 使用全局日志记录器
		taskService:  taskService,
	}
}

// HealthCheck 健康检查
func (s *VolcengineService) HealthCheck(ctx context.Context) error {
	// 简单的健康检查
	return nil
}
