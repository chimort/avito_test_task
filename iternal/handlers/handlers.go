package handlers

import (
	"errors"
	"net/http"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/pkg/logger"
	"github.com/chimort/avito_test_task/iternal/repository"
	"github.com/chimort/avito_test_task/iternal/service"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	userService service.UserServiceInterface
	log         *logger.Logger
}

func NewHandlers(us service.UserServiceInterface, log *logger.Logger) *Handlers {
	return &Handlers{
		userService: us,
		log:         log,
	}
}

func (h *Handlers) PostTeamAdd(ctx echo.Context) error {
	var body api.PostTeamAddJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		h.log.Error("failed to bind request body", "error", err)
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.NOTFOUND,
				Message: "invalid body",
			},
		})
	}

	if body.TeamName == "" || len(body.Members) == 0 {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.TEAMEXISTS,
				Message: "team_name and members are required",
			},
		})
	}

	for _, m := range body.Members {
		if m.UserId == "" || m.Username == "" {
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.NOTASSIGNED,
					Message: "user_id and username are required",
				},
			})
		}
	}

	team, err := h.userService.TeamAdd(ctx.Request().Context(), body.TeamName, body.Members)
	if err != nil {
		if errors.Is(err, repository.ErrTeamExists) {
			h.log.Warn("team already exists", "team_name", body.TeamName)
			return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.TEAMEXISTS,
					Message: "team_name already exists",
				},
			})
		}
		h.log.Error("failed to create team", "error", err, "team_name", body.TeamName)
		return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.NOTFOUND,
				Message: "failed to create team",
			},
		})
	}

	h.log.Info("team created", "team", team)
	return ctx.JSON(http.StatusCreated, map[string]interface{}{"team": team})
}

func (h *Handlers) GetTeamGet(ctx echo.Context, params api.GetTeamGetParams) error {
	teamName := params.TeamName
	h.log.Info("getting team", "team_name", teamName)
	team, err := h.userService.GetTeam(ctx.Request().Context(), teamName)

	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			h.log.Warn("team not found", "team_name", teamName)
			return ctx.JSON(http.StatusNotFound, api.ErrorResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.NOTFOUND,
					Message: "team not found",
				},
			})
		}
		h.log.Error("failed to get team", "error", err, "team_name", teamName)
		return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.NOTFOUND,
				Message: "failed to get team",
			},
		})
	}

	h.log.Info("team retrieved", "team", team)
	return ctx.JSON(http.StatusOK, map[string]interface{}{"team": team})
}

func (h *Handlers) PostUsersSetIsActive(ctx echo.Context) error {
	var body api.PostUsersSetIsActiveJSONBody
	if err := ctx.Bind(&body); err != nil {
		h.log.Error("failed to bind request body", "error", err)
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.NOTFOUND,
				Message: "invalid body",
			},
		})
	}

	if body.UserId == "" {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.NOTFOUND,
				Message: "user_id is required",
			},
		})
	}

	user, err := h.userService.SetIsActive(ctx.Request().Context(), body.UserId, body.IsActive)
	if err != nil {
		h.log.Error("failed to set user active status", "error", err)
		return ctx.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.NOTFOUND,
				Message: "user not found",
			},
		})
	}

	h.log.Info("user updated", "user", user)
	return ctx.JSON(http.StatusOK, map[string]interface{}{"user": user})
}

func (h *Handlers) PostPullRequestCreate(ctx echo.Context) error {
	var body api.PostPullRequestCreateJSONBody
	if err := ctx.Bind(&body); err != nil {
		h.log.Error("failed to bind request body", "error", err)
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.NOTFOUND,
				Message: "invalid body",
			},
		})
	}

	if body.PullRequestId == "" || body.PullRequestName == "" || body.AuthorId == "" {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.PREXISTS,
				Message: "pull_request_id, pull_request_name and author_id are required",
			},
		})
	}

	pr, err := h.userService.PullRequestCreate(ctx.Request().Context(), body.PullRequestId, body.PullRequestName, body.AuthorId)
	if err != nil {
		if errors.Is(err, repository.ErrPRExists) {
			return ctx.JSON(http.StatusConflict, api.ErrorResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.PREXISTS,
					Message: "PR id already exists",
				},
			})
		} else if errors.Is(err, repository.ErrTeamNotFound) || errors.Is(err, repository.ErrUserNotFound) {
			return ctx.JSON(http.StatusNotFound, api.ErrorResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.NOTFOUND,
					Message: "author or team not found",
				},
			})
		} else {
			h.log.Error("failed to create PR", "error", err)
			return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.NOTFOUND,
					Message: "failed to create PR",
				},
			})
		}
	}

	return ctx.JSON(http.StatusCreated, map[string]interface{}{"pr": pr})
}

func (h *Handlers) PostPullRequestMerge(ctx echo.Context) error    { 
	var body api.PostPullRequestMergeJSONBody
	if err := ctx.Bind(&body); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.NOTFOUND,
				Message: "invalid body",
			},
		})
	}

	pr, err := h.userService.PullRequestMerge(ctx.Request().Context(), body.PullRequestId)
	if err != nil {
		if errors.Is(err, repository.ErrPRNotFound) {
			return ctx.JSON(http.StatusNotFound, api.ErrorResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.NOTFOUND,
					Message: "PR not found",
				},
			})
		}
		return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{
				Code:    api.NOTFOUND,
				Message: "failed to merge PR",
			},
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{"pr": pr})
}


func (h *Handlers) PostPullRequestReassign(ctx echo.Context) error { return nil }
func (h *Handlers) GetUsersGetReview(ctx echo.Context, params api.GetUsersGetReviewParams) error {
	return nil
}
