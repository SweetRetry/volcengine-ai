package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"

	"jimeng-go-server/internal/config"
)

// VolcengineAIService 火山方舟AI服务
type VolcengineAIService struct {
	config config.AIConfig
	client *arkruntime.Client
	logger *logrus.Logger
}

// 图像生成请求结构
type ImageRequest struct {
	Prompt string `json:"prompt"`          // 必填：文本描述
	Model  string `json:"model,omitempty"` // 模型ID，默认使用豆包图像生成模型
	Size   string `json:"size,omitempty"`  // 图像尺寸，如"1024x1024"
	N      int    `json:"n,omitempty"`     // 生成图片数量，默认1
}

// 图像生成响应结构
type ImageResponse struct {
	Data    []ImageData `json:"data"`
	Created int64       `json:"created"`
}

type ImageData struct {
	URL string `json:"url"` // 图片URL
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

	return &VolcengineAIService{
		config: cfg,
		client: client,
		logger: logrus.New(),
	}
}

// GenerateImage 生成图像（同步）
func (s *VolcengineAIService) GenerateImage(ctx context.Context, request *ImageRequest) (*ImageResponse, error) {
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

	generateReq := model.GenerateImagesRequest{
		Model:  modelID,
		Prompt: request.Prompt,
		Size:   &size,
	}

	// 调用火山方舟图像生成API
	imagesResponse, err := s.client.GenerateImages(ctx, generateReq)
	if err != nil {
		s.logger.Errorf("图像生成失败: %v", err)
		return nil, fmt.Errorf("图像生成失败: %v", err)
	}

	// 转换响应格式
	response := &ImageResponse{
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

// HealthCheck 健康检查
func (s *VolcengineAIService) HealthCheck(ctx context.Context) error {
	// 简单的健康检查
	return nil
}
