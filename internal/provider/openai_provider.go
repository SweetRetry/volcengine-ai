package provider

import (
	"context"
	"fmt"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/service"
	"volcengine-go-server/pkg/logger"
)

// OpenAIProvider OpenAI任务分发器 - Provider层
// 只负责根据模型参数决定调用OpenAIService的哪个具体方法
type OpenAIProvider struct {
	openaiService *service.OpenAIService // 依赖具体的Service实现
	taskService   *service.TaskService
}

// NewOpenAIProvider 创建OpenAI任务分发器
func NewOpenAIProvider(openaiService *service.OpenAIService, taskService *service.TaskService) *OpenAIProvider {
	return &OpenAIProvider{
		openaiService: openaiService,
		taskService:   taskService,
	}
}

// GetProviderName 获取分发器名称
func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

// DispatchImageTask 分发图像生成任务
func (p *OpenAIProvider) DispatchImageTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	log := logger.GetLogger()
	log.Infof("OpenAI图像任务分发: taskID=%s, model=%s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.OpenAIImageModel: // dall-e-3
		log.Infof("分发到DALL-E图像生成服务: %s", taskID)
		return p.openaiService.GenerateImageByDALLE(ctx, taskID, input)

	default:
		return fmt.Errorf("不支持的OpenAI图像生成模型: %s", model)
	}
}

// DispatchTextTask 分发文本生成任务
func (p *OpenAIProvider) DispatchTextTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	log := logger.GetLogger()
	log.Infof("OpenAI文本任务分发: taskID=%s, model=%s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.OpenAITextModel: // gpt-4
		return p.openaiService.GenerateTextByGPT(ctx, taskID, input)
	default:
		return fmt.Errorf("不支持的OpenAI文本生成模型: %s", model)
	}
}

// DispatchVideoTask 分发视频生成任务
func (p *OpenAIProvider) DispatchVideoTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	log := logger.GetLogger()
	log.Infof("OpenAI视频任务分发: taskID=%s, model=%s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.OpenAIVideoModel: // sora
		return p.openaiService.GenerateVideoBySora(ctx, taskID, input)
	default:
		return fmt.Errorf("不支持的OpenAI视频生成模型: %s", model)
	}
}
