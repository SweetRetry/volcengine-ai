package volcengine

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"volcengine-go-server/internal/util"
)

// GenerateTextByDoubao 豆包文本生成具体实现
func (s *VolcengineService) GenerateTextByDoubao(ctx context.Context, taskID string, input map[string]interface{}) error {
	// TODO: 实现豆包文本生成逻辑
	s.logger.Infof("豆包文本生成任务处理中: %s", taskID)

	// 模拟处理时间
	time.Sleep(2 * time.Second)

	s.logger.Infof("豆包文本生成任务完成: %s", taskID)
	return nil
}

// GenerateVideoByJimeng 即梦AI视频生成具体实现
func (s *VolcengineService) GenerateVideoByJimeng(ctx context.Context, taskID string, input map[string]interface{}) error {
	s.logger.Infof("即梦AI视频生成开始: taskID=%s", taskID)

	// 从input参数中获取任务信息
	prompt, ok := input["prompt"].(string)
	if !ok {
		err := fmt.Errorf("无效的prompt参数")
		s.logger.Errorf("获取任务输入失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	// 检查prompt长度限制
	if len(prompt) > 150 {
		err := fmt.Errorf("prompt长度超过150字符限制，当前长度: %d", len(prompt))
		s.logger.Errorf("prompt长度检查失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	aspectRatio, _ := input["aspect_ratio"].(string)
	if aspectRatio == "" {
		aspectRatio = "16:9" // 默认比例
	}

	seed := -1 // 默认随机种子
	if seedValue, exists := input["seed"]; exists {
		if seedInt, ok := seedValue.(int); ok {
			seed = seedInt
		}
	}

	// 构建即梦AI视频生成请求参数
	request := &JimengVideoRequest{
		Prompt:      prompt,
		Seed:        seed,
		AspectRatio: aspectRatio,
	}

	// 提交视频生成任务
	externalTaskID, err := s.submitJimengVideoTask(ctx, request)
	if err != nil {
		s.logger.Errorf("提交即梦AI视频任务失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	s.logger.Infof("即梦AI视频任务已提交，外部任务ID: %s", externalTaskID)

	// 轮询任务结果
	result, err := s.pollJimengVideoResult(ctx, externalTaskID)
	if err != nil {
		s.logger.Errorf("轮询即梦AI视频任务结果失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	s.logger.Infof("即梦AI视频生成成功: %s, 视频URL: %s", externalTaskID, result.VideoURL)

	// 更新数据库中的任务状态
	if err := s.taskService.UpdateTaskResult(ctx, taskID, result.VideoURL); err != nil {
		s.logger.Errorf("更新任务状态失败: %v", err)
		return err
	}

	s.logger.Infof("即梦AI视频任务状态已更新为完成: %s", taskID)
	return nil
}

// GenerateI2VByJimeng 即梦AI图生视频具体实现
func (s *VolcengineService) GenerateI2VByJimeng(ctx context.Context, taskID string, input map[string]interface{}) error {
	s.logger.Infof("即梦AI图生视频开始: taskID=%s", taskID)

	// 从input参数中获取图片URLs
	imageURLsInterface, ok := input["image_urls"]
	if !ok {
		err := fmt.Errorf("缺少image_urls参数")
		s.logger.Errorf("获取任务输入失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	// 转换图片URLs
	var imageURLs []string
	switch urls := imageURLsInterface.(type) {
	case []string:
		imageURLs = urls
	case []interface{}:
		for _, url := range urls {
			if urlStr, ok := url.(string); ok {
				imageURLs = append(imageURLs, urlStr)
			}
		}
	default:
		err := fmt.Errorf("image_urls参数格式错误")
		s.logger.Errorf("获取任务输入失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	if len(imageURLs) == 0 {
		err := fmt.Errorf("image_urls不能为空")
		s.logger.Errorf("获取任务输入失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	// 获取prompt（可选）
	prompt, _ := input["prompt"].(string)
	if prompt != "" && len(prompt) > 150 {
		err := fmt.Errorf("prompt长度超过150字符限制，当前长度: %d", len(prompt))
		s.logger.Errorf("prompt长度检查失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	// 获取aspect_ratio，如果没有提供则通过图片检测
	aspectRatio, _ := input["aspect_ratio"].(string)
	if aspectRatio == "" {
		// 检测第一张图片的尺寸比例（图生视频只使用第一张图片）
		detectedRatio, err := s.detectImageAspectRatio(ctx, imageURLs[0])
		if err != nil {
			s.logger.Warnf("检测第一张图片尺寸失败，使用默认比例16:9: %v", err)
			aspectRatio = "16:9"
		} else {
			aspectRatio = detectedRatio
			s.logger.Infof("检测到第一张图片尺寸比例: %s", aspectRatio)
		}
	}

	// 验证aspect_ratio是否在支持的范围内
	if !s.isValidAspectRatio(aspectRatio) {
		err := fmt.Errorf("不支持的aspect_ratio: %s，支持的比例: 16:9, 4:3, 1:1, 3:4, 9:16, 21:9, 9:21", aspectRatio)
		s.logger.Errorf("aspect_ratio验证失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	seed := -1 // 默认随机种子
	if seedValue, exists := input["seed"]; exists {
		if seedInt, ok := seedValue.(int); ok {
			seed = seedInt
		}
	}

	// 构建即梦AI图生视频请求参数
	request := &JimengI2VRequest{
		ImageURLs:   imageURLs,
		Prompt:      prompt,
		Seed:        seed,
		AspectRatio: aspectRatio,
	}

	// 提交图生视频任务
	externalTaskID, err := s.submitJimengI2VTask(ctx, request)
	if err != nil {
		s.logger.Errorf("提交即梦AI图生视频任务失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	s.logger.Infof("即梦AI图生视频任务已提交，外部任务ID: %s", externalTaskID)

	// 轮询任务结果（复用文生视频的轮询逻辑）
	result, err := s.pollJimengVideoResult(ctx, externalTaskID)
	if err != nil {
		s.logger.Errorf("轮询即梦AI图生视频任务结果失败: %v", err)
		s.taskService.UpdateTaskError(ctx, taskID, err.Error())
		return err
	}

	s.logger.Infof("即梦AI图生视频生成成功: %s, 视频URL: %s", externalTaskID, result.VideoURL)

	// 更新数据库中的任务状态
	if err := s.taskService.UpdateTaskResult(ctx, taskID, result.VideoURL); err != nil {
		s.logger.Errorf("更新任务状态失败: %v", err)
		return err
	}

	s.logger.Infof("即梦AI图生视频任务状态已更新为完成: %s", taskID)
	return nil
}

// submitJimengVideoTask 提交即梦AI视频生成任务
func (s *VolcengineService) submitJimengVideoTask(ctx context.Context, request *JimengVideoRequest) (string, error) {
	s.logger.Infof("开始调用即梦AI视频生成API: prompt=%s", request.Prompt)

	// 构建即梦AI视频任务参数
	taskParams := map[string]interface{}{
		"req_key":      "jimeng_vgfm_t2v_l20", // 即梦AI视频服务标识
		"prompt":       request.Prompt,
		"aspect_ratio": request.AspectRatio,
	}

	// 如果指定了种子，添加到参数中
	if request.Seed != -1 {
		taskParams["seed"] = request.Seed
	}

	// 记录详细的API调用信息
	s.logger.WithFields(logrus.Fields{
		"api_endpoint": "cvSync2AsyncSubmitTask",
		"req_key":      taskParams["req_key"],
		"prompt":       taskParams["prompt"],
		"aspect_ratio": taskParams["aspect_ratio"],
		"seed":         taskParams["seed"],
	}).Info("即梦AI视频API调用开始")

	// 调用cvSync2AsyncSubmitTask提交任务
	startTime := time.Now()
	resp, status, err := s.visualClient.CVSync2AsyncSubmitTask(taskParams)
	duration := time.Since(startTime)

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"api_endpoint": "cvSync2AsyncSubmitTask",
			"duration_ms":  duration.Milliseconds(),
			"status_code":  status,
			"error":        err.Error(),
		}).Error("即梦AI视频API调用失败")
		return "", fmt.Errorf("提交即梦AI视频任务失败: %v", err)
	}

	// 记录成功的API调用
	s.logger.WithFields(logrus.Fields{
		"api_endpoint": "cvSync2AsyncSubmitTask",
		"duration_ms":  duration.Milliseconds(),
		"status_code":  status,
		"response":     resp,
	}).Info("即梦AI视频API调用成功")

	// 检查响应是否包含task_id（异步任务）
	if taskID, ok := resp["task_id"].(string); ok && taskID != "" {
		s.logger.Infof("即梦AI视频任务提交成功，获得task_id: %s", taskID)
		return taskID, nil
	}

	// 如果没有task_id，检查是否有其他标识符
	if data, exists := resp["data"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			// 检查是否有task_id在data中
			if taskID, exists := dataMap["task_id"]; exists {
				if taskIDStr, ok := taskID.(string); ok && taskIDStr != "" {
					s.logger.Infof("即梦AI视频任务提交成功，从data中获得task_id: %s", taskIDStr)
					return taskIDStr, nil
				}
			}
		}
	}

	return "", fmt.Errorf("响应中未找到有效的task_id")
}

// submitJimengI2VTask 提交即梦AI图生视频任务
func (s *VolcengineService) submitJimengI2VTask(ctx context.Context, request *JimengI2VRequest) (string, error) {
	s.logger.Infof("开始调用即梦AI图生视频API: image_count=%d", len(request.ImageURLs))

	// 构建即梦AI图生视频任务参数
	taskParams := map[string]interface{}{
		"req_key":      "jimeng_vgfm_i2v_l20", // 即梦AI图生视频服务标识
		"image_urls":   request.ImageURLs,
		"aspect_ratio": request.AspectRatio,
	}

	// 如果有prompt，添加到参数中
	if request.Prompt != "" {
		taskParams["prompt"] = request.Prompt
	}

	// 如果指定了种子，添加到参数中
	if request.Seed != -1 {
		taskParams["seed"] = request.Seed
	}

	// 记录详细的API调用信息
	s.logger.WithFields(logrus.Fields{
		"api_endpoint": "cvSync2AsyncSubmitTask",
		"req_key":      taskParams["req_key"],
		"image_count":  len(request.ImageURLs),
		"prompt":       taskParams["prompt"],
		"aspect_ratio": taskParams["aspect_ratio"],
		"seed":         taskParams["seed"],
	}).Info("即梦AI图生视频API调用开始")

	// 调用cvSync2AsyncSubmitTask提交任务
	startTime := time.Now()
	resp, status, err := s.visualClient.CVSync2AsyncSubmitTask(taskParams)
	duration := time.Since(startTime)

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"api_endpoint": "cvSync2AsyncSubmitTask",
			"duration_ms":  duration.Milliseconds(),
			"status_code":  status,
			"error":        err.Error(),
		}).Error("即梦AI图生视频API调用失败")
		return "", fmt.Errorf("提交即梦AI图生视频任务失败: %v", err)
	}

	// 记录成功的API调用
	s.logger.WithFields(logrus.Fields{
		"api_endpoint": "cvSync2AsyncSubmitTask",
		"duration_ms":  duration.Milliseconds(),
		"status_code":  status,
		"response":     resp,
	}).Info("即梦AI图生视频API调用成功")

	// 检查响应是否包含task_id（异步任务）
	if taskID, ok := resp["task_id"].(string); ok && taskID != "" {
		s.logger.Infof("即梦AI图生视频任务提交成功，获得task_id: %s", taskID)
		return taskID, nil
	}

	// 如果没有task_id，检查是否有其他标识符
	if data, exists := resp["data"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			// 检查是否有task_id在data中
			if taskID, exists := dataMap["task_id"]; exists {
				if taskIDStr, ok := taskID.(string); ok && taskIDStr != "" {
					s.logger.Infof("即梦AI图生视频任务提交成功，从data中获得task_id: %s", taskIDStr)
					return taskIDStr, nil
				}
			}
		}
	}

	return "", fmt.Errorf("响应中未找到有效的task_id")
}

// pollJimengVideoResult 轮询即梦AI视频生成结果
func (s *VolcengineService) pollJimengVideoResult(ctx context.Context, taskID string) (*JimengVideoResult, error) {
	// 创建即梦AI视频结果检查器
	checker := &JimengVideoResultChecker{service: s}

	// 使用默认的自适应轮询配置
	config := util.DefaultPollConfig("即梦AI视频").WithLogger(s.logger)

	// 使用通用轮询方法
	result, err := util.PollTaskResult(ctx, taskID, checker, config)
	if err != nil {
		return nil, err
	}

	// 类型断言转换结果
	videoResult, ok := result.(*JimengVideoResult)
	if !ok {
		return nil, fmt.Errorf("轮询结果类型转换失败")
	}

	return videoResult, nil
}

// queryJimengVideoResult 查询即梦AI视频任务结果
func (s *VolcengineService) queryJimengVideoResult(ctx context.Context, taskID string) (*JimengVideoResult, error) {
	// 构建查询参数
	queryParams := map[string]interface{}{
		"req_key": "jimeng_vgfm_t2v_l20", // 即梦AI视频服务标识
		"task_id": taskID,
	}

	// 记录详细的API调用信息
	s.logger.WithFields(logrus.Fields{
		"api_endpoint": "CVGetResult",
		"req_key":      queryParams["req_key"],
		"task_id":      queryParams["task_id"],
	}).Info("即梦AI视频结果查询API调用开始")

	// 调用CVGetResult查询结果
	startTime := time.Now()
	resp, status, err := s.visualClient.CVGetResult(queryParams)
	duration := time.Since(startTime)

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"api_endpoint": "CVGetResult",
			"duration_ms":  duration.Milliseconds(),
			"status_code":  status,
			"task_id":      taskID,
			"error":        err.Error(),
		}).Error("即梦AI视频结果查询API调用失败")
		return nil, fmt.Errorf("查询即梦AI视频任务结果失败: %v", err)
	}

	// 记录成功的API调用
	s.logger.WithFields(logrus.Fields{
		"api_endpoint": "CVGetResult",
		"duration_ms":  duration.Milliseconds(),
		"status_code":  status,
		"task_id":      taskID,
		"response":     resp,
	}).Info("即梦AI视频结果查询API调用成功")

	// 解析响应获取结果
	result, err := s.parseJimengVideoResultResponse(resp)
	if err != nil {
		s.logger.Errorf("解析即梦AI视频结果响应失败: %v", err)
		return nil, fmt.Errorf("解析结果响应失败: %v", err)
	}

	return result, nil
}

// parseJimengVideoResultResponse 解析即梦AI视频结果响应
func (s *VolcengineService) parseJimengVideoResultResponse(resp map[string]interface{}) (*JimengVideoResult, error) {
	// 解析data字段
	data, exists := resp["data"]
	if !exists {
		return nil, fmt.Errorf("响应中缺少data字段")
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data字段格式错误")
	}

	// 获取status
	status, exists := dataMap["status"]
	if !exists {
		return nil, fmt.Errorf("响应中缺少status字段")
	}

	statusStr, ok := status.(string)
	if !ok {
		return nil, fmt.Errorf("status字段格式错误")
	}

	// 构建结果对象
	result := &JimengVideoResult{
		Status: statusStr,
	}

	// 如果任务完成，获取视频URL
	if statusStr == "done" {
		if videoURL, exists := dataMap["video_url"]; exists {
			if videoURLStr, ok := videoURL.(string); ok && videoURLStr != "" {
				result.VideoURL = videoURLStr
			}
		}

		// 如果没有找到视频URL，任务虽然完成但结果无效
		if result.VideoURL == "" {
			return nil, fmt.Errorf("任务已完成但未找到视频URL")
		}
	}

	return result, nil
}

// JimengVideoResultChecker 即梦AI视频结果检查器
type JimengVideoResultChecker struct {
	service *VolcengineService
}

// CheckResult 实现util.TaskResultChecker接口
func (c *JimengVideoResultChecker) CheckResult(ctx context.Context, taskID string) (interface{}, bool, error) {
	result, err := c.service.queryJimengVideoResult(ctx, taskID)
	if err != nil {
		return nil, false, err
	}

	// 检查任务状态
	isCompleted := result.Status == "done"
	return result, isCompleted, nil
}
