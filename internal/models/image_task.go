package models

import (
	"time"
)

// ImageTask 图像任务数据模型
type ImageTask struct {
	ID       string    `json:"id" bson:"_id,omitempty"`
	UserID   string    `json:"user_id" bson:"user_id"`
	Prompt   string    `json:"prompt" bson:"prompt"`
	Model    string    `json:"model" bson:"model"`
	Size     string    `json:"size" bson:"size"`
	N        int       `json:"n" bson:"n"`
	Status   string    `json:"status" bson:"status"`       // pending, processing, completed, failed
	ImageURL string    `json:"image_url" bson:"image_url"` // 生成的图像URL
	Error    string    `json:"error" bson:"error"`         // 错误信息
	Created  time.Time `json:"created" bson:"created"`
	Updated  time.Time `json:"updated" bson:"updated"`
}
