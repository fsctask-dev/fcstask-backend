package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/service"
)

type SessionHandler struct {
	sessionService *service.SessionService
	userService    *service.UserService
}

func NewSessionHandler(sessionService *service.SessionService, userService *service.UserService) *SessionHandler {
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

	items := make([]sessionWithUserResponse, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, sessionWithUserResponse{
			sessionResponse: sessionResponse{
				Id:        openapi_types.UUID(s.ID),
				Ip:        s.IP,
				UserAgent: s.UserAgent,
				CreatedAt: s.CreatedAt,
				UpdatedAt: s.UpdatedAt,
			},
			User: userToAPI(&s.User),
		})
	}

	return ctx.JSON(http.StatusOK, paginatedSessionsResponse{
		Items:  items,
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

	items := make([]userWithSessionsResponse, 0, len(users))
	for _, u := range users {
		sessions := make([]sessionResponse, 0, len(u.Sessions))
		for _, s := range u.Sessions {
			sessions = append(sessions, sessionResponse{
				Id:        openapi_types.UUID(s.ID),
				Ip:        s.IP,
				UserAgent: s.UserAgent,
				CreatedAt: s.CreatedAt,
				UpdatedAt: s.UpdatedAt,
			})
		}

		items = append(items, userWithSessionsResponse{
			User:     userToAPI(&u),
			Sessions: sessions,
		})
	}

	return ctx.JSON(http.StatusOK, paginatedUsersWithSessionsResponse{
		Items:  items,
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
