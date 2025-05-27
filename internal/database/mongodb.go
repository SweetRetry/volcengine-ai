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

type MongoImageTask struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   string             `bson:"user_id"`
	Prompt   string             `bson:"prompt"`
	Model    string             `bson:"model"`
	Size     string             `bson:"size"`
	N        int                `bson:"n"`
	Status   string             `bson:"status"`
	ImageURL string             `bson:"image_url"`
	Error    string             `bson:"error"`
	Created  time.Time          `bson:"created"`
	Updated  time.Time          `bson:"updated"`
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

	database := client.Database("volcengine_db")

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

	// 图像任务用户ID索引
	imageTaskCollection := m.database.Collection("image_tasks")
	_, err = imageTaskCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
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

// 图像任务相关方法
func (m *MongoDB) CreateImageTask(ctx context.Context, task *ImageTask) error {
	mongoTask := &MongoImageTask{
		UserID:  task.UserID,
		Prompt:  task.Prompt,
		Model:   task.Model,
		Size:    task.Size,
		N:       task.N,
		Status:  task.Status,
		Created: time.Now(),
		Updated: time.Now(),
	}

	collection := m.database.Collection("image_tasks")
	result, err := collection.InsertOne(ctx, mongoTask)
	if err != nil {
		return err
	}

	task.ID = result.InsertedID.(primitive.ObjectID).Hex()
	task.Created = mongoTask.Created
	task.Updated = mongoTask.Updated
	return nil
}

func (m *MongoDB) GetImageTaskByID(ctx context.Context, id string) (*ImageTask, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mongoTask MongoImageTask
	collection := m.database.Collection("image_tasks")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mongoTask)
	if err != nil {
		return nil, err
	}

	return &ImageTask{
		ID:       mongoTask.ID.Hex(),
		UserID:   mongoTask.UserID,
		Prompt:   mongoTask.Prompt,
		Model:    mongoTask.Model,
		Size:     mongoTask.Size,
		N:        mongoTask.N,
		Status:   mongoTask.Status,
		ImageURL: mongoTask.ImageURL,
		Error:    mongoTask.Error,
		Created:  mongoTask.Created,
		Updated:  mongoTask.Updated,
	}, nil
}

func (m *MongoDB) GetImageTasksByUserID(ctx context.Context, userID string, limit, offset int) ([]*ImageTask, error) {
	collection := m.database.Collection("image_tasks")

	opts := options.Find().
		SetSort(bson.D{{Key: "created", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mongoTasks []MongoImageTask
	if err := cursor.All(ctx, &mongoTasks); err != nil {
		return nil, err
	}

	tasks := make([]*ImageTask, len(mongoTasks))
	for i, mongoTask := range mongoTasks {
		tasks[i] = &ImageTask{
			ID:       mongoTask.ID.Hex(),
			UserID:   mongoTask.UserID,
			Prompt:   mongoTask.Prompt,
			Model:    mongoTask.Model,
			Size:     mongoTask.Size,
			N:        mongoTask.N,
			Status:   mongoTask.Status,
			ImageURL: mongoTask.ImageURL,
			Error:    mongoTask.Error,
			Created:  mongoTask.Created,
			Updated:  mongoTask.Updated,
		}
	}

	return tasks, nil
}

func (m *MongoDB) UpdateImageTaskStatus(ctx context.Context, id, status, imageURL, errorMsg string) error {
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

	collection := m.database.Collection("image_tasks")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (m *MongoDB) DeleteImageTask(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	collection := m.database.Collection("image_tasks")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (m *MongoDB) Close() error {
	return m.client.Disconnect(context.Background())
}
