package volcengine

import (
	"context"
	"fmt"

	"volcengine-go-server/config"
	"volcengine-go-server/pkg/logger"
)

// Provider 火山引擎任务分发器 - Provider层
// 只负责根据模型参数决定调用VolcengineService的哪个具体方法
type Provider struct {
	service     *VolcengineService
	taskService TaskService
}

// NewProvider 创建火山引擎任务分发器
func NewProvider(service *VolcengineService, taskService TaskService) *Provider {
	return &Provider{
		service:     service,
		taskService: taskService,
	}
}

// GetProviderName 获取分发器名称
func (p *Provider) GetProviderName() string {
	return "volcengine"
}

// DispatchImageTask 分发图像生成任务
func (p *Provider) DispatchImageTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	log := logger.GetLogger()
	log.Infof("火山引擎图像任务分发: taskID=%s, model=%s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.VolcengineJimengImageModel:
		log.Infof("分发到即梦AI图像生成服务: %s", taskID)
		return p.service.GenerateImageByJimeng(ctx, taskID, input)

	case config.VolcengineImageModel: // doubao-seedream-3-0-t2i-250415
		log.Infof("分发到豆包图像生成服务: %s", taskID)
		return p.service.GenerateImageByDoubao(ctx, taskID, input)

	default:
		// 默认使用豆包模型
		log.Warnf("未知模型 %s，使用默认豆包模型", model)
		return p.service.GenerateImageByDoubao(ctx, taskID, input)
	}
}

// DispatchTextTask 分发文本生成任务
func (p *Provider) DispatchTextTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	log := logger.GetLogger()
	log.Infof("火山引擎文本任务分发: taskID=%s, model=%s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.VolcengineTextModel:
		return p.service.GenerateTextByDoubao(ctx, taskID, input)
	default:
		return fmt.Errorf("不支持的文本生成模型: %s", model)
	}
}

// DispatchVideoTask 分发视频生成任务
func (p *Provider) DispatchVideoTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	log := logger.GetLogger()
	log.Infof("火山引擎视频任务分发: taskID=%s, model=%s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.VolcengineJimengVideoModel:
		return p.service.GenerateVideoByJimeng(ctx, taskID, input)
	case config.VolcengineJimengI2VModel:
		return p.service.GenerateI2VByJimeng(ctx, taskID, input)
	default:
		return fmt.Errorf("不支持的视频生成模型: %s", model)
	}
}
