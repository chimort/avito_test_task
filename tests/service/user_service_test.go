package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/repository"
	"github.com/chimort/avito_test_task/iternal/service"
)

var now = time.Now()

type mockRepo struct{}

func (m *mockRepo) UpdateActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	if userID == "notfound" {
		return nil, sql.ErrNoRows
	}
	return &api.User{UserId: userID, Username: "test", IsActive: isActive}, nil
}

func (m *mockRepo) TeamAdd(ctx context.Context, teamName string, members []api.TeamMember) (*api.Team, error) {
	if teamName == "existing" {
		return nil, repository.ErrTeamExists
	}
	return &api.Team{TeamName: teamName, Members: members}, nil
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

func (m *mockRepo) PullRequestCreate(ctx context.Context, prID, prName, authorID string) (*api.PullRequest, error) {
	if prID == "pr-existing" {
		return nil, repository.ErrPRExists
	}
	return &api.PullRequest{
		PullRequestId:   prID,
		PullRequestName: prName,
		AuthorId:        authorID,
		Status:          "open",
		CreatedAt:       &now,
	}, nil
}

func (m *mockRepo) PullRequestMerge(ctx context.Context, prID string) (*api.PullRequest, error) {
	if prID == "pr-notfound" {
		return nil, repository.ErrPRNotFound
	}
	t := now
	return &api.PullRequest{
		PullRequestId: prID,
		Status:        "merged",
		CreatedAt:     &t,
		MergedAt:      &t,
	}, nil
}

func (m *mockRepo) PullRequestReassign(ctx context.Context, prID, oldUserID string) (*api.PullRequest, string, error) {
	switch prID {
	case "pr-notfound":
		return nil, "", repository.ErrPRNotFound
	case "pr-merged":
		return nil, "", repository.ErrPRMerged
	case "pr-nocandidate":
		return nil, "", repository.ErrNoCandidates
	}
	if oldUserID == "notassigned" {
		return nil, "", repository.ErrReviewerNotAssign
	}
	return &api.PullRequest{
		PullRequestId: prID,
		Status:        "open",
		CreatedAt:     &now,
	}, "u5", nil
}

func (m *mockRepo) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]*api.PullRequestShort, error) {
	if reviewerID == "empty" {
		return []*api.PullRequestShort{}, nil
	}
	return []*api.PullRequestShort{
		{PullRequestId: "pr-1", PullRequestName: "Fix bug", AuthorId: "u1", Status: "open"},
		{PullRequestId: "pr-2", PullRequestName: "Add feature", AuthorId: "u2", Status: "open"},
	}, nil
}

func TestUserService_SetIsActive(t *testing.T) {
	svc := service.NewUserService(&mockRepo{}, logger.NewLogger("app", logger.LevelInfo))
	user, err := svc.SetIsActive(context.Background(), "123", true)
	if err != nil {
		t.Fatal(err)
	}
	if !user.IsActive {
		t.Errorf("expected active true")
	}
	_, err = svc.SetIsActive(context.Background(), "notfound", true)
	if err == nil {
		t.Errorf("expected error for notfound")
	}
}

func TestUserService_TeamAdd(t *testing.T) {
	svc := service.NewUserService(&mockRepo{}, logger.NewLogger("app", logger.LevelInfo))
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
	}
	team, err := svc.TeamAdd(context.Background(), "payments", members)
	if err != nil {
		t.Fatal(err)
	}
	if team.TeamName != "payments" {
		t.Errorf("expected payments")
	}
	_, err = svc.TeamAdd(context.Background(), "existing", members)
	if !errors.Is(err, repository.ErrTeamExists) {
		t.Errorf("expected ErrTeamExists")
	}
}

func TestUserService_GetTeam(t *testing.T) {
	svc := service.NewUserService(&mockRepo{}, logger.NewLogger("app", logger.LevelInfo))
	team, err := svc.GetTeam(context.Background(), "backend")
	if err != nil {
		t.Fatal(err)
	}
	if team.TeamName != "backend" {
		t.Errorf("expected backend")
	}
	_, err = svc.GetTeam(context.Background(), "notfound")
	if !errors.Is(err, repository.ErrTeamNotFound) {
		t.Errorf("expected ErrTeamNotFound")
	}
}

func TestUserService_PullRequestCreate(t *testing.T) {
	svc := service.NewUserService(&mockRepo{}, logger.NewLogger("app", logger.LevelInfo))
	pr, err := svc.PullRequestCreate(context.Background(), "pr-new", "Test PR", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if pr.PullRequestId != "pr-new" {
		t.Errorf("expected pr-new")
	}
	_, err = svc.PullRequestCreate(context.Background(), "pr-existing", "Test PR", "u1")
	if !errors.Is(err, repository.ErrPRExists) {
		t.Errorf("expected ErrPRExists")
	}
}

func TestUserService_PullRequestMerge(t *testing.T) {
	svc := service.NewUserService(&mockRepo{}, logger.NewLogger("app", logger.LevelInfo))
	pr, err := svc.PullRequestMerge(context.Background(), "pr-1")
	if err != nil {
		t.Fatal(err)
	}
	if pr.Status != "merged" {
		t.Errorf("expected merged")
	}
	_, err = svc.PullRequestMerge(context.Background(), "pr-notfound")
	if !errors.Is(err, repository.ErrPRNotFound) {
		t.Errorf("expected ErrPRNotFound")
	}
}

func TestUserService_PullRequestReassign(t *testing.T) {
	svc := service.NewUserService(&mockRepo{}, logger.NewLogger("app", logger.LevelInfo))
	_, newUser, err := svc.PullRequestReassign(context.Background(), "pr-1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if newUser != "u5" {
		t.Errorf("expected new user u5")
	}
	_, _, err = svc.PullRequestReassign(context.Background(), "pr-notfound", "u1")
	if !errors.Is(err, repository.ErrPRNotFound) {
		t.Errorf("expected ErrPRNotFound")
	}
}

func TestUserService_GetPRsByReviewer(t *testing.T) {
	svc := service.NewUserService(&mockRepo{}, logger.NewLogger("app", logger.LevelInfo))
	prs, err := svc.GetPRsByReviewer(context.Background(), "u3")
	if err != nil {
		t.Fatal(err)
	}
	if len(prs) != 2 {
		t.Errorf("expected 2 PRs")
	}
	prs, err = svc.GetPRsByReviewer(context.Background(), "empty")
	if err != nil {
		t.Fatal(err)
	}
	if len(prs) != 0 {
		t.Errorf("expected 0 PRs")
	}
}
