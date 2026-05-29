package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/service"
)

type SessionHandler struct {
	sessionService ISessionService
	userService    IUserService
}

func NewSessionHandler(sessionService ISessionService, userService IUserService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService, userService: userService}
}

func (h *SessionHandler) GetSessions(ctx echo.Context, params api.GetSessionsParams) error {
	if err := validatePaginationParams(params.Limit, params.Offset); err != nil {
		return serviceError(ctx, err)
	}

	limit, offset := optionalInt(params.Limit), optionalInt(params.Offset)
	sessions, total, err := h.sessionService.GetSessions(ctx.Request().Context(), limit, offset)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, paginatedSessionsResponse{
		Items:  sessionResultsToAPI(sessions),
		Total:  total,
		Limit:  limitOrDefault(params.Limit),
		Offset: offset,
	})
}

func (h *SessionHandler) GetUsersWithSessions(ctx echo.Context, params api.GetUsersWithSessionsParams) error {
	if err := validatePaginationParams(params.Limit, params.Offset); err != nil {
		return serviceError(ctx, err)
	}

	limit, offset := optionalInt(params.Limit), optionalInt(params.Offset)
	users, total, err := h.userService.GetUsersWithSessions(ctx.Request().Context(), limit, offset)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, paginatedUsersWithSessionsResponse{
		Items:  userSessionsResultsToAPI(users),
		Total:  total,
		Limit:  limitOrDefault(params.Limit),
		Offset: offset,
	})
}

func validatePaginationParams(limit, offset *int) error {
	if limit != nil && (*limit < 1 || *limit > 100) {
		return service.BadRequest("Limit must be between 1 and 100")
	}
	if offset != nil && *offset < 0 {
		return service.BadRequest("Offset must be non-negative")
	}
	return nil
}

func optionalInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func limitOrDefault(value *int) int {
	if value == nil {
		return 20
	}
	return *value
}
