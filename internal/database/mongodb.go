package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

// MongoDB文档结构
type MongoUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	Name      string             `bson:"name"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type MongoTask struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UserID         string             `bson:"user_id"`
	Type           string             `bson:"type"`
	Status         string             `bson:"status"`
	Input          string             `bson:"input"`
	Output         string             `bson:"output"`
	ErrorMsg       string             `bson:"error_msg"`
	ExternalTaskID string             `bson:"external_task_id"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
	CompletedAt    *time.Time         `bson:"completed_at,omitempty"`
}

type MongoAIRequest struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    string             `bson:"user_id"`
	TaskID    string             `bson:"task_id"`
	Provider  string             `bson:"provider"`
	Model     string             `bson:"model"`
	Prompt    string             `bson:"prompt"`
	Response  string             `bson:"response"`
	Tokens    int                `bson:"tokens"`
	Cost      float64            `bson:"cost"`
	Duration  int64              `bson:"duration"`
	CreatedAt time.Time          `bson:"created_at"`
}

func NewMongoDB(uri string) (*MongoDB, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, err
	}

	database := client.Database("jimeng_db")

	// 创建索引
	m := &MongoDB{client: client, database: database}
	if err := m.createIndexes(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *MongoDB) createIndexes() error {
	ctx := context.Background()

	// 用户邮箱唯一索引
	userCollection := m.database.Collection("users")
	_, err := userCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// 任务用户ID索引
	taskCollection := m.database.Collection("tasks")
	_, err = taskCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
	if err != nil {
		return err
	}

	// AI请求用户ID和任务ID索引
	aiRequestCollection := m.database.Collection("ai_requests")
	_, err = aiRequestCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "task_id", Value: 1}}},
	})

	return err
}

// 用户相关方法
func (m *MongoDB) CreateUser(ctx context.Context, user *User) error {
	mongoUser := &MongoUser{
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := m.database.Collection("users")
	result, err := collection.InsertOne(ctx, mongoUser)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID).Hex()
	user.CreatedAt = mongoUser.CreatedAt
	user.UpdatedAt = mongoUser.UpdatedAt
	return nil
}

func (m *MongoDB) GetUserByID(ctx context.Context, id string) (*User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mongoUser MongoUser
	collection := m.database.Collection("users")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mongoUser)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        mongoUser.ID.Hex(),
		Email:     mongoUser.Email,
		Name:      mongoUser.Name,
		CreatedAt: mongoUser.CreatedAt,
		UpdatedAt: mongoUser.UpdatedAt,
	}, nil
}

func (m *MongoDB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var mongoUser MongoUser
	collection := m.database.Collection("users")
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&mongoUser)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        mongoUser.ID.Hex(),
		Email:     mongoUser.Email,
		Name:      mongoUser.Name,
		CreatedAt: mongoUser.CreatedAt,
		UpdatedAt: mongoUser.UpdatedAt,
	}, nil
}

