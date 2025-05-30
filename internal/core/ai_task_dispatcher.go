package core

import (
	"context"
)

// AITaskDispatcher AI任务分发器接口 - Provider层职责
// Provider只负责根据参数决定调用哪个具体的Service，不包含具体实现
type AITaskDispatcher interface {
	// 获取分发器名称
	GetProviderName() string

	// 分发图像生成任务到具体的Service
	DispatchImageTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error

	// 分发文本生成任务到具体的Service
	DispatchTextTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error

	// 分发视频生成任务到具体的Service
	DispatchVideoTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error
}

// AIImageService AI图像生成服务接口 - Service层职责
// Service负责具体的API调用和业务逻辑实现
type AIImageService interface {
	GenerateImage(ctx context.Context, taskID string, input map[string]interface{}) error
}

// AITextService AI文本生成服务接口
type AITextService interface {
	GenerateText(ctx context.Context, taskID string, input map[string]interface{}) error
}

// AIVideoService AI视频生成服务接口
type AIVideoService interface {
	GenerateVideo(ctx context.Context, taskID string, input map[string]interface{}) error
}
