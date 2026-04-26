package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/service"
)

const (
	errorCodeBadRequest   = "bad_request"
	errorCodeUnauthorized = "unauthorized"
	errorCodeNotFound     = "not_found"
	errorCodeConflict     = "conflict"
	errorCodeInternal     = "internal_error"
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
	return apiError(ctx, http.StatusBadRequest, errorCodeBadRequest, message)
}

func unauthorized(ctx echo.Context, message string) error {
	return apiError(ctx, http.StatusUnauthorized, errorCodeUnauthorized, message)
}

func conflict(ctx echo.Context, message string) error {
	return apiError(ctx, http.StatusConflict, errorCodeConflict, message)
}

func internalError(ctx echo.Context, message string) error {
	return apiError(ctx, http.StatusInternalServerError, errorCodeInternal, message)
}

func serviceError(ctx echo.Context, err error) error {
	if err == nil {
		return nil
	}

	if serviceErr, ok := err.(*service.Error); ok {
		switch serviceErr.Code {
		case errorCodeBadRequest:
			return badRequest(ctx, serviceErr.Message)
		case errorCodeUnauthorized:
			return unauthorized(ctx, serviceErr.Message)
		case errorCodeNotFound:
			return apiError(ctx, http.StatusNotFound, errorCodeNotFound, serviceErr.Message)
		case errorCodeConflict:
			return conflict(ctx, serviceErr.Message)
		default:
			return internalError(ctx, serviceErr.Message)
		}
	}

	return internalError(ctx, "Internal server error")
}
