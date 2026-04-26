package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/service"
)

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

func serviceError(ctx echo.Context, err error) error {
	if err == nil {
		return nil
	}

	if serviceErr, ok := err.(*service.Error); ok {
		switch serviceErr.Code {
		case "bad_request":
			return badRequest(ctx, serviceErr.Message)
		case "unauthorized":
			return unauthorized(ctx, serviceErr.Message)
		case "not_found":
			return apiError(ctx, http.StatusNotFound, "not_found", serviceErr.Message)
		case "conflict":
			return conflict(ctx, serviceErr.Message)
		default:
			return internalError(ctx, serviceErr.Message)
		}
	}

	return internalError(ctx, "Internal server error")
}
