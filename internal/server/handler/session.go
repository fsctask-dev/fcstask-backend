package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"fcstask/internal/api"
	"fcstask/internal/db/repo"
)

func parsePagination(limit, offset *int) (int, int, error) {
	l := 20
	o := 0

	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}

	if l < 1 || l > 100 {
		return 0, 0, errPagination("Limit must be between 1 and 100")
	}

	if o < 0 {
		return 0, 0, errPagination("Offset must be non-negative")
	}

	return l, o, nil
}

type paginationError string

func errPagination(msg string) paginationError {
	return paginationError(msg)
}

func (e paginationError) Error() string {
	return string(e)
}

func GetSessionsHandler(sessionRepo repo.SessionRepositoryInterface, ctx echo.Context, params api.GetSessionsParams) error {
	limit, offset, err := parsePagination(params.Limit, params.Offset)
	if err != nil {
		return badRequest(ctx, err.Error())
	}

	total, err := sessionRepo.CountAll(ctx.Request().Context())
	if err != nil {
		return internalError(ctx, "Failed to count sessions")
	}

	sessions, err := sessionRepo.GetAllWithUser(ctx.Request().Context(), limit, offset)
	if err != nil {
		return internalError(ctx, "Failed to get sessions")
	}

	items := make([]api.SessionWithUser, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, api.SessionWithUser{
			Id:        openapi_types.UUID(s.ID),
			Ip:        s.IP,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
			User:      userToAPI(&s.User),
		})
	}

	return ctx.JSON(http.StatusOK, api.PaginatedSessionsResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func GetUsersWithSessionsHandler(userRepo repo.UserRepositoryInterface, ctx echo.Context, params api.GetUsersWithSessionsParams) error {
	limit, offset, err := parsePagination(params.Limit, params.Offset)
	if err != nil {
		return badRequest(ctx, err.Error())
	}

	total, err := userRepo.CountUsersWithSessions(ctx.Request().Context())
	if err != nil {
		return internalError(ctx, "Failed to count users")
	}

	users, err := userRepo.GetAllWithSessions(ctx.Request().Context(), limit, offset)
	if err != nil {
		return internalError(ctx, "Failed to get users with sessions")
	}

	items := make([]api.UserWithSessions, 0, len(users))
	for _, u := range users {
		sessions := make([]api.Session, 0, len(u.Sessions))
		for _, s := range u.Sessions {
			sessions = append(sessions, api.Session{
				Id:        openapi_types.UUID(s.ID),
				Ip:        s.IP,
				UserAgent: s.UserAgent,
				CreatedAt: s.CreatedAt,
				UpdatedAt: s.UpdatedAt,
			})
		}

		items = append(items, api.UserWithSessions{
			User:     userToAPI(&u),
			Sessions: sessions,
		})
	}

	return ctx.JSON(http.StatusOK, api.PaginatedUsersWithSessionsResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
