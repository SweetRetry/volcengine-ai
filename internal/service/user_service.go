package service

import (
	"context"
	"fmt"

	"volcengine-go-server/internal/models"
	"volcengine-go-server/internal/repository"
)

type UserService struct {
	db repository.Database
}

func NewUserService(db repository.Database) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(ctx context.Context, email, name string) (*models.User, error) {
	// 检查用户是否已存在
	existingUser, err := s.db.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("用户已存在: %s", email)
	}

	user := &models.User{
		Email: email,
		Name:  name,
	}

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.db.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	if err := s.db.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if err := s.db.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	return nil
}
