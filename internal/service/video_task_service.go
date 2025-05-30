package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"volcengine-go-server/config"
	"volcengine-go-server/internal/models"
)

// VideoTaskInput 视频任务输入
type VideoTaskInput struct {
	Prompt      string `json:"prompt" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
	Model       string `json:"model"`
	ReqKey      string `json:"req_key"`
	Seed        int64  `json:"seed"`
	AspectRatio string `json:"aspect_ratio"`
}

// VideoTaskResult 视频任务结果
type VideoTaskResult struct {
	ID       string    `json:"id"`
	Status   string    `json:"status"`
	VideoURL string    `json:"video_url"`
	Error    string    `json:"error"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

// VideoTaskService 视频任务服务
type VideoTaskService struct {
	collection *mongo.Collection
}

// NewVideoTaskService 创建视频任务服务
func NewVideoTaskService(db *mongo.Database) *VideoTaskService {
	return &VideoTaskService{
		collection: db.Collection("video_tasks"),
	}
}

// CreateVideoTask 创建视频任务
func (s *VideoTaskService) CreateVideoTask(ctx context.Context, input *VideoTaskInput) (*models.VideoTask, error) {
	// 设置默认值
	if input.ReqKey == "" {
		input.ReqKey = config.VolcengineJimengVideoModel
	}
	if input.Seed == 0 {
		input.Seed = config.DefaultVideoSeed
	}
	if input.AspectRatio == "" {
		input.AspectRatio = config.DefaultVideoAspectRatio
	}

	task := &models.VideoTask{
		ID:          primitive.NewObjectID().Hex(),
		UserID:      input.UserID,
		Prompt:      input.Prompt,
		Model:       input.Model,
		ReqKey:      input.ReqKey,
		Seed:        input.Seed,
		AspectRatio: input.AspectRatio,
		Status:      config.TaskStatusPending,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	_, err := s.collection.InsertOne(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetVideoTask 获取视频任务
func (s *VideoTaskService) GetVideoTask(ctx context.Context, taskID string) (*VideoTaskResult, error) {
	var task models.VideoTask
	err := s.collection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &VideoTaskResult{
		ID:       task.ID,
		Status:   task.Status,
		VideoURL: task.VideoURL,
		Error:    task.Error,
		Created:  task.Created,
		Updated:  task.Updated,
	}, nil
}

// UpdateVideoTaskStatus 更新视频任务状态
func (s *VideoTaskService) UpdateVideoTaskStatus(ctx context.Context, taskID, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":  status,
			"updated": time.Now(),
		},
	}
	_, err := s.collection.UpdateOne(ctx, bson.M{"_id": taskID}, update)
	return err
}

// UpdateVideoTaskResult 更新视频任务结果
func (s *VideoTaskService) UpdateVideoTaskResult(ctx context.Context, taskID, videoURL string) error {
	update := bson.M{
		"$set": bson.M{
			"status":    config.TaskStatusCompleted,
			"video_url": videoURL,
			"updated":   time.Now(),
		},
	}
	_, err := s.collection.UpdateOne(ctx, bson.M{"_id": taskID}, update)
	return err
}

// UpdateVideoTaskError 更新视频任务错误
func (s *VideoTaskService) UpdateVideoTaskError(ctx context.Context, taskID, errorMsg string) error {
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

// GetUserVideoTasks 获取用户视频任务列表
func (s *VideoTaskService) GetUserVideoTasks(ctx context.Context, userID string, limit, offset int) ([]*models.VideoTask, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []*models.VideoTask
	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// DeleteVideoTask 删除视频任务
func (s *VideoTaskService) DeleteVideoTask(ctx context.Context, taskID string) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{"_id": taskID})
	return err
}
