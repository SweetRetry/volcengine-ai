package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"jimeng-go-server/internal/config"
)

// VolcengineAIProvider 火山引擎AI服务提供商适配器
type VolcengineAIProvider struct {
	volcengineAIService *VolcengineAIService
	imageTaskService    *ImageTaskService
}

// NewVolcengineAIProvider 创建火山引擎AI服务提供商
func NewVolcengineAIProvider(
	volcengineAIService *VolcengineAIService,
	imageTaskService *ImageTaskService,
) *VolcengineAIProvider {
	return &VolcengineAIProvider{
		volcengineAIService: volcengineAIService,
		imageTaskService:    imageTaskService,
	}
}

// GetProviderName 获取提供商名称
func (v *VolcengineAIProvider) GetProviderName() string {
	return "volcengine"
}

// ProcessImageTask 处理图像生成任务
func (v *VolcengineAIProvider) ProcessImageTask(ctx context.Context, taskID string, input map[string]interface{}) error {
	// 获取任务输入参数
	taskInput, err := v.imageTaskService.GetImageTaskInput(ctx, taskID)
	if err != nil {
		logrus.Errorf("获取任务输入失败: %v", err)
		v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", err.Error())
		return err
	}

	// 构建图像生成请求参数
	request := &VolcengineImageRequest{
		Prompt: taskInput.Prompt,
		Model:  config.VolcengineImageModel,
		Size:   v.parseOptimalSizeString(taskInput.Size),
		N:      1, // 生成1张图片
	}

	// 直接生成图像（火山方舟是同步API）
	result, err := v.volcengineAIService.GenerateImage(ctx, request)
	if err != nil {
		logrus.Errorf("火山方舟图像生成失败: %v", err)
		v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", err.Error())
		return err
	}

	logrus.Infof("火山方舟图像生成成功: %s (尺寸: %s)", taskID, request.Size)

	// 检查是否有生成的图像
	if len(result.Data) == 0 {
		errorMsg := "未生成任何图像"
		logrus.Errorf("图像生成失败: %s", errorMsg)
		v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", errorMsg)
		return fmt.Errorf(errorMsg)
	}

	// 获取第一张图片的URL
	imageURL := result.Data[0].URL
	logrus.Infof("图像生成任务完成: %s, 图像URL: %s", taskID, imageURL)

	// 更新数据库中的任务状态
	if err := v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "completed", imageURL, ""); err != nil {
		logrus.Errorf("更新任务状态失败: %v", err)
		return err
	}

	logrus.Infof("任务状态已更新为完成: %s", taskID)
	return nil
}

// parseOptimalSizeString 解析并返回最优的图像尺寸字符串
func (v *VolcengineAIProvider) parseOptimalSizeString(size string) string {
	// 火山方舟支持的尺寸格式
	switch size {
	case "1:1", "1024x1024", "":
		return config.ImageSize1x1
	case "3:4", "864x1152":
		return config.ImageSize3x4
	case "4:3", "1152x864":
		return config.ImageSize4x3
	case "16:9", "1280x720":
		return config.ImageSize16x9
	case "9:16", "720x1280":
		return config.ImageSize9x16
	case "2:3", "832x1248":
		return config.ImageSize2x3
	case "3:2", "1248x832":
		return config.ImageSize3x2
	case "21:9", "1512x648":
		return config.ImageSize21x9
	default:
		// 默认使用1:1比例
		return config.DefaultImageSize
	}
}

// ProcessTextTask 处理文本生成任务
func (v *VolcengineAIProvider) ProcessTextTask(ctx context.Context, taskID string, input map[string]interface{}) error {
	// TODO: 实现火山引擎文本生成逻辑
	logrus.Infof("火山引擎文本生成任务处理中: %s", taskID)

	// 模拟处理时间
	time.Sleep(2 * time.Second)

	logrus.Infof("火山引擎文本生成任务完成: %s", taskID)
	return nil
}

// ProcessVideoTask 处理视频生成任务
func (v *VolcengineAIProvider) ProcessVideoTask(ctx context.Context, taskID string, input map[string]interface{}) error {
	// TODO: 实现火山引擎视频生成逻辑
	logrus.Infof("火山引擎视频生成任务处理中: %s", taskID)

	// 模拟处理时间（视频生成通常需要更长时间）
	time.Sleep(10 * time.Second)

	logrus.Infof("火山引擎视频生成任务完成: %s", taskID)
	return nil
}
