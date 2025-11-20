package handlers

import (
	"context"
	"net/http"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/labstack/echo/v4"
)

type UserServiceInterface interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*api.User, error)
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

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

func (h *Handlers) PostPullRequestCreate(ctx echo.Context) error                   { return nil }
func (h *Handlers) PostPullRequestMerge(ctx echo.Context) error                    { return nil }
func (h *Handlers) PostPullRequestReassign(ctx echo.Context) error                 { return nil }
func (h *Handlers) PostTeamAdd(ctx echo.Context) error                             { return nil }
func (h *Handlers) GetTeamGet(ctx echo.Context, params api.GetTeamGetParams) error { return nil }
func (h *Handlers) GetUsersGetReview(ctx echo.Context, params api.GetUsersGetReviewParams) error {
	return nil
}
