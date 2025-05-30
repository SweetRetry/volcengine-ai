package repository

import (
	"context"

	"volcengine-go-server/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
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

// ImageTaskRepository 图像任务数据访问接口
type ImageTaskRepository interface {
	CreateImageTask(ctx context.Context, task *models.ImageTask) error
	GetImageTaskByID(ctx context.Context, id string) (*models.ImageTask, error)
	GetImageTasksByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.ImageTask, error)
	UpdateImageTaskStatus(ctx context.Context, id, status, imageURL, errorMsg string) error
	DeleteImageTask(ctx context.Context, id string) error
	CreateImageTaskIndexes(ctx context.Context) error
}

// Database 数据库接口 - 组合所有repository接口
type Database interface {
	UserRepository
	ImageTaskRepository
	// 获取底层的mongo.Database实例
	GetDatabase() *mongo.Database
	// 关闭连接
	Close() error
}
