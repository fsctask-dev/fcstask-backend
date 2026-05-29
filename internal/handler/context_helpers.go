package handler

import (
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
)

const (
	UserContextKey    = "user"
	SessionContextKey = "session"
)

func authenticatedUser(ctx echo.Context) (*model.User, bool) {
	user, ok := ctx.Get(UserContextKey).(*model.User)
	return user, ok && user != nil
}

func authenticatedSession(ctx echo.Context) (*model.Session, bool) {
	session, ok := ctx.Get(SessionContextKey).(*model.Session)
	return session, ok && session != nil
}

func mustAuthenticatedUser(ctx echo.Context) *model.User {
	return ctx.Get(UserContextKey).(*model.User)
}
