package handlers_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/handlers"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/repository"
	"github.com/labstack/echo/v4"
)

type mockUserService struct{}

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
		Members:  []api.TeamMember{},
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
