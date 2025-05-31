package volcengine

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/disintegration/imaging"

	"volcengine-go-server/config"
)

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

// detectImageAspectRatio 检测图片的宽高比
func (s *VolcengineService) detectImageAspectRatio(ctx context.Context, imageURL string) (string, error) {
	s.logger.Infof("检测图片尺寸: %s", imageURL)

	// 获取图片尺寸
	width, height, err := s.getImageDimensions(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("获取图片尺寸失败: %v", err)
	}

	s.logger.Infof("图片尺寸: %dx%d", width, height)

	// 计算宽高比并匹配到最接近的标准比例
	aspectRatio := s.calculateBestAspectRatio(width, height)
	s.logger.Infof("匹配的标准比例: %s", aspectRatio)

	return aspectRatio, nil
}

// getImageDimensions 获取图片的宽高尺寸（使用imaging库）
func (s *VolcengineService) getImageDimensions(ctx context.Context, imageURL string) (int, int, error) {
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("User-Agent", "VolcengineAI/1.0")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return 0, 0, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 使用imaging库解码图片并获取尺寸
	img, err := imaging.Decode(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("解码图片失败: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	return width, height, nil
}

// calculateBestAspectRatio 计算最接近的标准宽高比
func (s *VolcengineService) calculateBestAspectRatio(width, height int) string {
	if width == 0 || height == 0 {
		return "16:9" // 默认比例
	}

	// 计算实际比例
	ratio := float64(width) / float64(height)

	// 定义标准比例及其数值
	standardRatios := map[string]float64{
		"1:1":  1.0,
		"4:3":  4.0 / 3.0,
		"3:4":  3.0 / 4.0,
		"16:9": 16.0 / 9.0,
		"9:16": 9.0 / 16.0,
		"21:9": 21.0 / 9.0,
		"9:21": 9.0 / 21.0,
	}

	// 找到最接近的比例
	bestRatio := "16:9"
	minDiff := math.Abs(ratio - standardRatios["16:9"])

	for ratioName, ratioValue := range standardRatios {
		diff := math.Abs(ratio - ratioValue)
		if diff < minDiff {
			minDiff = diff
			bestRatio = ratioName
		}
	}

	return bestRatio
}

// isValidAspectRatio 验证aspect_ratio是否在支持的范围内
func (s *VolcengineService) isValidAspectRatio(aspectRatio string) bool {
	validRatios := []string{"16:9", "4:3", "1:1", "3:4", "9:16", "21:9", "9:21"}
	for _, ratio := range validRatios {
		if aspectRatio == ratio {
			return true
		}
	}
	return false
}
