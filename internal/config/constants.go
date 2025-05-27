package config

// AI模型常量
const (
	// 火山引擎豆包模型
	VolcengineImageModel = "doubao-seedream-3-0-t2i-250415"
	VolcengineTextModel  = "doubao-pro-4k"
	VolcengineVideoModel = "doubao-video-pro"

	// OpenAI模型
	OpenAIImageModel = "dall-e-3"
	OpenAITextModel  = "gpt-4"
	OpenAIVideoModel = "sora"
)

// 图像尺寸常量
const (
	ImageSize1x1     = "1024x1024" // 1:1 比例
	ImageSize4x3     = "1024x768"  // 4:3 比例
	ImageSize3x4     = "768x1024"  // 3:4 比例
	ImageSize16x9    = "1024x576"  // 16:9 比例
	ImageSize9x16    = "576x1024"  // 9:16 比例
	DefaultImageSize = ImageSize1x1
)

// 分页常量
const (
	DefaultPageLimit  = 20
	MaxPageLimit      = 100
	DefaultPageOffset = 0
)

// 任务状态常量
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// 默认提供商
const (
	DefaultAIProvider = "volcengine"
)

// 队列配置常量
const (
	QueueConcurrency    = 10
	QueueCriticalWeight = 6
	QueueDefaultWeight  = 3
	QueueLowWeight      = 1
)
