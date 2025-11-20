package service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/service"
)

type mockUserRepo struct {
	UpdateFunc func(ctx context.Context, userID string, isActive bool) (*api.User, error)
}

func (m *mockUserRepo) UpdateActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	return m.UpdateFunc(ctx, userID, isActive)
}

func TestUserService_SetActive(t *testing.T) {
	mockRepo := &mockUserRepo{
		UpdateFunc: func(ctx context.Context, userID string, isActive bool) (*api.User, error) {
			if userID == "notfound" {
				return nil, sql.ErrNoRows
			}
			return &api.User{UserId: userID, Username: "test", IsActive: isActive}, nil
		},
	}
	log := logger.NewLogger("app", logger.LevelInfo)
	svc := service.NewUserService(mockRepo, log)

	user, err := svc.SetIsActive(context.Background(), "123", true)
	if err != nil {
		t.Fatal(err)
	}
	if !user.IsActive {
		t.Errorf("expected active true, got false")
	}

	_, err = svc.SetIsActive(context.Background(), "notfound", true)
	if err == nil {
		t.Errorf("expected error for notfound")
	}
}
