package service

import (
	"context"

	"volcengine-go-server/internal/models"
	"volcengine-go-server/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(db repository.Database) *UserService {
	return &UserService{
		userRepo: db.UserRepository(),
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	return s.userRepo.CreateUser(ctx, user)
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetUserByID(ctx, id)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetUserByEmail(ctx, email)
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	return s.userRepo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepo.DeleteUser(ctx, id)
}
