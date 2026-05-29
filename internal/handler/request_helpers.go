package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func bindRequest(ctx echo.Context, dst interface{}, message string) bool {
	if err := ctx.Bind(dst); err != nil {
		_ = badRequest(ctx, message)
		return false
	}

	return true
}

func parseUUIDParam(ctx echo.Context, name, message string) (uuid.UUID, bool) {
	value, err := uuid.Parse(ctx.Param(name))
	if err != nil {
		_ = badRequest(ctx, message)
		return uuid.Nil, false
	}

	return value, true
}
