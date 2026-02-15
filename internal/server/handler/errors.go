package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	"fcstask/internal/api"
)

// uniqueConstraintColumn extracts the column name from a Postgres unique
// constraint violation (error code 23505). It maps GORM-generated index names
// like "idx_users_email" to "email". Returns "" if the error is not a unique
// violation or the constraint is unrecognised.
func uniqueConstraintColumn(err error) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return ""
	}
	// GORM generates index names as idx_<table>_<column>.
	parts := strings.SplitN(pgErr.ConstraintName, "_", 3)
	if len(parts) == 3 {
		return parts[2]
	}
	return pgErr.ConstraintName
}

func apiError(ctx echo.Context, status int, code, message string) error {
	return ctx.JSON(status, api.Error{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{Code: code, Message: message},
	})
}

func badRequest(ctx echo.Context, message string) error {
	return apiError(ctx, http.StatusBadRequest, "bad_request", message)
}

func unauthorized(ctx echo.Context, message string) error {
	return apiError(ctx, http.StatusUnauthorized, "unauthorized", message)
}

func conflict(ctx echo.Context, message string) error {
	return apiError(ctx, http.StatusConflict, "conflict", message)
}

func internalError(ctx echo.Context, message string) error {
	return apiError(ctx, http.StatusInternalServerError, "internal_error", message)
}
