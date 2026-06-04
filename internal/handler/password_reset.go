package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/config"
	"fcstask-backend/internal/service"
)

type PasswordResetHandler struct {
	passwordResetService *service.PasswordResetService
	mailerCfg            config.MailerConfig
}

func NewPasswordResetHandler(passwordResetService *service.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{passwordResetService: passwordResetService}
}

// WithMailerConfig supplies the reset-code resend cooldown and max attempts.
func (h *PasswordResetHandler) WithMailerConfig(cfg config.MailerConfig) *PasswordResetHandler {
	h.mailerCfg = cfg
	return h
}

// PasswordResetRequest starts a password reset and emails a code (202 Accepted).
func (h *PasswordResetHandler) PasswordResetRequest(ctx echo.Context) error {
	var req api.PasswordResetRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	pr, err := h.passwordResetService.PasswordResetRequest(ctx.Request().Context(), string(req.Email))
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusAccepted, passwordResetPendingToAPI(pr, h.mailerCfg.ResendCooldown))
}

// PasswordResetResend re-issues the password reset code (200 OK).
func (h *PasswordResetHandler) PasswordResetResend(ctx echo.Context) error {
	var req api.PasswordResetResendRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	pr, err := h.passwordResetService.PasswordResetResend(ctx.Request().Context(), string(req.Email))
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, passwordResetPendingToAPI(pr, h.mailerCfg.ResendCooldown))
}

// PasswordResetConfirm validates the code and sets the new password (204 No Content).
func (h *PasswordResetHandler) PasswordResetConfirm(ctx echo.Context) error {
	var req api.PasswordResetConfirmRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}

	err := h.passwordResetService.PasswordResetConfirm(ctx.Request().Context(), service.PasswordResetConfirmInput{
		Email:       string(req.Email),
		Code:        req.Code,
		Password:    req.NewPassword,
		MaxAttempts: h.mailerCfg.MaxAttempts,
	})
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}
