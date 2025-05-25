package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"jimeng-go-server/internal/config"
)

// AIResponse AI响应结果
type AIResponse struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Content    string                 `json:"content"`
	ImageURL   string                 `json:"image_url,omitempty"`
	TokensUsed int                    `json:"tokens_used"`
	Model      string                 `json:"model"`
	Provider   string                 `json:"provider"`
	Duration   time.Duration          `json:"duration"`
	Cost       float64                `json:"cost"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// AsyncTaskResponse 异步任务响应
type AsyncTaskResponse struct {
	TaskID   string `json:"task_id"`
	Status   string `json:"status"` // pending, processing, completed, failed
	Message  string `json:"message"`
	Provider string `json:"provider"`
}

// VolcengineAIService 火山引擎即梦AI服务
type VolcengineAIService struct {
	config     config.AIConfig
	httpClient *http.Client
	logger     *logrus.Logger
}

// 火山方舟图像生成请求结构
type VolcengineImageRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	N       int    `json:"n,omitempty"`
	Size    string `json:"size,omitempty"`
	Quality string `json:"quality,omitempty"`
	Style   string `json:"style,omitempty"`
	User    string `json:"user,omitempty"`
}

// 火山方舟图像生成响应结构
type VolcengineImageResponse struct {
	Created int64                 `json:"created"`
	Data    []VolcengineImageData `json:"data"`
	Error   *VolcengineError      `json:"error,omitempty"`
}

type VolcengineImageData struct {
	URL           string `json:"url"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type VolcengineError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// NewVolcengineAIService 创建火山引擎AI服务实例
func NewVolcengineAIService(cfg config.AIConfig) *VolcengineAIService {
	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		timeout = 30 * time.Second
	}

	return &VolcengineAIService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logrus.New(),
	}
}

// CreateImageTask 创建异步图像生成任务
func (s *VolcengineAIService) CreateImageTask(ctx context.Context, prompt string, options map[string]interface{}) (*AsyncTaskResponse, error) {
	// 生成任务ID
	taskID := fmt.Sprintf("volcengine_img_%d", time.Now().UnixNano())

	s.logger.Infof("创建火山引擎图像生成任务: %s, prompt: %s", taskID, prompt)

	// 构建请求参数
	request := &VolcengineImageRequest{
		Model:   "doubao-seedream-3.0-t2i", // 使用豆包图像生成模型
		Prompt:  prompt,
		N:       1,
		Size:    "1024x1024",
		Quality: "standard",
	}

	// 应用自定义选项
	if options != nil {
		if model, ok := options["model"].(string); ok && model != "" {
			request.Model = model
		}
		if n, ok := options["n"].(int); ok && n > 0 {
			request.N = n
		}
		if size, ok := options["size"].(string); ok && size != "" {
			request.Size = size
		}
		if quality, ok := options["quality"].(string); ok && quality != "" {
			request.Quality = quality
		}
		if style, ok := options["style"].(string); ok && style != "" {
			request.Style = style
		}
	}

	// 这里我们模拟异步处理，实际项目中应该将任务放入队列
	// 返回任务ID，让客户端通过taskID查询结果
	return &AsyncTaskResponse{
		TaskID:   taskID,
		Status:   "pending",
		Message:  "任务已创建，正在处理中",
		Provider: "volcengine_jimeng",
	}, nil
}

