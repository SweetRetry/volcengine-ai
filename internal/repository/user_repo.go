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

// 用户repository的具体实现在mongodb.go中
// 接口定义已移动到internal/models/interfaces.go

// UserRepositoryImpl MongoDB用户repository实现
type UserRepositoryImpl struct {
	database *mongo.Database
}

// NewUserRepository 创建用户repository实例
func NewUserRepository(database *mongo.Database) UserRepository {
	return &UserRepositoryImpl{database: database}
}

// CreateUser 创建用户
func (r *UserRepositoryImpl) CreateUser(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	collection := r.database.Collection("users")
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

// GetUserByID 根据ID获取用户
func (r *UserRepositoryImpl) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user models.User
	collection := r.database.Collection("users")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	user.ID = objectID.Hex()
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (r *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	collection := r.database.Collection("users")
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser 更新用户
func (r *UserRepositoryImpl) UpdateUser(ctx context.Context, user *models.User) error {
	objectID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return err
	}

	user.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"name":       user.Name,
			"updated_at": user.UpdatedAt,
		},
	}

	collection := r.database.Collection("users")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// DeleteUser 删除用户
func (r *UserRepositoryImpl) DeleteUser(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	collection := r.database.Collection("users")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// CreateUserIndexes 创建用户相关的索引
func (r *UserRepositoryImpl) CreateUserIndexes(ctx context.Context) error {
	userCollection := r.database.Collection("users")
	_, err := userCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}
