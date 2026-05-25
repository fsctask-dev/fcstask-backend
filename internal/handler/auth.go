package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/service"
)

type AuthHandler struct {
	authService IAuthService
}

func NewAuthHandler(authService IAuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) SignUp(ctx echo.Context) error {
	var req api.SignUpRequest
	if !bindRequest(ctx, &req, "Invalid request body") {
		return nil
	}

	result, err := h.authService.SignUp(ctx.Request().Context(), service.SignUpInput{
		Email:     string(req.Email),
		Username:  req.Username,
		Password:  req.Password,
		TgUID:     req.TgUid,
		IP:        ctx.RealIP(),
		UserAgent: ctx.Request().UserAgent(),
	})
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, authResultToAPI(result))
}

func (h *AuthHandler) SignIn(ctx echo.Context) error {
	var req api.SignInRequest
	if !bindRequest(ctx, &req, "Invalid request body") {
		return nil
	}

	var email *string
	if req.Email != nil {
		value := string(*req.Email)
		email = &value
	}

	result, err := h.authService.SignIn(ctx.Request().Context(), service.SignInInput{
		Email:     email,
		Username:  req.Username,
		Password:  req.Password,
		IP:        ctx.RealIP(),
		UserAgent: ctx.Request().UserAgent(),
	})
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, authResultToAPI(result))
}

func (h *AuthHandler) GetMe(ctx echo.Context) error {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return unauthorized(ctx, "Not authenticated")
	}

	initials, role, err := h.authService.GetMe(ctx.Request().Context(), user)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, api.MeResponse{
		Username: user.Username,
		Initials: initials,
		Role:     role,
	})
}

func (h *AuthHandler) SignOut(ctx echo.Context) error {
	session, ok := authenticatedSession(ctx)
	if !ok {
		return unauthorized(ctx, "Not authenticated")
	}

	if err := h.authService.SignOut(ctx.Request().Context(), session); err != nil {
		return serviceError(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}
