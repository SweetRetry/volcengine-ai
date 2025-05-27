package database

import (
	"context"
	"time"
)

// 简化的数据库接口 - 只保留实际使用的方法
type Database interface {
	// 用户相关 - 只保留核心方法
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error

	// 图像任务相关 - 当前主要使用的功能
	CreateImageTask(ctx context.Context, task *ImageTask) error
	GetImageTaskByID(ctx context.Context, id string) (*ImageTask, error)
	GetImageTasksByUserID(ctx context.Context, userID string, limit, offset int) ([]*ImageTask, error)
	UpdateImageTaskStatus(ctx context.Context, id, status, imageURL, errorMsg string) error
	DeleteImageTask(ctx context.Context, id string) error

	// 关闭连接
	Close() error
}

// User 用户模型
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 简化的图像任务模型
type ImageTask struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Prompt   string    `json:"prompt"`
	Model    string    `json:"model"`
	Size     string    `json:"size"`
	N        int       `json:"n"`
	Status   string    `json:"status"`    // pending, processing, completed, failed
	ImageURL string    `json:"image_url"` // 生成的图像URL
	Error    string    `json:"error"`     // 错误信息
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}
