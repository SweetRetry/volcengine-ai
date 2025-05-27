package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/volcengine/volc-sdk-golang/service/visual"

	"jimeng-go-server/internal/config"
)

// VolcengineAIService 火山引擎即梦AI服务
type VolcengineAIService struct {
	config       config.AIConfig
	visualClient *visual.Visual
	logger       *logrus.Logger
}

// 即梦AI图像生成请求结构（根据官方文档）
type JimengImageRequest struct {
	Prompt    string `json:"prompt"`              // 必填：文本描述
	Width     int    `json:"width,omitempty"`     // 图像宽度，默认512，范围[256, 768]
	Height    int    `json:"height,omitempty"`    // 图像高度，默认512，范围[256, 768]
	Seed      int64  `json:"seed,omitempty"`      // 随机种子，默认-1（随机）
	UsePreLLM bool   `json:"use_pre_llm"`         // 开启文本扩写，默认true
	UseSR     bool   `json:"use_sr"`              // 开启AIGC超分，默认true
	ReturnURL bool   `json:"return_url"`          // 返回图片链接，默认true
	LogoInfo  string `json:"logo_info,omitempty"` // 水印信息
}

// 即梦AI图像生成响应结构
type JimengImageResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    JimengImageData `json:"data"`
	Status  int             `json:"status"`
	Time    int64           `json:"time"`
}

type JimengImageData struct {
	PrimaryImageUrl string   `json:"primary_image_url"` // 主图URL
	ImageUrls       []string `json:"image_urls"`        // 所有图片URL列表
	ImageBase64     []string `json:"image_base64"`      // Base64编码的图片数据
	Seed            int64    `json:"seed"`              // 使用的随机种子
}

// NewVolcengineAIService 创建火山引擎AI服务实例
func NewVolcengineAIService(cfg config.AIConfig) *VolcengineAIService {
	// 创建火山引擎Visual服务客户端
	visualClient := visual.NewInstance()

	// 设置认证信息 - 使用正确的SDK方法
	visualClient.Client.SetAccessKey(cfg.VolcengineAccessKey)
	visualClient.Client.SetSecretKey(cfg.VolcengineSecretKey)

	return &VolcengineAIService{
		config:       cfg,
		visualClient: visualClient,
		logger:       logrus.New(),
	}
}

// SubmitImageTask 提交图像生成任务（异步）
func (s *VolcengineAIService) SubmitImageTask(ctx context.Context, request *JimengImageRequest) (string, error) {
	s.logger.Infof("提交即梦AI图像生成任务: prompt=%s", request.Prompt)

	// 构建即梦AI任务参数 - 根据官方文档
	taskParams := map[string]interface{}{
		"req_key":     "jimeng_high_aes_general_v21_L", // 即梦AI服务标识
		"prompt":      request.Prompt,
		"width":       request.Width,
		"height":      request.Height,
		"use_pre_llm": request.UsePreLLM, // 开启文本扩写
		"use_sr":      request.UseSR,     // 开启AIGC超分
		"return_url":  request.ReturnURL, // 返回图片链接
	}

	// 添加可选参数
	if request.Seed > 0 {
		taskParams["seed"] = request.Seed
	} else {
		taskParams["seed"] = -1 // 默认随机种子
	}

	if request.LogoInfo != "" {
		taskParams["logo_info"] = request.LogoInfo
	}

	s.logger.Infof("提交任务参数: %v", taskParams)
	// 调用CVSubmitTask提交任务
	resp, status, err := s.visualClient.CVSubmitTask(taskParams)
	if err != nil {
		s.logger.Errorf("提交即梦AI任务失败: %v", err)
		return "", fmt.Errorf("提交即梦AI任务失败: %v", err)
	}

	s.logger.Infof("提交任务响应: %v", resp)
	s.logger.Infof("提交任务响应状态: %d", status)

	// 解析响应获取task_id
	if data, exists := resp["data"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if taskID, exists := dataMap["task_id"]; exists {
				if taskIDStr, ok := taskID.(string); ok {
					s.logger.Infof("即梦AI任务提交成功，task_id: %s", taskIDStr)
					return taskIDStr, nil
				}
			}
		}
	}

	return "", fmt.Errorf("无法获取task_id，响应: %v", resp)
}