// GetTaskResult 通过taskID获取任务结果
func (s *VolcengineAIService) GetTaskResult(ctx context.Context, taskID string) (*AIResponse, error) {
	s.logger.Infof("查询任务结果: %s", taskID)

	// 这里应该从数据库或缓存中查询任务状态
	// 为了演示，我们模拟一个处理流程

	// 模拟任务处理时间
	// 实际项目中，这里应该查询数据库中的任务状态

	// 如果任务还在处理中
	if time.Now().Unix()%3 == 0 { // 模拟33%的概率任务还在处理
		return &AIResponse{
			ID:       taskID,
			Type:     "image",
			Provider: "volcengine_jimeng",
			Error:    "任务处理中，请稍后查询",
		}, nil
	}

	// 在开发环境下模拟成功响应，避免调用真实API
	if s.config.VolcengineAPIKey == "WkRreU1HUmxNRFUzTXpJM05EQTBNVGsxTldVNE9XUmtaV0ZpWm1VeE0yWQ==" {
		// 模拟成功的响应
		return &AIResponse{
			ID:       taskID,
			Type:     "image",
			ImageURL: "https://example.com/generated-image-" + taskID + ".jpg",
			Provider: "volcengine_jimeng",
			Duration: time.Since(time.Now().Add(-5 * time.Second)), // 模拟处理时间
			Cost:     s.calculateImageCost("doubao-seedream-3.0-t2i"),
			Metadata: map[string]interface{}{
				"task_id": taskID,
				"model":   "doubao-seedream-3.0-t2i",
				"prompt":  "模拟生成的图像",
				"size":    "1024x1024",
				"quality": "standard",
				"created": time.Now().Unix(),
				"data": []map[string]interface{}{
					{
						"url":            "https://example.com/generated-image-" + taskID + ".jpg",
						"revised_prompt": "A simulated generated image",
					},
				},
			},
		}, nil
	}

	// 模拟调用实际的火山引擎API
	prompt := "一只可爱的小猫咪在花园里玩耍" // 实际应该从任务记录中获取
	resp, err := s.callVolcengineImageAPI(ctx, &VolcengineImageRequest{
		Model:   "doubao-seedream-3.0-t2i",
		Prompt:  prompt,
		N:       1,
		Size:    "1024x1024",
		Quality: "standard",
	})
	if err != nil {
		return &AIResponse{
			ID:       taskID,
			Type:     "image",
			Provider: "volcengine_jimeng",
			Error:    err.Error(),
		}, err
	}

	if resp.Error != nil {
		return &AIResponse{
			ID:       taskID,
			Type:     "image",
			Provider: "volcengine_jimeng",
			Error:    resp.Error.Message,
		}, fmt.Errorf("火山引擎API错误: %s", resp.Error.Message)
	}

	// 获取第一张图片URL
	imageURL := ""
	if len(resp.Data) > 0 {
		imageURL = resp.Data[0].URL
	}

	return &AIResponse{
		ID:       taskID,
		Type:     "image",
		ImageURL: imageURL,
		Provider: "volcengine_jimeng",
		Duration: time.Since(time.Now().Add(-5 * time.Second)), // 模拟处理时间
		Cost:     s.calculateImageCost("doubao-seedream-3.0-t2i"),
		Metadata: map[string]interface{}{
			"task_id": taskID,
			"model":   "doubao-seedream-3.0-t2i",
			"prompt":  prompt,
			"size":    "1024x1024",
			"quality": "standard",
			"created": resp.Created,
			"data":    resp.Data,
		},
	}, nil
}

// callVolcengineImageAPI 调用火山方舟图像生成API
func (s *VolcengineAIService) callVolcengineImageAPI(ctx context.Context, request *VolcengineImageRequest) (*VolcengineImageResponse, error) {
	// 序列化请求体
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 构建API URL - 使用火山方舟的图像生成接口
	apiURL := "https://ark.cn-beijing.volces.com/api/v3/images/generations"

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.VolcengineAPIKey))

	s.logger.Infof("调用火山方舟图像生成API: %s", string(requestBody))

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	s.logger.Infof("火山方舟API响应状态: %d, 响应体: %s", resp.StatusCode, string(body))

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var volcengineResp VolcengineImageResponse
	if err := json.Unmarshal(body, &volcengineResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &volcengineResp, nil
}

// calculateImageCost 计算图像生成成本
func (s *VolcengineAIService) calculateImageCost(model string) float64 {
	// 根据不同模型计算成本
	switch model {
	case "doubao-seedream-3.0-t2i":
		return 0.02 // 假设每张图片0.02元
	default:
		return 0.02
	}
}

// HealthCheck 健康检查
func (s *VolcengineAIService) HealthCheck(ctx context.Context) error {
	// 简单的健康检查
	return nil
}
