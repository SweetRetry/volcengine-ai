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

	// 构建火山引擎API选项
	options := map[string]interface{}{
		"model":   taskInput.Model,
		"size":    taskInput.Size,
		"quality": taskInput.Quality,
		"style":   taskInput.Style,
		"n":       taskInput.N,
	}

	// 调用火山引擎AI服务创建任务
	volcengineResp, err := v.volcengineAIService.CreateImageTask(ctx, taskInput.Prompt, options)
	if err != nil {
		logrus.Errorf("调用火山引擎AI服务失败: %v", err)
		v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", err.Error())
		return err
	}

	logrus.Infof("火山引擎任务创建成功: %s -> %s", taskID, volcengineResp.TaskID)

	// 轮询查询火山引擎任务结果
	maxAttempts := 60 // 最多轮询60次，每次间隔5秒，总共5分钟
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 查询火山引擎任务结果
		result, err := v.volcengineAIService.GetTaskResult(ctx, volcengineResp.TaskID)
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
		if result.Error != "" {
			logrus.Errorf("火山引擎任务失败: %s", result.Error)
			v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", result.Error)
			return fmt.Errorf("火山引擎任务失败: %s", result.Error)
		}

		// 如果有图像URL，说明任务完成
		if result.ImageURL != "" {
			logrus.Infof("图像生成任务完成: %s, 图像URL: %s", taskID, result.ImageURL)

			// 更新数据库中的任务状态
			if err := v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "completed", result.ImageURL, ""); err != nil {
				logrus.Errorf("更新任务状态失败: %v", err)
				return err
			}

			logrus.Infof("任务状态已更新为完成: %s", taskID)
			return nil
		}

		// 任务还在处理中，等待后继续轮询
		logrus.Infof("火山引擎任务处理中: %s, 等待5秒后重试 (尝试 %d/%d)", volcengineResp.TaskID, attempt+1, maxAttempts)
		time.Sleep(5 * time.Second)
	}

	// 超时处理
	errorMsg := "任务处理超时，超过5分钟未完成"
	logrus.Errorf("图像生成任务超时: %s", taskID)
	v.imageTaskService.UpdateImageTaskStatus(ctx, taskID, "failed", "", errorMsg)
	return fmt.Errorf(errorMsg)
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
