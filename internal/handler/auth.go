package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/api"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

const UserContextKey = "user"
const SessionContextKey = "session"

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) SignUp(ctx echo.Context) error {
	var req api.SignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
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
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
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
	user, ok := ctx.Get(UserContextKey).(*models.User)
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
	session, ok := ctx.Get(SessionContextKey).(*models.Session)
	if !ok {
		return unauthorized(ctx, "Not authenticated")
	}

	if err := h.authService.SignOut(ctx.Request().Context(), session); err != nil {
		return serviceError(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}
