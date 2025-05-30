package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/models"
)

// TaskInput 统一任务输入
type TaskInput struct {
	Prompt   string `json:"prompt" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
	Type     string `json:"type" binding:"required"` // image, video, text
	Model    string `json:"model"`
	Provider string `json:"provider"`

	// 图像生成特有字段
	Size string `json:"size,omitempty"`
	N    int    `json:"n,omitempty"`

	// 视频生成特有字段
	ReqKey      string `json:"req_key,omitempty"`
	Seed        int64  `json:"seed,omitempty"`
	AspectRatio string `json:"aspect_ratio,omitempty"`

	// 文本生成特有字段
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// TaskService 统一任务服务
type TaskService struct {
	collection *mongo.Collection
}

// NewTaskService 创建任务服务
func NewTaskService(db *mongo.Database) *TaskService {
	return &TaskService{
		collection: db.Collection("tasks"),
	}
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(ctx context.Context, input *TaskInput) (*models.Task, error) {
	task := &models.Task{
		ID:       primitive.NewObjectID().Hex(),
		UserID:   input.UserID,
		Type:     input.Type,
		Prompt:   input.Prompt,
		Model:    input.Model,
		Provider: input.Provider,
		Status:   config.TaskStatusPending,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	// 根据任务类型设置特有字段和默认值
	switch input.Type {
	case models.TaskTypeImage:
		task.Size = input.Size
		task.N = input.N
		if task.N == 0 {
			task.N = 1
		}
	case models.TaskTypeVideo:
		task.ReqKey = input.ReqKey
		task.Seed = input.Seed
		task.AspectRatio = input.AspectRatio
		// 设置默认值
		if task.ReqKey == "" {
			task.ReqKey = config.VolcengineJimengVideoModel
		}
		if task.Seed == 0 {
			task.Seed = config.DefaultVideoSeed
		}
		if task.AspectRatio == "" {
			task.AspectRatio = config.DefaultVideoAspectRatio
		}
	case models.TaskTypeText:
		task.MaxTokens = input.MaxTokens
		task.Temperature = input.Temperature
	}

	_, err := s.collection.InsertOne(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask 获取任务
func (s *TaskService) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	var task models.Task
	err := s.collection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// UpdateTaskStatus 更新任务状态
func (s *TaskService) UpdateTaskStatus(ctx context.Context, taskID, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":  status,
			"updated": time.Now(),
		},
	}
	_, err := s.collection.UpdateOne(ctx, bson.M{"_id": taskID}, update)
	return err
}

// UpdateTaskResult 更新任务结果
func (s *TaskService) UpdateTaskResult(ctx context.Context, taskID, resultURL string) error {
	// 首先获取任务以确定类型
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":  config.TaskStatusCompleted,
			"updated": time.Now(),
		},
	}

	// 根据任务类型设置相应的结果字段
	switch task.Type {
	case models.TaskTypeImage:
		update["$set"].(bson.M)["image_url"] = resultURL
	case models.TaskTypeVideo:
		update["$set"].(bson.M)["video_url"] = resultURL
	case models.TaskTypeText:
		update["$set"].(bson.M)["text_result"] = resultURL
	}

	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": taskID}, update)
	return err
}

// UpdateTaskError 更新任务错误
func (s *TaskService) UpdateTaskError(ctx context.Context, taskID, errorMsg string) error {
	update := bson.M{
		"$set": bson.M{
			"status":  config.TaskStatusFailed,
			"error":   errorMsg,
			"updated": time.Now(),
		},
	}
	_, err := s.collection.UpdateOne(ctx, bson.M{"_id": taskID}, update)
	return err
}

// GetUserTasks 获取用户任务列表
func (s *TaskService) GetUserTasks(ctx context.Context, userID string, taskType string, limit, offset int) ([]*models.Task, error) {
	filter := bson.M{"user_id": userID}
	if taskType != "" {
		filter["type"] = taskType
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []*models.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(ctx context.Context, taskID string) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{"_id": taskID})
	return err
}

// CreateTaskIndexes 创建任务索引
func (s *TaskService) CreateTaskIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "type", Value: 1},
			},
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
