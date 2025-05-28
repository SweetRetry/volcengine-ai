package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"volcengine-go-server/internal/models"
)

// ImageTaskRepositoryImpl MongoDB图像任务repository实现
type ImageTaskRepositoryImpl struct {
	database *mongo.Database
}

// NewImageTaskRepository 创建图像任务repository实例
func NewImageTaskRepository(database *mongo.Database) ImageTaskRepository {
	return &ImageTaskRepositoryImpl{database: database}
}

// CreateImageTask 创建图像任务
func (r *ImageTaskRepositoryImpl) CreateImageTask(ctx context.Context, task *models.ImageTask) error {
	task.Created = time.Now()
	task.Updated = time.Now()

	collection := r.database.Collection("image_tasks")
	result, err := collection.InsertOne(ctx, task)
	if err != nil {
		return err
	}

	task.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

// GetImageTaskByID 根据ID获取图像任务
func (r *ImageTaskRepositoryImpl) GetImageTaskByID(ctx context.Context, id string) (*models.ImageTask, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var task models.ImageTask
	collection := r.database.Collection("image_tasks")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&task)
	if err != nil {
		return nil, err
	}

	task.ID = objectID.Hex()
	return &task, nil
}

// GetImageTasksByUserID 根据用户ID获取图像任务列表
func (r *ImageTaskRepositoryImpl) GetImageTasksByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.ImageTask, error) {
	collection := r.database.Collection("image_tasks")

	opts := options.Find().
		SetSort(bson.D{{Key: "created", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	tasks := make([]*models.ImageTask, len(results))
	for i, result := range results {
		task := &models.ImageTask{
			UserID:   result["user_id"].(string),
			Prompt:   result["prompt"].(string),
			Model:    result["model"].(string),
			Size:     result["size"].(string),
			N:        int(result["n"].(int32)),
			Status:   result["status"].(string),
			ImageURL: result["image_url"].(string),
			Error:    result["error"].(string),
			Created:  result["created"].(primitive.DateTime).Time(),
			Updated:  result["updated"].(primitive.DateTime).Time(),
		}

		if objectID, ok := result["_id"].(primitive.ObjectID); ok {
			task.ID = objectID.Hex()
		}

		tasks[i] = task
	}

	return tasks, nil
}

// UpdateImageTaskStatus 更新图像任务状态
func (r *ImageTaskRepositoryImpl) UpdateImageTaskStatus(ctx context.Context, id, status, imageURL, errorMsg string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"image_url": imageURL,
			"error":     errorMsg,
			"updated":   time.Now(),
		},
	}

	collection := r.database.Collection("image_tasks")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// DeleteImageTask 删除图像任务
func (r *ImageTaskRepositoryImpl) DeleteImageTask(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	collection := r.database.Collection("image_tasks")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// CreateImageTaskIndexes 创建图像任务相关的索引
func (r *ImageTaskRepositoryImpl) CreateImageTaskIndexes(ctx context.Context) error {
	imageTaskCollection := r.database.Collection("image_tasks")
	_, err := imageTaskCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
	return err
}
