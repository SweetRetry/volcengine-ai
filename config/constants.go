package config

// AI模型常量
const (
	// 火山引擎豆包模型
	VolcengineImageModel = "doubao-seedream-3-0-t2i-250415"
	VolcengineTextModel  = "doubao-pro-4k"
	VolcengineVideoModel = "doubao-video-pro"

	// 火山引擎即梦AI模型
	VolcengineJimengImageModel = "jimeng_high_aes_general_v21_L"

	// OpenAI模型
	OpenAIImageModel = "dall-e-3"
	OpenAITextModel  = "gpt-4"
	OpenAIVideoModel = "sora"
)

// 图像尺寸常量 - 火山引擎支持的尺寸
const (
	ImageSize1x1     = "1024x1024" // 1:1 比例
	ImageSize3x4     = "864x1152"  // 3:4 比例
	ImageSize4x3     = "1152x864"  // 4:3 比例
	ImageSize16x9    = "1280x720"  // 16:9 比例
	ImageSize9x16    = "720x1280"  // 9:16 比例
	ImageSize2x3     = "832x1248"  // 2:3 比例
	ImageSize3x2     = "1248x832"  // 3:2 比例
	ImageSize21x9    = "1512x648"  // 21:9 比例
	DefaultImageSize = ImageSize1x1
)

// 即梦AI推荐尺寸常量 - 官方建议的最佳效果尺寸
// 宽、高与512差距过大，则出图效果不佳、延迟过长概率显著增加
const (
	JimengImageSize1x1  = "512x512" // 1:1 比例
	JimengImageSize4x3  = "512x384" // 4:3 比例
	JimengImageSize3x4  = "384x512" // 3:4 比例
	JimengImageSize3x2  = "512x341" // 3:2 比例
	JimengImageSize2x3  = "341x512" // 2:3 比例
	JimengImageSize16x9 = "512x288" // 16:9 比例
	JimengImageSize9x16 = "288x512" // 9:16 比例
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