// GetImageTaskResult 查询图像生成任务结果（异步）
func (s *VolcengineAIService) GetImageTaskResult(ctx context.Context, taskID string) (*JimengImageResponse, error) {
	s.logger.Infof("查询即梦AI任务结果: task_id=%s", taskID)

	// 构建查询参数
	queryParams := map[string]interface{}{
		"req_key": "jimeng_high_aes_general_v21_L", // 即梦AI服务标识
		"task_id": taskID,
	}

	// 调用CVGetResult查询任务结果
	resp, status, err := s.visualClient.CVGetResult(queryParams)
	if err != nil {
		s.logger.Errorf("查询即梦AI任务结果失败: %v", err)
		return nil, fmt.Errorf("查询即梦AI任务结果失败: %v", err)
	}

	s.logger.Infof("查询任务结果响应状态: %d", status)

	// 解析响应
	response := &JimengImageResponse{}
	respMap := resp

	// 解析响应码
	if code, exists := respMap["code"]; exists {
		if codeInt, ok := code.(int); ok {
			response.Code = codeInt
		} else if codeFloat, ok := code.(float64); ok {
			response.Code = int(codeFloat)
		}
	}

	// 解析消息
	if message, exists := respMap["message"]; exists {
		if msgStr, ok := message.(string); ok {
			response.Message = msgStr
		}
	}

	// 解析数据
	if data, exists := respMap["data"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			response.Data = s.parseImageData(dataMap)
		}
	}

	response.Status = status
	response.Time = time.Now().Unix()

	// 检查API响应状态
	if response.Code != 0 {
		return response, fmt.Errorf("即梦AI任务失败: %s (code: %d)", response.Message, response.Code)
	}

	s.logger.Infof("即梦AI任务查询成功，状态: %s", response.Message)
	return response, nil
}

// parseImageData 解析图像数据
func (s *VolcengineAIService) parseImageData(dataMap map[string]interface{}) JimengImageData {
	data := JimengImageData{}

	// 解析主图URL
	if primaryUrl, exists := dataMap["primary_image_url"]; exists {
		if urlStr, ok := primaryUrl.(string); ok {
			data.PrimaryImageUrl = urlStr
		}
	}

	// 解析图片URL列表
	if imageUrls, exists := dataMap["image_urls"]; exists {
		if urlsSlice, ok := imageUrls.([]interface{}); ok {
			for _, url := range urlsSlice {
				if urlStr, ok := url.(string); ok {
					data.ImageUrls = append(data.ImageUrls, urlStr)
				}
			}
		}
	}

	// 解析Base64图片数据
	if imageBase64, exists := dataMap["image_base64"]; exists {
		if base64Slice, ok := imageBase64.([]interface{}); ok {
			for _, base64 := range base64Slice {
				if base64Str, ok := base64.(string); ok {
					data.ImageBase64 = append(data.ImageBase64, base64Str)
				}
			}
		}
	}

	// 解析种子值
	if seed, exists := dataMap["seed"]; exists {
		if seedInt, ok := seed.(int64); ok {
			data.Seed = seedInt
		} else if seedFloat, ok := seed.(float64); ok {
			data.Seed = int64(seedFloat)
		}
	}

	return data
}

// parseSize 解析尺寸字符串，范围[256, 768]
func (s *VolcengineAIService) parseSize(size string) (width, height int) {
	switch size {
	case "256x256":
		return 256, 256
	case "512x512":
		return 512, 512
	case "768x768":
		return 768, 768
	case "512x768":
		return 512, 768
	case "768x512":
		return 768, 512
	default:
		return 512, 512 // 默认尺寸
	}
}

// calculateImageCost 计算图像生成成本
func (s *VolcengineAIService) calculateImageCost(model string) float64 {
	// 根据不同模型计算成本
	switch model {
	case "jimeng-1.4":
		return 0.02 // 假设每张图片0.02元
	case "jimeng-1.0":
		return 0.015
	default:
		return 0.02
	}
}

// HealthCheck 健康检查
func (s *VolcengineAIService) HealthCheck(ctx context.Context) error {
	// 简单的健康检查
	return nil
}
