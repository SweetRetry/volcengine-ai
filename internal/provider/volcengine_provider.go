package provider

import (
	"context"
	"fmt"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/service"
	"volcengine-go-server/pkg/logger"
)

// VolcengineProvider 火山引擎任务分发器 - Provider层
// 只负责根据模型参数决定调用VolcengineService的哪个具体方法
type VolcengineProvider struct {
	volcengineService *service.VolcengineService // 依赖具体的Service实现
	taskService       *service.TaskService
}

// NewVolcengineProvider 创建火山引擎任务分发器
func NewVolcengineProvider(volcengineService *service.VolcengineService, taskService *service.TaskService) *VolcengineProvider {
	return &VolcengineProvider{
		volcengineService: volcengineService,
		taskService:       taskService,
	}
}

// GetProviderName 获取分发器名称
func (p *VolcengineProvider) GetProviderName() string {
	return "volcengine"
}

// DispatchImageTask 分发图像生成任务
func (p *VolcengineProvider) DispatchImageTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	log := logger.GetLogger()
	log.Infof("火山引擎图像任务分发: taskID=%s, model=%s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.VolcengineJimengImageModel:
		log.Infof("分发到即梦AI图像生成服务: %s", taskID)
		return p.volcengineService.GenerateImageByJimeng(ctx, taskID, input)

	case config.VolcengineImageModel: // doubao-seedream-3-0-t2i-250415
		log.Infof("分发到豆包图像生成服务: %s", taskID)
		return p.volcengineService.GenerateImageByDoubao(ctx, taskID, input)

	default:
		// 默认使用豆包模型
		log.Warnf("未知模型 %s，使用默认豆包模型", model)
		return p.volcengineService.GenerateImageByDoubao(ctx, taskID, input)
	}
}

// DispatchTextTask 分发文本生成任务
func (p *VolcengineProvider) DispatchTextTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	log := logger.GetLogger()
	log.Infof("火山引擎文本任务分发: taskID=%s, model=%s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.VolcengineTextModel:
		return p.volcengineService.GenerateTextByDoubao(ctx, taskID, input)
	default:
		return fmt.Errorf("不支持的文本生成模型: %s", model)
	}
}

// DispatchVideoTask 分发视频生成任务
func (p *VolcengineProvider) DispatchVideoTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	log := logger.GetLogger()
	log.Infof("火山引擎视频任务分发: taskID=%s, model=%s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.VolcengineJimengVideoModel:
		return p.volcengineService.GenerateVideoByJimeng(ctx, taskID, input)
	default:
		return fmt.Errorf("不支持的视频生成模型: %s", model)
	}
}
