package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
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
	return "volcengine_jimeng"
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

	// 构建即梦AI请求参数
	request := &JimengImageRequest{
		Prompt:    taskInput.Prompt,
		Width:     512,  // 默认宽度
		Height:    512,  // 默认高度
		ReturnURL: true, // 返回图片链接
	}

	// 智能设置文本扩写参数
	request.UsePreLLM = v.shouldUsePreLLM(taskInput.Prompt)

	// 解析尺寸参数并设置推荐的宽高比例
	request.Width, request.Height = v.parseOptimalSize(taskInput.Size)

	// 智能设置超分参数（根据尺寸和性能需求）
	request.UseSR = v.shouldUseSR(request.Width, request.Height)

	// 提交火山引擎即梦AI任务
	volcengineTaskID, err := v.volcengineAIService.SubmitImageTask(ctx, request)
	if err != nil {
		logrus.Errorf("提交火山引擎即梦AI任务失败: %v", err)
		v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", err.Error())
		return err
	}

	logrus.Infof("火山引擎即梦AI任务提交成功: %s -> %s (尺寸: %dx%d, 扩写: %t, 超分: %t)",
		taskID, volcengineTaskID, request.Width, request.Height, request.UsePreLLM, request.UseSR)

	// 轮询查询火山引擎任务结果
	maxAttempts := 60 // 最多轮询60次，每次间隔5秒，总共5分钟
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 查询火山引擎任务结果
		result, err := v.volcengineAIService.GetImageTaskResult(ctx, volcengineTaskID)
		if err != nil {
			logrus.Errorf("查询火山引擎任务结果失败: %v", err)
			if attempt == maxAttempts-1 {
				v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", err.Error())
				return err
			}
			time.Sleep(5 * time.Second)
			continue
		}

		// 检查任务状态
		if result.Code != 0 {
			logrus.Errorf("火山引擎任务失败: %s (code: %d)", result.Message, result.Code)
			v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", result.Message)
			return fmt.Errorf("火山引擎任务失败: %s (code: %d)", result.Message, result.Code)
		}

		// 检查是否有图像URL
		if result.Data.PrimaryImageUrl != "" {
			logrus.Infof("图像生成任务完成: %s, 图像URL: %s", taskID, result.Data.PrimaryImageUrl)

			// 更新数据库中的任务状态
			if err := v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "completed", result.Data.PrimaryImageUrl, ""); err != nil {
				logrus.Errorf("更新任务状态失败: %v", err)
				return err
			}

			logrus.Infof("任务状态已更新为完成: %s", taskID)
			return nil
		}

		// 如果有多个图像URL，使用第一个
		if len(result.Data.ImageUrls) > 0 {
			imageURL := result.Data.ImageUrls[0]
			logrus.Infof("图像生成任务完成: %s, 图像URL: %s", taskID, imageURL)

			// 更新数据库中的任务状态
			if err := v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "completed", imageURL, ""); err != nil {
				logrus.Errorf("更新任务状态失败: %v", err)
				return err
			}

			logrus.Infof("任务状态已更新为完成: %s", taskID)
			return nil
		}

		// 任务还在处理中，等待后继续轮询
		logrus.Infof("火山引擎任务处理中: %s, 等待5秒后重试 (尝试 %d/%d)", volcengineTaskID, attempt+1, maxAttempts)
		time.Sleep(5 * time.Second)
	}

	// 超时处理
	errorMsg := "任务处理超时，超过5分钟未完成"
	logrus.Errorf("图像生成任务超时: %s", taskID)
	v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", errorMsg)
	return fmt.Errorf(errorMsg)
}

// shouldUsePreLLM 智能判断是否使用文本扩写
func (v *VolcengineAIProvider) shouldUsePreLLM(prompt string) bool {
	// 根据官方建议：prompt过短（长度小于4）推荐开启扩写
	if len(prompt) < 4 {
		return true
	}

	// prompt较长时，可以考虑关闭扩写以保证多样性
	// 这里设置一个阈值，超过100个字符的长prompt建议关闭扩写
	if len(prompt) > 100 {
		return false
	}

	// 中等长度的prompt默认开启扩写
	return true
}

// parseOptimalSize 解析并返回最优的图像尺寸
func (v *VolcengineAIProvider) parseOptimalSize(size string) (width, height int) {
	// 根据官方推荐的超分前比例及对应宽高
	switch size {
	case "1:1", "512x512", "":
		return 512, 512 // 1:1 比例
	case "4:3", "512x384":
		return 512, 384 // 4:3 比例
	case "3:4", "384x512":
		return 384, 512 // 3:4 比例
	case "3:2", "512x341":
		return 512, 341 // 3:2 比例
	case "2:3", "341x512":
		return 341, 512 // 2:3 比例
	case "16:9", "512x288":
		return 512, 288 // 16:9 比例
	case "9:16", "288x512":
		return 288, 512 // 9:16 比例

	default:
		// 默认使用推荐的1:1比例
		return 512, 512
	}
}

// shouldUseSR 智能判断是否使用超分功能
func (v *VolcengineAIProvider) shouldUseSR(width, height int) bool {
	// 根据官方建议：超分会增加延迟，但能提升图像质量
	// 对于较小尺寸的图像，建议开启超分以获得更好的效果
	// 对于已经较大的尺寸，可以关闭超分以减少延迟

	totalPixels := width * height

	// 小于等于512x512的图像建议开启超分
	if totalPixels <= 512*512 {
		return true
	}

	// 大于512x512的图像可以关闭超分以减少延迟
	// 特别是接近768x768的大尺寸图像
	return false
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
