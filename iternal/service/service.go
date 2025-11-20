package service

import (
	"context"
	"database/sql"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/repository"
)

type UserService struct {
	repo repository.UserRepo
	log  *logger.Logger
}

func NewUserService(repo repository.UserRepo, log *logger.Logger) *UserService {
	return &UserService{
		repo: repo,
		log:  log,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	s.log.Info("updating user active status", "user_id", userID, "active", isActive)
	user, err := s.repo.UpdateActive(ctx, userID, isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Error("no rows in sql", "error", err)
			return nil, err
		}
		s.log.Error("failed to update user active status", "error", err)
		return nil, err
	}
	s.log.Info("user updated", "user", user)
	return user, nil
}
