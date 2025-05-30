package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/volcengine/volc-sdk-golang/service/visual"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/util"
	"volcengine-go-server/pkg/logger"
)

// VolcengineService 火山引擎AI服务 - Service层，负责具体的API调用实现
type VolcengineService struct {
	config       config.AIConfig
	client       *arkruntime.Client
	logger       *logrus.Logger
	visualClient *visual.Visual
	taskService  *TaskService
}

// 即梦AI图像尺寸信息
type JimengImageSize struct {
	Width  string `json:"width"`
	Height string `json:"height"`
}

// 图像生成请求结构
type VolcengineImageRequest struct {
	Prompt string `json:"prompt"`          // 必填：文本描述
	Model  string `json:"model,omitempty"` // 模型ID，默认使用豆包图像生成模型
	Size   string `json:"size,omitempty"`  // 图像尺寸，如"1024x1024"
	N      int    `json:"n,omitempty"`     // 生成图片数量，默认1
}

// 图像生成响应结构
type VolcengineImageResponse struct {
	Data    []ImageData `json:"data"`
	Created int64       `json:"created"`
}

type ImageData struct {
	URL string `json:"url"` // 图片URL
}

// 即梦AI图像生成响应结构
type JimengImageResult struct {
	ImageURL string `json:"image_url"` // 图片URL
}

type VolcJimentImageRequest struct {
	Prompt    string `json:"prompt"`
	Width     string `json:"width"`
	Height    string `json:"height"`
	UsePreLLM bool   `json:"use_pre_llm"`
	UseSr     bool   `json:"use_sr"`
}

// 即梦AI视频生成请求结构
type JimengVideoRequest struct {
	Prompt      string `json:"prompt"`                 // 必填：生成视频的提示词，支持中英文，150字符以内
	Seed        int    `json:"seed,omitempty"`         // 可选：随机种子，默认-1（随机）
	AspectRatio string `json:"aspect_ratio,omitempty"` // 可选：生成视频的尺寸，默认16:9
}

// 即梦AI视频生成结果
type JimengVideoResult struct {
	VideoURL string `json:"video_url"` // 视频URL
	Status   string `json:"status"`    // 任务状态
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

// NewVolcengineService 创建火山引擎AI服务实例
func NewVolcengineService(cfg config.AIConfig, taskService *TaskService) *VolcengineService {
	// 设置API Key到环境变量
	if cfg.VolcengineAPIKey != "" {
		os.Setenv("ARK_API_KEY", cfg.VolcengineAPIKey)
	}

	// 创建火山方舟客户端
	client := arkruntime.NewClientWithApiKey(
		os.Getenv("ARK_API_KEY"),
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)

	visualClient := visual.NewInstance()
	visualClient.Client.SetAccessKey(cfg.VolcengineAccessKey)
	visualClient.Client.SetSecretKey(cfg.VolcengineSecretKey)

	return &VolcengineService{
		config:       cfg,
		client:       client,
		visualClient: visualClient,
		logger:       logger.GetLogger(), // 使用全局日志记录器
		taskService:  taskService,
	}
}

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
		return fmt.Errorf(errorMsg)
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

// parseJimengImageSize 解析宽高比并返回即梦AI的尺寸参数
// 即梦AI要求：width和height取值范围[256, 768]，默认值512
func (s *VolcengineService) parseJimengImageSize(aspectRatio string) JimengImageSize {
	switch aspectRatio {
	case "1:1", "":
		// 1:1 比例 - 512*512 (默认)
		return JimengImageSize{
			Width:  "512",
			Height: "512",
		}
	case "4:3":
		// 4:3 比例 - 768*576 (在范围内的最大尺寸)
		return JimengImageSize{
			Width:  "768",
			Height: "576",
		}
	case "3:4":
		// 3:4 比例 - 576*768 (在范围内的最大尺寸)
		return JimengImageSize{
			Width:  "576",
			Height: "768",
		}
	case "3:2":
		// 3:2 比例 - 768*512
		return JimengImageSize{
			Width:  "768",
			Height: "512",
		}
	case "2:3":
		// 2:3 比例 - 512*768
		return JimengImageSize{
			Width:  "512",
			Height: "768",
		}
	case "16:9":
		// 16:9 比例 - 768*432
		return JimengImageSize{
			Width:  "768",
			Height: "432",
		}
	case "9:16":
		// 9:16 比例 - 432*768
		return JimengImageSize{
			Width:  "432",
			Height: "768",
		}
	case "21:9":
		// 21:9 比例 - 768*329 (接近21:9比例，在范围内)
		return JimengImageSize{
			Width:  "768",
			Height: "329",
		}
	default:
		// 默认使用1:1比例 - 512*512
		s.logger.Warnf("未知宽高比格式 %s，使用默认1:1比例(512*512)", aspectRatio)
		return JimengImageSize{
			Width:  "512",
			Height: "512",
		}
	}
}

// parseOptimalSizeString 解析宽高比并返回最优的图像尺寸字符串（用于豆包模型）
func (s *VolcengineService) parseOptimalSizeString(aspectRatio string) string {
	// 火山方舟支持的尺寸格式
	switch aspectRatio {
	case "1:1", "":
		return config.ImageSize1x1
	case "3:4":
		return config.ImageSize3x4
	case "4:3":
		return config.ImageSize4x3
	case "16:9":
		return config.ImageSize16x9
	case "9:16":
		return config.ImageSize9x16
	case "2:3":
		return config.ImageSize2x3
	case "3:2":
		return config.ImageSize3x2
	case "21:9":
		return config.ImageSize21x9
	default:
		// 默认使用1:1比例
		return config.DefaultImageSize
	}
}

// HealthCheck 健康检查
func (s *VolcengineService) HealthCheck(ctx context.Context) error {
	// 简单的健康检查
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

// pollJimengVideoResult 轮询即梦AI视频生成结果
func (s *VolcengineService) pollJimengVideoResult(ctx context.Context, taskID string) (*JimengVideoResult, error) {
	// 创建即梦AI视频结果检查器
	checker := &JimengVideoResultChecker{service: s}

	// 配置轮询参数
	config := util.NewPollConfig("即梦AI视频", 60, 10*time.Second).WithLogger(s.logger)

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
