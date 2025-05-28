package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"volcengine-go-server/internal/models"
)

// MongoDB 数据库连接管理器，组合各个repository
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database

	// 组合各个repository
	userRepo      UserRepository
	imageTaskRepo ImageTaskRepository
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
	imageTaskRepo := NewImageTaskRepository(database)

	// 创建索引
	if err := userRepo.CreateUserIndexes(context.Background()); err != nil {
		return nil, err
	}
	if err := imageTaskRepo.CreateImageTaskIndexes(context.Background()); err != nil {
		return nil, err
	}

	return &MongoDB{
		client:        client,
		database:      database,
		userRepo:      userRepo,
		imageTaskRepo: imageTaskRepo,
	}, nil
}

// 实现Database接口 - 用户相关方法
func (m *MongoDB) CreateUser(ctx context.Context, user *models.User) error {
	return m.userRepo.CreateUser(ctx, user)
}

func (m *MongoDB) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return m.userRepo.GetUserByID(ctx, id)
}

func (m *MongoDB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.userRepo.GetUserByEmail(ctx, email)
}

func (m *MongoDB) UpdateUser(ctx context.Context, user *models.User) error {
	return m.userRepo.UpdateUser(ctx, user)
}

func (m *MongoDB) DeleteUser(ctx context.Context, id string) error {
	return m.userRepo.DeleteUser(ctx, id)
}

func (m *MongoDB) CreateUserIndexes(ctx context.Context) error {
	return m.userRepo.CreateUserIndexes(ctx)
}

// 实现Database接口 - 图像任务相关方法
func (m *MongoDB) CreateImageTask(ctx context.Context, task *models.ImageTask) error {
	return m.imageTaskRepo.CreateImageTask(ctx, task)
}

func (m *MongoDB) GetImageTaskByID(ctx context.Context, id string) (*models.ImageTask, error) {
	return m.imageTaskRepo.GetImageTaskByID(ctx, id)
}

func (m *MongoDB) GetImageTasksByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.ImageTask, error) {
	return m.imageTaskRepo.GetImageTasksByUserID(ctx, userID, limit, offset)
}

func (m *MongoDB) UpdateImageTaskStatus(ctx context.Context, id, status, imageURL, errorMsg string) error {
	return m.imageTaskRepo.UpdateImageTaskStatus(ctx, id, status, imageURL, errorMsg)
}

func (m *MongoDB) DeleteImageTask(ctx context.Context, id string) error {
	return m.imageTaskRepo.DeleteImageTask(ctx, id)
}

func (m *MongoDB) CreateImageTaskIndexes(ctx context.Context) error {
	return m.imageTaskRepo.CreateImageTaskIndexes(ctx)
}

// Close 关闭数据库连接
func (m *MongoDB) Close() error {
	return m.client.Disconnect(context.Background())
}