func (m *MongoDB) UpdateUser(ctx context.Context, user *User) error {
	objectID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"name":       user.Name,
			"updated_at": time.Now(),
		},
	}

	collection := m.database.Collection("users")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (m *MongoDB) DeleteUser(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	collection := m.database.Collection("users")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// 任务相关方法
func (m *MongoDB) CreateTask(ctx context.Context, task *Task) error {
	mongoTask := &MongoTask{
		UserID:         task.UserID,
		Type:           task.Type,
		Status:         task.Status,
		Input:          task.Input,
		Output:         task.Output,
		ErrorMsg:       task.ErrorMsg,
		ExternalTaskID: task.ExternalTaskID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	collection := m.database.Collection("tasks")
	result, err := collection.InsertOne(ctx, mongoTask)
	if err != nil {
		return err
	}

	task.ID = result.InsertedID.(primitive.ObjectID).Hex()
	task.CreatedAt = mongoTask.CreatedAt
	task.UpdatedAt = mongoTask.UpdatedAt
	return nil
}

func (m *MongoDB) GetTaskByID(ctx context.Context, id string) (*Task, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mongoTask MongoTask
	collection := m.database.Collection("tasks")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mongoTask)
	if err != nil {
		return nil, err
	}

	return &Task{
		ID:             mongoTask.ID.Hex(),
		UserID:         mongoTask.UserID,
		Type:           mongoTask.Type,
		Status:         mongoTask.Status,
		Input:          mongoTask.Input,
		Output:         mongoTask.Output,
		ErrorMsg:       mongoTask.ErrorMsg,
		ExternalTaskID: mongoTask.ExternalTaskID,
		CreatedAt:      mongoTask.CreatedAt,
		UpdatedAt:      mongoTask.UpdatedAt,
		CompletedAt:    mongoTask.CompletedAt,
	}, nil
}

func (m *MongoDB) GetTasksByUserID(ctx context.Context, userID string, limit, offset int) ([]*Task, error) {
	collection := m.database.Collection("tasks")

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mongoTasks []MongoTask
	if err := cursor.All(ctx, &mongoTasks); err != nil {
		return nil, err
	}

	tasks := make([]*Task, len(mongoTasks))
	for i, mongoTask := range mongoTasks {
		tasks[i] = &Task{
			ID:             mongoTask.ID.Hex(),
			UserID:         mongoTask.UserID,
			Type:           mongoTask.Type,
			Status:         mongoTask.Status,
			Input:          mongoTask.Input,
			Output:         mongoTask.Output,
			ErrorMsg:       mongoTask.ErrorMsg,
			ExternalTaskID: mongoTask.ExternalTaskID,
			CreatedAt:      mongoTask.CreatedAt,
			UpdatedAt:      mongoTask.UpdatedAt,
			CompletedAt:    mongoTask.CompletedAt,
		}
	}

	return tasks, nil
}

func (m *MongoDB) UpdateTask(ctx context.Context, task *Task) error {
	objectID, err := primitive.ObjectIDFromHex(task.ID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"type":             task.Type,
			"status":           task.Status,
			"input":            task.Input,
			"output":           task.Output,
			"error_msg":        task.ErrorMsg,
			"external_task_id": task.ExternalTaskID,
			"updated_at":       time.Now(),
			"completed_at":     task.CompletedAt,
		},
	}

	collection := m.database.Collection("tasks")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (m *MongoDB) DeleteTask(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	collection := m.database.Collection("tasks")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// AI请求记录相关方法
func (m *MongoDB) CreateAIRequest(ctx context.Context, request *AIRequest) error {
	mongoRequest := &MongoAIRequest{
		UserID:    request.UserID,
		TaskID:    request.TaskID,
		Provider:  request.Provider,
		Model:     request.Model,
		Prompt:    request.Prompt,
		Response:  request.Response,
		Tokens:    request.Tokens,
		Cost:      request.Cost,
		Duration:  request.Duration,
		CreatedAt: time.Now(),
	}

	collection := m.database.Collection("ai_requests")
	result, err := collection.InsertOne(ctx, mongoRequest)
	if err != nil {
		return err
	}

	request.ID = result.InsertedID.(primitive.ObjectID).Hex()
	request.CreatedAt = mongoRequest.CreatedAt
	return nil
}

func (m *MongoDB) GetAIRequestByID(ctx context.Context, id string) (*AIRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mongoRequest MongoAIRequest
	collection := m.database.Collection("ai_requests")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mongoRequest)
	if err != nil {
		return nil, err
	}

	return &AIRequest{
		ID:        mongoRequest.ID.Hex(),
		UserID:    mongoRequest.UserID,
		TaskID:    mongoRequest.TaskID,
		Provider:  mongoRequest.Provider,
		Model:     mongoRequest.Model,
		Prompt:    mongoRequest.Prompt,
		Response:  mongoRequest.Response,
		Tokens:    mongoRequest.Tokens,
		Cost:      mongoRequest.Cost,
		Duration:  mongoRequest.Duration,
		CreatedAt: mongoRequest.CreatedAt,
	}, nil
}

func (m *MongoDB) GetAIRequestsByUserID(ctx context.Context, userID string, limit, offset int) ([]*AIRequest, error) {
	collection := m.database.Collection("ai_requests")

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mongoRequests []MongoAIRequest
	if err := cursor.All(ctx, &mongoRequests); err != nil {
		return nil, err
	}

	requests := make([]*AIRequest, len(mongoRequests))
	for i, mongoRequest := range mongoRequests {
		requests[i] = &AIRequest{
			ID:        mongoRequest.ID.Hex(),
			UserID:    mongoRequest.UserID,
			TaskID:    mongoRequest.TaskID,
			Provider:  mongoRequest.Provider,
			Model:     mongoRequest.Model,
			Prompt:    mongoRequest.Prompt,
			Response:  mongoRequest.Response,
			Tokens:    mongoRequest.Tokens,
			Cost:      mongoRequest.Cost,
			Duration:  mongoRequest.Duration,
			CreatedAt: mongoRequest.CreatedAt,
		}
	}

	return requests, nil
}

func (m *MongoDB) Close() error {
	return m.client.Disconnect(context.Background())
}
