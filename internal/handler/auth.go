package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/config"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

const UserContextKey = "user"
const SessionContextKey = "session"

type AuthHandler struct {
	authService *service.AuthService
	mailerCfg   config.MailerConfig
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// WithMailerConfig supplies the verification-code timing/attempt limits used by
// the email registration flow (code TTL, resend cooldown, max attempts).
func (h *AuthHandler) WithMailerConfig(cfg config.MailerConfig) *AuthHandler {
	h.mailerCfg = cfg
	return h
}

// SignUp begins email/password registration and emails a verification code.
// The client must call SignUpVerify with the returned verification_token to
// finish (202 Accepted).
func (h *AuthHandler) SignUp(ctx echo.Context) error {
	var req api.SignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	reg, err := h.authService.SignUp(ctx.Request().Context(), service.SignUpInput{
		Email:     string(req.Email),
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusAccepted, signUpPendingToAPI(reg, h.mailerCfg.ResendCooldown))
}

// SignUpVerify finishes registration by validating the emailed code, creating
// the user and opening a session (201 Created).
func (h *AuthHandler) SignUpVerify(ctx echo.Context) error {
	var req api.SignUpVerifyRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	result, err := h.authService.SignUpVerify(ctx.Request().Context(), service.SignUpVerifyInput{
		Token:       req.VerificationToken,
		Code:        req.Code,
		MaxAttempts: h.mailerCfg.MaxAttempts,
		IP:          ctx.RealIP(),
		UserAgent:   ctx.Request().UserAgent(),
	})
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, authResultToAPI(result))
}

// SignUpResendCode re-issues the verification code for a pending registration
// (200 OK).
func (h *AuthHandler) SignUpResendCode(ctx echo.Context) error {
	var req api.SignUpResendRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	reg, err := h.authService.SignUpResend(ctx.Request().Context(), req.VerificationToken)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, signUpPendingToAPI(reg, h.mailerCfg.ResendCooldown))
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
