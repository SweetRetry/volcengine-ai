package models

import (
	"time"
)

// VideoTask 视频任务数据模型
type VideoTask struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	UserID      string    `json:"user_id" bson:"user_id"`
	Prompt      string    `json:"prompt" bson:"prompt"`
	Model       string    `json:"model" bson:"model"`
	ReqKey      string    `json:"req_key" bson:"req_key"`           // 服务标识
	Seed        int64     `json:"seed" bson:"seed"`                 // 随机种子
	AspectRatio string    `json:"aspect_ratio" bson:"aspect_ratio"` // 视频尺寸比例
	Status      string    `json:"status" bson:"status"`             // pending, processing, completed, failed
	VideoURL    string    `json:"video_url" bson:"video_url"`       // 生成的视频URL
	Error       string    `json:"error" bson:"error"`               // 错误信息
	Created     time.Time `json:"created" bson:"created"`
	Updated     time.Time `json:"updated" bson:"updated"`
}
