package models

import (
	"time"
)

// TaskInput 统一任务输入
type TaskInput struct {
	Prompt   string `json:"prompt" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
	Type     string `json:"type" binding:"required"` // image, video, text
	Model    string `json:"model"`
	Provider string `json:"provider"`

	// 图像和视频生成共用字段
	AspectRatio string `json:"aspect_ratio,omitempty"` // 宽高比例
	N           int    `json:"n,omitempty"`            // 生成数量

	// 视频生成特有字段
	Seed int64 `json:"seed,omitempty"` // 随机种子

	// 文本生成特有字段
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// Task 统一任务数据模型
type Task struct {
	ID       string    `json:"id" bson:"_id,omitempty"`
	UserID   string    `json:"user_id" bson:"user_id"`
	Type     string    `json:"type" bson:"type"` // image, video, text
	Prompt   string    `json:"prompt" bson:"prompt"`
	Model    string    `json:"model" bson:"model"`
	Provider string    `json:"provider" bson:"provider"`
	Status   string    `json:"status" bson:"status"` // pending, processing, completed, failed
	Error    string    `json:"error" bson:"error"`   // 错误信息
	Created  time.Time `json:"created" bson:"created"`
	Updated  time.Time `json:"updated" bson:"updated"`

	// 图像和视频生成共用字段
	AspectRatio string `json:"aspect_ratio,omitempty" bson:"aspect_ratio,omitempty"` // 宽高比例

	// 图像生成特有字段
	N        int    `json:"n,omitempty" bson:"n,omitempty"`
	ImageURL string `json:"image_url,omitempty" bson:"image_url,omitempty"`

	// 视频生成特有字段
	Seed     int64  `json:"seed,omitempty" bson:"seed,omitempty"`           // 随机种子
	VideoURL string `json:"video_url,omitempty" bson:"video_url,omitempty"` // 生成的视频URL

	// 文本生成特有字段
	MaxTokens   int     `json:"max_tokens,omitempty" bson:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty" bson:"temperature,omitempty"`
	TextResult  string  `json:"text_result,omitempty" bson:"text_result,omitempty"` // 生成的文本结果
}

// TaskType 任务类型常量
const (
	TaskTypeImage = "image"
	TaskTypeVideo = "video"
	TaskTypeText  = "text"
)

// GetResultURL 根据任务类型获取结果URL
func (t *Task) GetResultURL() string {
	switch t.Type {
	case TaskTypeImage:
		return t.ImageURL
	case TaskTypeVideo:
		return t.VideoURL
	default:
		return ""
	}
}

// SetResultURL 根据任务类型设置结果URL
func (t *Task) SetResultURL(url string) {
	switch t.Type {
	case TaskTypeImage:
		t.ImageURL = url
	case TaskTypeVideo:
		t.VideoURL = url
	}
}
