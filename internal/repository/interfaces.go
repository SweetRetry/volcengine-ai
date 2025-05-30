package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"volcengine-go-server/internal/models"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	CreateUserIndexes(ctx context.Context) error
}

// TaskRepository 任务数据访问接口
type TaskRepository interface {
	CreateTask(ctx context.Context, task *models.Task) error
	GetTaskByID(ctx context.Context, id string) (*models.Task, error)
	GetTasksByUserID(ctx context.Context, userID string, taskType string, limit, offset int) ([]*models.Task, error)
	UpdateTaskStatus(ctx context.Context, id, status string) error
	UpdateTaskResult(ctx context.Context, taskID string, task *models.Task, resultURL string) error
	UpdateTaskError(ctx context.Context, id, errorMsg string) error
	DeleteTask(ctx context.Context, id string) error
	CreateTaskIndexes(ctx context.Context) error
}

// Database 数据库接口 - 提供Repository实例的工厂
type Database interface {
	// 获取Repository实例
	UserRepository() UserRepository
	TaskRepository() TaskRepository

	// 获取底层的mongo.Database实例
	GetDatabase() *mongo.Database
	// 关闭连接
	Close() error
}
