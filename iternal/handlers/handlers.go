package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/repository"
	"github.com/labstack/echo/v4"
)

type UserServiceInterface interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error)
	TeamAdd(ctx context.Context, teamName string, teamMembers []api.TeamMember) (*api.Team, error)
}

type Handlers struct {
	userService UserServiceInterface
	log         *logger.Logger
}

func NewHandlers(us UserServiceInterface, log *logger.Logger) *Handlers {
	return &Handlers{
		userService: us,
		log:         log,
	}
}

func (h *Handlers) PostTeamAdd(ctx echo.Context) error {
	var body api.PostTeamAddJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		h.log.Error("failed to bind request body", "error", err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}

	team, err := h.userService.TeamAdd(ctx.Request().Context(), body.TeamName, body.Members)
	if err != nil {
		if errors.Is(err, repository.ErrTeamExists) {
			h.log.Warn("team already exists", "team_name", body.TeamName)
			return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": map[string]string{
					"code":    "TEAM_EXISTS",
					"message": "team_name already exists",
				},
			})
		}
		h.log.Error("failed to create team", "error", err, "team_name", body.TeamName)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create team",
			},
		})
	}

	h.log.Info("team created", "team", team)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"team": team,
	})
}

func (h *Handlers) PostUsersSetIsActive(ctx echo.Context) error {
	var body api.PostUsersSetIsActiveJSONBody
	if err := ctx.Bind(&body); err != nil {
		h.log.Error("failed to bind request body", "error", err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}

	user, err := h.userService.SetIsActive(ctx.Request().Context(), body.UserId, body.IsActive)
	if err != nil {
		h.log.Error("failed to set user active status", "error", err)
		return ctx.JSON(http.StatusNotFound, map[string]interface{}{
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "user not found",
			},
		})
	}
	h.log.Info("user updated", "user", user)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

func (h *Handlers) PostPullRequestCreate(ctx echo.Context) error                   { return nil }
func (h *Handlers) PostPullRequestMerge(ctx echo.Context) error                    { return nil }
func (h *Handlers) PostPullRequestReassign(ctx echo.Context) error                 { return nil }
func (h *Handlers) GetTeamGet(ctx echo.Context, params api.GetTeamGetParams) error { return nil }
func (h *Handlers) GetUsersGetReview(ctx echo.Context, params api.GetUsersGetReviewParams) error {
	return nil
}
