package service

import (
	"context"

	"github.com/chimort/avito_test_task/iternal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	return s.repo.UpdateActive(ctx, userID, isActive)
}
