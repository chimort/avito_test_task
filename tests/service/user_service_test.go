package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/repository"
	"github.com/chimort/avito_test_task/iternal/service"
)

type mockRepo struct {
	UpdateFunc  func(ctx context.Context, userID string, isActive bool) (*api.User, error)
	TeamAddFunc func(ctx context.Context, teamName string, members []api.TeamMember) (*api.Team, error)
	GetTeamFunc func(ctx context.Context, teamName string) (*api.Team, error)
}

func (m *mockRepo) UpdateActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	return m.UpdateFunc(ctx, userID, isActive)
}

func (m *mockRepo) TeamAdd(ctx context.Context, teamName string, members []api.TeamMember) (*api.Team, error) {
	return m.TeamAddFunc(ctx, teamName, members)
}

func (m *mockRepo) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	if teamName == "notfound" {
		return nil, repository.ErrTeamNotFound
	}
	return &api.Team{
		TeamName: teamName,
		Members: []api.TeamMember{
			{UserId: "u1", Username: "Alice", IsActive: true},
			{UserId: "u2", Username: "Bob", IsActive: true},
		},
	}, nil
}

func TestUserService_SetActive(t *testing.T) {
	mock := &mockRepo{
		UpdateFunc: func(ctx context.Context, userID string, isActive bool) (*api.User, error) {
			if userID == "notfound" {
				return nil, sql.ErrNoRows
			}
			return &api.User{UserId: userID, Username: "test", IsActive: isActive}, nil
		},
	}
	log := logger.NewLogger("app", logger.LevelInfo)
	svc := service.NewUserService(mock, log)

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

func TestUserService_TeamAdd(t *testing.T) {
	mock := &mockRepo{
		TeamAddFunc: func(ctx context.Context, teamName string, members []api.TeamMember) (*api.Team, error) {
			if teamName == "existing" {
				return nil, repository.ErrTeamExists
			}
			return &api.Team{TeamName: teamName, Members: members}, nil
		},
	}
	log := logger.NewLogger("app", logger.LevelInfo)
	svc := service.NewUserService(mock, log)

	teamMembers := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
	}

	team, err := svc.TeamAdd(context.Background(), "payments", teamMembers)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if team.TeamName != "payments" {
		t.Errorf("expected team_name payments, got %s", team.TeamName)
	}
	if len(team.Members) != 2 {
		t.Errorf("unexpected number of members: %d", len(team.Members))
	}

	_, err = svc.TeamAdd(context.Background(), "existing", teamMembers)
	if !errors.Is(err, repository.ErrTeamExists) {
		t.Fatalf("expected ErrTeamExists, got %v", err)
	}
}


func TestUserService_GetTeam(t *testing.T) {
	mock := &mockRepo{}
	log := logger.NewLogger("app", logger.LevelInfo)
	svc := service.NewUserService(mock, log)

	team, err := svc.GetTeam(context.Background(), "backend")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if team.TeamName != "backend" {
		t.Errorf("expected team_name backend, got %s", team.TeamName)
	}
	if len(team.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(team.Members))
	}

	_, err = svc.GetTeam(context.Background(), "notfound")
	if !errors.Is(err, repository.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}
