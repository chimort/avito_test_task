package service

import (
	"context"
	"errors"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/repository"
)

type UserServiceInterface interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error)
	TeamAdd(ctx context.Context, teamName string, teamMembers []api.TeamMember) (*api.Team, error)
	GetTeam(ctx context.Context, teamName string) (*api.Team, error)
	PullRequestCreate(ctx context.Context, pullRequestId string, pullRequestName string, authorId string) (*api.PullRequest, error)
	PullRequestMerge(ctx context.Context, pullRequestId string) (*api.PullRequest, error)
}

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

func (s *UserService) TeamAdd(ctx context.Context, teamName string, teamMembers []api.TeamMember) (*api.Team, error) {
	s.log.Info("adding team", "team_name", teamName, "members", teamMembers)
	team, err := s.repo.TeamAdd(ctx, teamName, teamMembers)
	if err != nil {
		if errors.Is(err, repository.ErrTeamExists) {
			s.log.Warn("team already exists", "team_name", teamName)
			return nil, repository.ErrTeamExists
		}
		s.log.Error("failed to create team", "error", err, "team_name", teamName)
		return nil, err
	}
	s.log.Info("added team", "team_name", teamName)
	return team, nil
}

func (s *UserService) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	s.log.Info("getting team", "team_name", teamName)
	team, err := s.repo.GetTeam(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			s.log.Warn("team not found", "team_name", teamName)
			return nil, repository.ErrTeamNotFound
		}
		s.log.Error("failed to get team", "error", err, "team_name", teamName)
		return nil, err
	}
	s.log.Info("got team", "team", team)
	return team, nil
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	s.log.Info("updating user active status", "user_id", userID, "active", isActive)
	user, err := s.repo.UpdateActive(ctx, userID, isActive)
	if err != nil {
		s.log.Error("failed to update user active status", "error", err)
		return nil, err
	}
	s.log.Info("user updated", "user", user)
	return user, nil
}

func (s *UserService) PullRequestCreate(ctx context.Context, pullRequestId string, pullRequestName string, authorId string) (*api.PullRequest, error) {
	s.log.Info("creating pull request", "pr_id", pullRequestId, "pr_name", pullRequestName, "author_id", authorId)
	pr, err := s.repo.PullRequestCreate(ctx, pullRequestId, pullRequestName, authorId)
	if err != nil {
		s.log.Error("failed to create pull request", "error", err)
		return nil, err
	}
	s.log.Info("pull request created", "pr_id", pullRequestId)
	return pr, nil
}

func (s *UserService) PullRequestMerge(ctx context.Context, pullRequestId string) (*api.PullRequest, error) {
	s.log.Info("merging pull request", "pr_id", pullRequestId)
	pr, err := s.repo.PullRequestMerge(ctx, pullRequestId)
	if err != nil {
		s.log.Error("failed to merge pull request", "error", err)
		return nil, err
	}
	s.log.Info("pull request merged", "pr_id", pullRequestId)
	return pr, nil
}