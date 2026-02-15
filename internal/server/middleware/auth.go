package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask/internal/api"
	"fcstask/internal/db/repo"
	"fcstask/internal/server/handler"
)

func authError(ctx echo.Context, message string) error {
	return ctx.JSON(http.StatusUnauthorized, api.Error{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{Code: "unauthorized", Message: message},
	})
}

func Auth(userRepo repo.UserRepositoryInterface, sessionRepo repo.SessionRepositoryInterface, protectedPaths []string) echo.MiddlewareFunc {
	protected := make(map[string]bool, len(protectedPaths))
	for _, p := range protectedPaths {
		protected[p] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if !protected[ctx.Path()] {
				return next(ctx)
			}

			authHeader := ctx.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return authError(ctx, "Missing or invalid Authorization header")
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			sessionID, err := uuid.Parse(tokenStr)
			if err != nil {
				return authError(ctx, "Invalid session token")
			}

			session, err := sessionRepo.GetByID(ctx.Request().Context(), sessionID)
			if err != nil {
				return authError(ctx, "Session not found")
			}

			user, err := userRepo.GetByID(ctx.Request().Context(), session.UserID)
			if err != nil {
				return authError(ctx, "User not found")
			}

			_ = sessionRepo.TouchAccessedAt(ctx.Request().Context(), session.ID)

			ctx.Set(handler.UserContextKey, user)
			ctx.Set(handler.SessionContextKey, session)
			return next(ctx)
		}
	}
}
