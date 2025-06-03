package volcengine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"

	"volcengine-go-server/config"
)

// GenerateImageByDoubao 豆包图像生成具体实现
func (s *VolcengineService) GenerateImageByDoubao(ctx context.Context, taskID string, input map[string]interface{}) error {
	s.logger.Infof("豆包图像生成开始: taskID=%s", taskID)

	// 从input参数中获取任务信息
	prompt, ok := input["prompt"].(string)
	if !ok {
		err := fmt.Errorf("无效的prompt参数")
		s.logger.Errorf("获取任务输入失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	aspectRatio, _ := input["aspect_ratio"].(string)
	if aspectRatio == "" {
		aspectRatio = "1:1" // 默认比例
	}

	// 构建豆包图像生成请求参数
	request := &VolcengineImageRequest{
		Prompt: prompt,
		Model:  config.VolcengineImageModel,
		Size:   s.parseOptimalSizeString(aspectRatio),
		N:      1, // 生成1张图片
	}

	// 调用豆包图像生成
	result, err := s.generateImage(ctx, request)
	if err != nil {
		s.logger.Errorf("豆包图像生成失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	s.logger.Infof("豆包图像生成成功: %s (比例: %s)", taskID, aspectRatio)

	// 检查是否有生成的图像
	if len(result.Data) == 0 {
		errorMsg := "未生成任何图像"
		s.logger.Errorf("图像生成失败: %s", errorMsg)
		s.taskService.UpdateTaskError(ctx, taskID, errorMsg)
		return errors.New(errorMsg)
	}

	// 获取第一张图片的URL
	imageURL := result.Data[0].URL
	s.logger.Infof("豆包图像生成任务完成: %s, 图像URL: %s", taskID, imageURL)

	// 更新数据库中的任务状态
	if err := s.taskService.UpdateTaskResult(ctx, taskID, imageURL); err != nil {
		s.logger.Errorf("更新任务状态失败: %v", err)
		return err
	}

	s.logger.Infof("豆包任务状态已更新为完成: %s", taskID)
	return nil
}

// GenerateImageByJimeng 即梦AI图像生成具体实现
func (s *VolcengineService) GenerateImageByJimeng(ctx context.Context, taskID string, input map[string]interface{}) error {
	s.logger.Infof("即梦AI图像生成开始: taskID=%s", taskID)

	// 从input参数中获取任务信息
	prompt, ok := input["prompt"].(string)
	if !ok {
		err := fmt.Errorf("无效的prompt参数")
		s.logger.Errorf("获取任务输入失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	aspectRatio, _ := input["aspect_ratio"].(string)
	if aspectRatio == "" {
		aspectRatio = "1:1" // 默认比例
	}

	// 解析即梦AI图像尺寸
	imageSize := s.parseJimengImageSize(aspectRatio)

	// 构建即梦AI请求参数
	request := &VolcJimentImageRequest{
		Prompt:    prompt,
		Width:     imageSize.Width,
		Height:    imageSize.Height,
		UsePreLLM: len(prompt) < 4, // prompt小于4才开启扩写
		UseSr:     true,            // 开启超分
	}

	// 调用即梦AI图像生成
	result, err := s.generateImageByJimeng(ctx, request)
	if err != nil {
		s.logger.Errorf("即梦AI图像生成失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	s.logger.Infof("即梦AI图像生成成功: %s", taskID)

	// 获取图片URL
	imageURL := result.ImageURL
	s.logger.Infof("即梦AI图像生成任务完成: %s, 图像URL: %s", taskID, imageURL)

	// 更新数据库中的任务状态
	if err := s.taskService.UpdateTaskResult(ctx, taskID, imageURL); err != nil {
		s.logger.Errorf("更新任务状态失败: %v", err)
		return err
	}

	s.logger.Infof("即梦AI任务状态已更新为完成: %s", taskID)
	return nil
}

// generateImage 生成图像（同步）- 内部方法
func (s *VolcengineService) generateImage(ctx context.Context, request *VolcengineImageRequest) (*VolcengineImageResponse, error) {
	s.logger.Infof("开始调用火山方舟图像生成API: prompt=%s", request.Prompt)

	// 设置默认模型
	modelID := request.Model
	if modelID == "" {
		modelID = config.VolcengineImageModel
	}

	// 构建请求
	size := request.Size
	if size == "" {
		size = config.DefaultImageSize
	}

	// 设置水印为false
	watermark := false
	generateReq := model.GenerateImagesRequest{
		Model:     modelID,
		Prompt:    request.Prompt,
		Size:      &size,
		Watermark: &watermark,
	}

	// 记录详细的API调用信息
	s.logger.WithFields(logrus.Fields{
		"api_endpoint": "GenerateImages",
		"model":        modelID,
		"prompt":       request.Prompt,
		"size":         size,
		"watermark":    watermark,
	}).Info("火山方舟API调用开始")

	// 调用火山方舟图像生成API
	startTime := time.Now()
	imagesResponse, err := s.client.GenerateImages(ctx, generateReq)
	duration := time.Since(startTime)

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"api_endpoint": "GenerateImages",
			"duration_ms":  duration.Milliseconds(),
			"error":        err.Error(),
		}).Error("火山方舟API调用失败")
		return nil, fmt.Errorf("图像生成失败: %v", err)
	}

	// 记录成功的API调用
	s.logger.WithFields(logrus.Fields{
		"api_endpoint":   "GenerateImages",
		"duration_ms":    duration.Milliseconds(),
		"response_count": len(imagesResponse.Data),
	}).Info("火山方舟API调用成功")

	// 转换响应格式
	response := &VolcengineImageResponse{
		Data:    make([]ImageData, len(imagesResponse.Data)),
		Created: time.Now().Unix(),
	}

	for i, data := range imagesResponse.Data {
		response.Data[i] = ImageData{
			URL: *data.Url,
		}
		s.logger.Infof("生成图片 %d: URL=%s", i+1, *data.Url)
	}

	s.logger.Infof("图像生成成功，生成了 %d 张图片", len(response.Data))
	return response, nil
}

// generateImageByJimeng 即梦AI图像生成 - 内部方法
func (s *VolcengineService) generateImageByJimeng(ctx context.Context, request *VolcJimentImageRequest) (*JimengImageResult, error) {
	s.logger.Infof("开始调用即梦AI图像生成API: prompt=%s", request.Prompt)

	// 构建即梦AI任务参数 - 根据官方文档
	taskParams := map[string]interface{}{
		"req_key":     "jimeng_high_aes_general_v21_L", // 即梦AI服务标识
		"prompt":      request.Prompt,
		"width":       request.Width,
		"height":      request.Height,
		"use_pre_llm": len(request.Prompt) < 4, // promot小于4才开启扩写
		"use_sr":      true,                    // 开启AIGC超分
		"return_url":  true,                    // 返回图片链接
	}

	// 记录详细的API调用信息
	s.logger.WithFields(logrus.Fields{
		"api_endpoint": "CVProcess",
		"req_key":      taskParams["req_key"],
		"prompt":       taskParams["prompt"],
		"width":        taskParams["width"],
		"height":       taskParams["height"],
		"use_pre_llm":  taskParams["use_pre_llm"],
		"use_sr":       taskParams["use_sr"],
		"return_url":   taskParams["return_url"],
	}).Info("即梦AI API调用开始")

	// 调用CVProcess提交任务
	startTime := time.Now()
	resp, status, err := s.visualClient.CVProcess(taskParams)
	duration := time.Since(startTime)

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"api_endpoint": "CVProcess",
			"duration_ms":  duration.Milliseconds(),
			"status_code":  status,
			"error":        err.Error(),
		}).Error("即梦AI API调用失败")
		return nil, fmt.Errorf("提交即梦AI任务失败: %v", err)
	}

	// 记录成功的API调用
	s.logger.WithFields(logrus.Fields{
		"api_endpoint": "CVProcess",
		"duration_ms":  duration.Milliseconds(),
		"status_code":  status,
		"response":     resp,
	}).Info("即梦AI API调用成功")

	// 解析响应获取图片数据
	result, err := s.parseJimengResponse(resp)
	if err != nil {
		s.logger.Errorf("解析即梦AI响应失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result == nil {
		return nil, fmt.Errorf("未找到图片数据")
	}

	return result, nil
}

// parseJimengResponse 解析即梦AI响应，支持多种返回格式
func (s *VolcengineService) parseJimengResponse(resp map[string]interface{}) (*JimengImageResult, error) {
	// 检查响应中是否存在data字段
	data, exists := resp["data"]
	if !exists {
		return nil, fmt.Errorf("响应中缺少data字段")
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data字段格式错误，期望为对象类型")
	}

	// 优先尝试解析image_urls
	if imageUrls, exists := dataMap["image_urls"]; exists {
		if urlArray, ok := imageUrls.([]interface{}); ok && len(urlArray) > 0 {
			if imageUrl, ok := urlArray[0].(string); ok && imageUrl != "" {
				s.logger.Infof("成功解析图片URL: %s", imageUrl)
				return &JimengImageResult{
					ImageURL: imageUrl,
				}, nil
			}
		}
	}

	// 如果没有image_urls，尝试解析binary_data_base64
	if binaryData, exists := dataMap["binary_data_base64"]; exists {
		if base64Array, ok := binaryData.([]interface{}); ok && len(base64Array) > 0 {
			if imageBase64, ok := base64Array[0].(string); ok && imageBase64 != "" {
				s.logger.Infof("成功解析图片Base64数据，长度: %d", len(imageBase64))
				return &JimengImageResult{
					ImageURL: "data:image/jpeg;base64," + imageBase64,
				}, nil
			}
		}
	}

	// 记录可用的字段以便调试
	availableKeys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		availableKeys = append(availableKeys, k)
	}
	s.logger.Warnf("响应中未找到有效的图片数据，可用字段: %v", availableKeys)
	return nil, fmt.Errorf("响应中未找到有效的图片数据")
}
