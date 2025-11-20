package handlers

import (
	"net/http"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/service"
	"github.com/labstack/echo/v4"
)


type Handlers struct {
	userService *service.UserService
}

func NewHandlers(us *service.UserService) *Handlers {
	return &Handlers{userService: us}
}

func (h *Handlers) PostUsersSetIsActive(ctx echo.Context) error {
	var body api.PostUsersSetIsActiveJSONBody
	if err := ctx.Bind(&body); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	if err := h.userService.SetIsActive(ctx.Request().Context(), body.UserId, body.IsActive); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handlers) PostPullRequestCreate(ctx echo.Context) error   { return nil }
func (h *Handlers) PostPullRequestMerge(ctx echo.Context) error    { return nil }
func (h *Handlers) PostPullRequestReassign(ctx echo.Context) error { return nil }
func (h *Handlers) PostTeamAdd(ctx echo.Context) error             { return nil }
func (h *Handlers) GetTeamGet(ctx echo.Context, params api.GetTeamGetParams) error { return nil }
func (h *Handlers) GetUsersGetReview(ctx echo.Context, params api.GetUsersGetReviewParams) error { return nil }