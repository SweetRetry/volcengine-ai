package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"volcengine-go-server/config"
)

// VolcengineAIProvider 火山引擎AI服务提供商适配器
type VolcengineAIProvider struct {
	volcengineAIService *VolcengineAIService
	taskService         *TaskService
}

// NewVolcengineAIProvider 创建火山引擎AI服务提供商
func NewVolcengineAIProvider(
	volcengineAIService *VolcengineAIService,
	taskService *TaskService,
) *VolcengineAIProvider {
	return &VolcengineAIProvider{
		volcengineAIService: volcengineAIService,
		taskService:         taskService,
	}
}

// GetProviderName 获取提供商名称
func (v *VolcengineAIProvider) GetProviderName() string {
	return "volcengine"
}

// ProcessImageTask 处理图像生成任务
func (v *VolcengineAIProvider) ProcessImageTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	logrus.Infof("处理火山引擎图像生成任务: %s, 模型: %s", taskID, model)

	// 根据模型选择不同的处理方法
	switch model {
	case config.VolcengineJimengImageModel:
		return v.processJimengImageTask(ctx, taskID, input)
	case config.VolcengineImageModel: // doubao-seedream-3-0-t2i-250415
		return v.processDoubaoImageTask(ctx, taskID, input)
	default:
		// 默认使用豆包模型
		logrus.Warnf("未知模型 %s，使用默认豆包模型", model)
		return v.processDoubaoImageTask(ctx, taskID, input)
	}
}

