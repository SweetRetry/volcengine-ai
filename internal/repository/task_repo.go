package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/models"
)

// TaskRepositoryImpl 任务仓储实现
type TaskRepositoryImpl struct {
	database   *mongo.Database
	collection *mongo.Collection
}

// NewTaskRepository 创建任务仓储
func NewTaskRepository(database *mongo.Database) TaskRepository {
	return &TaskRepositoryImpl{
		database:   database,
		collection: database.Collection("tasks"),
	}
}

// CreateTask 创建任务
func (r *TaskRepositoryImpl) CreateTask(ctx context.Context, task *models.Task) error {
	_, err := r.collection.InsertOne(ctx, task)
	return err
}

// GetTaskByID 根据ID获取任务
func (r *TaskRepositoryImpl) GetTaskByID(ctx context.Context, id string) (*models.Task, error) {
	var task models.Task
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTasksByUserID 获取用户任务列表
func (r *TaskRepositoryImpl) GetTasksByUserID(ctx context.Context, userID string, taskType string, limit, offset int) ([]*models.Task, error) {
	filter := bson.M{"user_id": userID}
	if taskType != "" {
		filter["type"] = taskType
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
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

// UpdateTaskStatus 更新任务状态
func (r *TaskRepositoryImpl) UpdateTaskStatus(ctx context.Context, id, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":  status,
			"updated": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// UpdateTaskResult 更新任务结果
func (r *TaskRepositoryImpl) UpdateTaskResult(ctx context.Context, taskID string, task *models.Task, resultURL string) error {
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

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": taskID}, update)
	return err
}

// UpdateTaskError 更新任务错误
func (r *TaskRepositoryImpl) UpdateTaskError(ctx context.Context, id, errorMsg string) error {
	update := bson.M{
		"$set": bson.M{
			"status":  config.TaskStatusFailed,
			"error":   errorMsg,
			"updated": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// DeleteTask 删除任务
func (r *TaskRepositoryImpl) DeleteTask(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// CreateTaskIndexes 创建任务索引
func (r *TaskRepositoryImpl) CreateTaskIndexes(ctx context.Context) error {
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

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
