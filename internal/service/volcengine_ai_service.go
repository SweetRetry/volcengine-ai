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
)

// VolcengineAIService 火山方舟AI服务
type VolcengineAIService struct {
	config       config.AIConfig
	client       *arkruntime.Client
	logger       *logrus.Logger
	visualClient *visual.Visual
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
	ImageURL    string `json:"image_url,omitempty"`    // 图片URL
	ImageBase64 string `json:"image_base64,omitempty"` // 图片Base64数据
	Format      string `json:"format"`                 // 返回格式类型：url 或 base64
}

// NewVolcengineAIService 创建火山方舟AI服务实例
func NewVolcengineAIService(cfg config.AIConfig) *VolcengineAIService {
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

	return &VolcengineAIService{
		config:       cfg,
		client:       client,
		visualClient: visualClient,
		logger:       logrus.New(),
	}
}

// GenerateImage 生成图像（同步）
func (s *VolcengineAIService) GenerateImage(ctx context.Context, request *VolcengineImageRequest) (*VolcengineImageResponse, error) {
	s.logger.Infof("生成图像: prompt=%s", request.Prompt)

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

	// 调用火山方舟图像生成API
	imagesResponse, err := s.client.GenerateImages(ctx, generateReq)
	if err != nil {
		s.logger.Errorf("图像生成失败: %v", err)
		return nil, fmt.Errorf("图像生成失败: %v", err)
	}

	// 转换响应格式
	response := &VolcengineImageResponse{
		Data:    make([]ImageData, len(imagesResponse.Data)),
		Created: time.Now().Unix(),
	}

	for i, data := range imagesResponse.Data {
		response.Data[i] = ImageData{
			URL: *data.Url,
		}
	}

	s.logger.Infof("图像生成成功，生成了 %d 张图片", len(response.Data))
	return response, nil
}

type VolcJimentImageRequest struct {
	Prompt    string `json:"prompt"`
	Width     string `json:"width"`
	Height    string `json:"height"`
	UsePreLLM bool   `json:"use_pre_llm"`
	UseSr     bool   `json:"use_sr"`
}

func (s *VolcengineAIService) GenerateImageByJimeng(ctx context.Context, request *VolcJimentImageRequest) (*JimengImageResult, error) {
	s.logger.Infof("提交即梦AI图像生成任务: prompt=%s", request.Prompt)

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

	s.logger.Infof("提交任务参数: %v", taskParams)
	// 调用CVSubmitTask提交任务
	resp, status, err := s.visualClient.CVProcess(taskParams)
	if err != nil {
		s.logger.Errorf("提交即梦AI任务失败: %v", err)
		return nil, fmt.Errorf("提交即梦AI任务失败: %v", err)
	}

	s.logger.Infof("提交任务响应: %v", resp)
	s.logger.Infof("提交任务响应状态: %d", status)

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
func (s *VolcengineAIService) parseJimengResponse(resp map[string]interface{}) (*JimengImageResult, error) {
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
					Format:   "url",
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
					ImageBase64: imageBase64,
					Format:      "base64",
				}, nil
			}
		}
	}

	// 记录可用的字段以便调试
	s.logger.Warnf("响应中未找到有效的图片数据，可用字段: %v", getMapKeys(dataMap))
	return nil, fmt.Errorf("响应中未找到有效的图片数据")
}

// getMapKeys 获取map的所有键，用于调试
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// HealthCheck 健康检查
func (s *VolcengineAIService) HealthCheck(ctx context.Context) error {
	// 简单的健康检查
	return nil
}