// processJimengImageTask 处理即梦AI图像生成任务
func (v *VolcengineAIProvider) processJimengImageTask(ctx context.Context, taskID string, input map[string]interface{}) error {
	// 从input参数中获取任务信息
	prompt, ok := input["prompt"].(string)
	if !ok {
		err := fmt.Errorf("无效的prompt参数")
		logrus.Errorf("获取任务输入失败: %v", err)
		v.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	size, _ := input["size"].(string)
	if size == "" {
		size = "1:1" // 默认尺寸
	}

	// 构建即梦AI请求参数
	request := &VolcJimentImageRequest{
		Prompt:    prompt,
		Width:     v.parseJimengImageSize(size, "width"),
		Height:    v.parseJimengImageSize(size, "height"),
		UsePreLLM: len(prompt) < 4, // prompt小于4才开启扩写
		UseSr:     true,            // 开启超分
	}

	// 调用即梦AI图像生成
	result, err := v.volcengineAIService.GenerateImageByJimeng(ctx, request)
	if err != nil {
		logrus.Errorf("即梦AI图像生成失败: %v", err)
		v.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	logrus.Infof("即梦AI图像生成成功: %s (格式: %s)", taskID, result.Format)

	// 根据返回格式处理结果
	var imageURL string
	switch result.Format {
	case "url":
		imageURL = result.ImageURL
	case "base64":
		// 如果返回的是Base64，可以选择保存到文件服务器或直接存储
		// 这里简化处理，直接使用Base64数据作为"URL"（实际应用中需要上传到文件服务器）
		imageURL = "data:image/jpeg;base64," + result.ImageBase64
		logrus.Infof("收到Base64格式图片，长度: %d", len(result.ImageBase64))
	default:
		errorMsg := fmt.Sprintf("未知的图片格式: %s", result.Format)
		logrus.Errorf(errorMsg)
		v.taskService.UpdateTaskError(ctx, taskID, errorMsg)
		return fmt.Errorf(errorMsg)
	}

	// 更新数据库中的任务状态
	if err := v.taskService.UpdateTaskResult(ctx, taskID, imageURL); err != nil {
		logrus.Errorf("更新任务状态失败: %v", err)
		return err
	}

	logrus.Infof("即梦AI任务状态已更新为完成: %s", taskID)
	return nil
}

// processDoubaoImageTask 处理豆包图像生成任务
func (v *VolcengineAIProvider) processDoubaoImageTask(ctx context.Context, taskID string, input map[string]interface{}) error {
	// 从input参数中获取任务信息
	prompt, ok := input["prompt"].(string)
	if !ok {
		err := fmt.Errorf("无效的prompt参数")
		logrus.Errorf("获取任务输入失败: %v", err)
		v.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	size, _ := input["size"].(string)
	if size == "" {
		size = "1024x1024" // 默认尺寸
	}

	// 构建豆包图像生成请求参数
	request := &VolcengineImageRequest{
		Prompt: prompt,
		Model:  config.VolcengineImageModel,
		Size:   v.parseOptimalSizeString(size),
		N:      1, // 生成1张图片
	}

	// 调用豆包图像生成
	result, err := v.volcengineAIService.GenerateImage(ctx, request)
	if err != nil {
		logrus.Errorf("豆包图像生成失败: %v", err)
		v.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	logrus.Infof("豆包图像生成成功: %s (尺寸: %s)", taskID, request.Size)

	// 检查是否有生成的图像
	if len(result.Data) == 0 {
		errorMsg := "未生成任何图像"
		logrus.Errorf("图像生成失败: %s", errorMsg)
		v.taskService.UpdateTaskError(ctx, taskID, errorMsg)
		return fmt.Errorf(errorMsg)
	}

	// 获取第一张图片的URL
	imageURL := result.Data[0].URL
	logrus.Infof("豆包图像生成任务完成: %s, 图像URL: %s", taskID, imageURL)

	// 更新数据库中的任务状态
	if err := v.taskService.UpdateTaskResult(ctx, taskID, imageURL); err != nil {
		logrus.Errorf("更新任务状态失败: %v", err)
		return err
	}

	logrus.Infof("豆包任务状态已更新为完成: %s", taskID)
	return nil
}

// parseJimengImageSize 解析即梦AI的尺寸参数
// 根据官方建议：宽、高与512差距过大，则出图效果不佳、延迟过长概率显著增加
// 超分前建议比例及对应宽高：
// 1:1：512*512, 4:3：512*384, 3:4：384*512, 3:2：512*341, 2:3：341*512, 16:9：512*288, 9:16：288*512
func (v *VolcengineAIProvider) parseJimengImageSize(size string, dimension string) string {
	switch size {
	case "1:1", "1024x1024", "":
		// 1:1 比例 - 512*512
		if dimension == "width" {
			return "512"
		}
		return "512"
	case "4:3", "1152x864":
		// 4:3 比例 - 512*384
		if dimension == "width" {
			return "512"
		}
		return "384"
	case "3:4", "864x1152":
		// 3:4 比例 - 384*512
		if dimension == "width" {
			return "384"
		}
		return "512"
	case "3:2", "1248x832":
		// 3:2 比例 - 512*341
		if dimension == "width" {
			return "512"
		}
		return "341"
	case "2:3", "832x1248":
		// 2:3 比例 - 341*512
		if dimension == "width" {
			return "341"
		}
		return "512"
	case "16:9", "1280x720":
		// 16:9 比例 - 512*288
		if dimension == "width" {
			return "512"
		}
		return "288"
	case "9:16", "720x1280":
		// 9:16 比例 - 288*512
		if dimension == "width" {
			return "288"
		}
		return "512"
	case "21:9", "1512x648":
		// 21:9 比例不在官方推荐中，使用最接近的16:9比例
		logrus.Warnf("21:9比例不在即梦AI官方推荐中，使用16:9比例(512*288)替代以获得最佳效果")
		if dimension == "width" {
			return "512"
		}
		return "288"
	default:
		// 默认使用1:1比例 - 512*512
		logrus.Warnf("未知尺寸格式 %s，使用默认1:1比例(512*512)", size)
		if dimension == "width" {
			return "512"
		}
		return "512"
	}
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
func (v *VolcengineAIProvider) ProcessTextTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	// TODO: 实现火山引擎文本生成逻辑
	logrus.Infof("火山引擎文本生成任务处理中: %s, 模型: %s", taskID, model)

	// 模拟处理时间
	time.Sleep(2 * time.Second)

	logrus.Infof("火山引擎文本生成任务完成: %s", taskID)
	return nil
}

// ProcessVideoTask 处理视频生成任务
func (v *VolcengineAIProvider) ProcessVideoTask(ctx context.Context, taskID string, model string, input map[string]interface{}) error {
	// TODO: 实现火山引擎视频生成逻辑
	logrus.Infof("火山引擎视频生成任务处理中: %s, 模型: %s", taskID, model)

	// 模拟处理时间（视频生成通常需要更长时间）
	time.Sleep(10 * time.Second)

	logrus.Infof("火山引擎视频生成任务完成: %s", taskID)
	return nil
}
