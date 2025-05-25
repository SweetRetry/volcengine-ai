package database

import (
	"context"
	"time"
)

// Database 数据库接口
type Database interface {
	// 用户相关
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error

	// 任务相关
	CreateTask(ctx context.Context, task *Task) error
	GetTaskByID(ctx context.Context, id string) (*Task, error)
	GetTasksByUserID(ctx context.Context, userID string, limit, offset int) ([]*Task, error)
	UpdateTask(ctx context.Context, task *Task) error
	DeleteTask(ctx context.Context, id string) error

	// AI请求记录相关
	CreateAIRequest(ctx context.Context, request *AIRequest) error
	GetAIRequestByID(ctx context.Context, id string) (*AIRequest, error)
	GetAIRequestsByUserID(ctx context.Context, userID string, limit, offset int) ([]*AIRequest, error)

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

// Task 任务模型
type Task struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"`        // AI生成类型
	Status      string    `json:"status"`      // pending, processing, completed, failed
	Input       string    `json:"input"`       // 输入参数
	Output      string    `json:"output"`      // 输出结果
	ErrorMsg    string    `json:"error_msg"`   // 错误信息
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// AIRequest AI请求记录
type AIRequest struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	TaskID     string    `json:"task_id"`
	Provider   string    `json:"provider"`   // 服务提供商，如"jimeng"
	Model      string    `json:"model"`      // 模型名称
	Prompt     string    `json:"prompt"`     // 请求提示词
	Response   string    `json:"response"`   // 响应结果
	Tokens     int       `json:"tokens"`     // 消耗的token数量
	Cost       float64   `json:"cost"`       // 消耗费用
	Duration   int64     `json:"duration"`   // 请求耗时(毫秒)
	CreatedAt  time.Time `json:"created_at"`
} 