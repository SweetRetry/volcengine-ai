package volcengine

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

// 即梦AI图生视频请求结构
type JimengI2VRequest struct {
	ImageURLs   []string `json:"image_urls"`             // 必填：图片链接数组
	Prompt      string   `json:"prompt,omitempty"`       // 可选：生成视频的提示词，支持中英文，150字符以内
	Seed        int      `json:"seed,omitempty"`         // 可选：随机种子，默认-1（随机）
	AspectRatio string   `json:"aspect_ratio,omitempty"` // 必填：生成视频的尺寸比例
}

// 即梦AI视频生成结果
type JimengVideoResult struct {
	VideoURL string `json:"video_url"` // 视频URL
	Status   string `json:"status"`    // 任务状态
}
