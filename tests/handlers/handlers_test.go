package handlers_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/handlers"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/repository"
	"github.com/labstack/echo/v4"
)

type mockUserService struct{}

var t = time.Now()

func (m *mockUserService) TeamAdd(ctx context.Context, teamName string, members []api.TeamMember) (*api.Team, error) {
	if teamName == "existing" {
		return nil, repository.ErrTeamExists
	}
	return &api.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

func (m *mockUserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {
	if userID == "notfound" {
		return nil, sql.ErrNoRows
	}
	return &api.User{
		UserId:   userID,
		Username: "test",
		IsActive: isActive,
	}, nil
}

func (m *mockUserService) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	if teamName == "notfound" {
		return nil, repository.ErrTeamNotFound
	}
	return &api.Team{
		TeamName: teamName,
		Members:  []api.TeamMember{{UserId: "u1", Username: "Alice", IsActive: true}},
	}, nil
}

func (m *mockUserService) PullRequestCreate(ctx context.Context, prID, prName, authorID string) (*api.PullRequest, error) {
	if prID == "pr-existing" {
		return nil, repository.ErrPRExists
	}
	return &api.PullRequest{
		PullRequestId:   prID,
		PullRequestName: prName,
		AuthorId:        authorID,
		Status:          "MERGED",
		AssignedReviewers: []string{"u3", "u4"},
		CreatedAt:       &t,
	}, nil
}

func (m *mockUserService) PullRequestMerge(ctx context.Context, prID string) (*api.PullRequest, error) {
	if prID == "pr-notfound" {
		return nil, repository.ErrPRNotFound
	}
	t := time.Now()
	return &api.PullRequest{
		PullRequestId:   prID,
		PullRequestName: "PR Name",
		AuthorId:        "u1",
		Status:          "MERGED",
		CreatedAt:       &t,
		MergedAt:        &t,
	}, nil
}

func (m *mockUserService) PullRequestReassign(ctx context.Context, prID, oldUserID string) (*api.PullRequest, string, error) {
	if prID == "pr-notfound" {
		return nil, "", repository.ErrPRNotFound
	}
	if oldUserID == "notassigned" {
		return nil, "", repository.ErrReviewerNotAssign
	}
	if prID == "pr-merged" {
		return nil, "", repository.ErrPRMerged
	}
	if prID == "pr-nocandidate" {
		return nil, "", repository.ErrNoCandidates
	}
	return &api.PullRequest{
		PullRequestId:   prID,
		PullRequestName: "PR Name",
		AuthorId:        "u1",
		Status:          "OPEN",
		CreatedAt:       &t,
	}, "u5", nil
}

func (m *mockUserService) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]*api.PullRequestShort, error) {
	if reviewerID == "empty" {
		return []*api.PullRequestShort{}, nil
	}
	return []*api.PullRequestShort{
		{
			PullRequestId:   "pr-1",
			PullRequestName: "Add feature X",
			AuthorId:        "u1",
			Status:          "OPEN",
		},
		{
			PullRequestId:   "pr-2",
			PullRequestName: "Fix bug Y",
			AuthorId:        "u1",
			Status:          "OPEN",
		},
	}, nil
}

func TestPostUsersSetIsActive(t *testing.T) {
	e := echo.New()
	us := &mockUserService{}
	log := logger.NewLogger("app", logger.LevelInfo)
	h := handlers.NewHandlers(us, log)

	e.POST("/users/setIsActive", h.PostUsersSetIsActive)

	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive",
		strings.NewReader(`{"user_id":"123","is_active":true}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/users/setIsActive",
		strings.NewReader(`{"user_id":"notfound","is_active":true}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestPostTeamAdd(t *testing.T) {
	e := echo.New()
	us := &mockUserService{}
	log := logger.NewLogger("app", logger.LevelInfo)
	h := handlers.NewHandlers(us, log)

	e.POST("/team/add", h.PostTeamAdd)

	body := `{"team_name":"payments","members":[{"user_id":"u1","username":"Alice","is_active":true}]}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	body = `{"team_name":"existing","members":[{"user_id":"u1","username":"Alice","is_active":true}]}`
	req = httptest.NewRequest(http.MethodPost, "/team/add", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetTeamGet(t *testing.T) {
	e := echo.New()
	us := &mockUserService{}
	log := logger.NewLogger("app", logger.LevelInfo)
	h := handlers.NewHandlers(us, log)

	e.GET("/team/get", func(c echo.Context) error {
		params := api.GetTeamGetParams{
			TeamName: c.QueryParam("team_name"),
		}
		return h.GetTeamGet(c, params)
	})

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=payments", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/team/get?team_name=notfound", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestPostPullRequestCreate(t *testing.T) {
	e := echo.New()
	us := &mockUserService{}
	log := logger.NewLogger("app", logger.LevelInfo)
	h := handlers.NewHandlers(us, log)

	e.POST("/pullRequest/create", h.PostPullRequestCreate)

	body := `{"pull_request_id":"pr-new","pull_request_name":"Test PR","author_id":"u1"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	body = `{"pull_request_id":"pr-existing","pull_request_name":"Test PR","author_id":"u1"}`
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
}

func TestPostPullRequestMerge(t *testing.T) {
	e := echo.New()
	us := &mockUserService{}
	log := logger.NewLogger("app", logger.LevelInfo)
	h := handlers.NewHandlers(us, log)

	e.POST("/pullRequest/merge", h.PostPullRequestMerge)

	body := `{"pull_request_id":"pr-new"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	body = `{"pull_request_id":"pr-notfound"}`
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/merge", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestPostPullRequestReassign(t *testing.T) {
	e := echo.New()
	us := &mockUserService{}
	log := logger.NewLogger("app", logger.LevelInfo)
	h := handlers.NewHandlers(us, log)

	e.POST("/pullRequest/reassign", h.PostPullRequestReassign)

	body := `{"pull_request_id":"pr-1","old_user_id":"u3"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	body = `{"pull_request_id":"pr-notfound","old_user_id":"u3"}`
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGetUsersGetReview(t *testing.T) {
	e := echo.New()
	us := &mockUserService{}
	log := logger.NewLogger("app", logger.LevelInfo)
	h := handlers.NewHandlers(us, log)

	e.GET("/users/getReview", func(c echo.Context) error {
		params := api.GetUsersGetReviewParams{
			UserId: c.QueryParam("user_id"),
		}
		return h.GetUsersGetReview(c, params)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u3", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=empty", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
