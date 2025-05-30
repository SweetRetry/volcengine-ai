package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB 数据库连接管理器，提供Repository实例
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database

	// Repository实例
	userRepo UserRepository
	taskRepo TaskRepository
}

func NewMongoDB(uri string) (Database, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, err
	}

	database := client.Database("volcengine_db")

	// 创建各个repository实例
	userRepo := NewUserRepository(database)
	taskRepo := NewTaskRepository(database)

	// 创建索引
	if err := userRepo.CreateUserIndexes(context.Background()); err != nil {
		return nil, err
	}
	if err := taskRepo.CreateTaskIndexes(context.Background()); err != nil {
		return nil, err
	}

	return &MongoDB{
		client:   client,
		database: database,
		userRepo: userRepo,
		taskRepo: taskRepo,
	}, nil
}

// UserRepository 返回用户Repository实例
func (m *MongoDB) UserRepository() UserRepository {
	return m.userRepo
}

// TaskRepository 返回任务Repository实例
func (m *MongoDB) TaskRepository() TaskRepository {
	return m.taskRepo
}

// Close 关闭数据库连接
func (m *MongoDB) Close() error {
	return m.client.Disconnect(context.Background())
}

// GetDatabase 获取底层的mongo.Database实例
func (m *MongoDB) GetDatabase() *mongo.Database {
	return m.database
}
