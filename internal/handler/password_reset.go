package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask-backend/internal/config"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/mailer"
)

type PasswordResetHandler struct {
	users    repo.IUserRepo
	sessions repo.SessionRepositoryInterface
	resets   repo.IPasswordResetRepo
	mailer   mailer.Mailer
	cfg      config.MailerConfig
	now      func() time.Time
}

func NewPasswordResetHandler(
	users repo.IUserRepo,
	sessions repo.SessionRepositoryInterface,
	resets repo.IPasswordResetRepo,
	m mailer.Mailer,
	cfg config.MailerConfig,
) *PasswordResetHandler {
	return &PasswordResetHandler{
		users:    users,
		sessions: sessions,
		resets:   resets,
		mailer:   m,
		cfg:      cfg,
		now:      time.Now,
	}
}

type passwordResetRequestBody struct {
	Email string `json:"email"`
}

type passwordResetPendingResponse struct {
	ResetToken  uuid.UUID `json:"reset_token"`
	ResendAfter time.Time `json:"resend_after"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type passwordResetResendBody struct {
	ResetToken uuid.UUID `json:"reset_token"`
}

type passwordResetConfirmBody struct {
	ResetToken  uuid.UUID `json:"reset_token"`
	Code        string    `json:"code"`
	NewPassword string    `json:"new_password"`
}

// Request handles POST /api/password-reset/request.
func (h *PasswordResetHandler) Request(ctx echo.Context) error {
	var req passwordResetRequestBody
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}
	if req.Email == "" {
		return badRequest(ctx, "email is required")
	}

	user, err := h.users.GetUserByEmail(ctx.Request().Context(), req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apiError(ctx, http.StatusNotFound, "not_found", "User with this email not found")
		}
		return internalError(ctx, "Failed to look up user")
	}

	code, err := generateCode()
	if err != nil {
		return internalError(ctx, "Failed to generate reset code")
	}
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return internalError(ctx, "Failed to hash reset code")
	}

	now := h.now().UTC()
	pr := &models.PasswordReset{
		UserID:     user.ID,
		CodeHash:   string(codeHash),
		LastSentAt: now,
		ExpiresAt:  now.Add(h.codeTTL()),
	}

	// At most one outstanding reset per user; replace any prior one.
	_ = h.resets.DeleteByUserID(ctx.Request().Context(), user.ID)

	if err := h.resets.Create(ctx.Request().Context(), pr); err != nil {
		return internalError(ctx, "Failed to create password reset")
	}

	if err := h.mailer.SendPasswordResetCode(ctx.Request().Context(), user.Email, code); err != nil {
		_ = h.resets.Delete(ctx.Request().Context(), pr.ID)
		return internalError(ctx, "Failed to send reset email")
	}

	return ctx.JSON(http.StatusAccepted, passwordResetPendingResponse{
		ResetToken:  pr.ID,
		ResendAfter: pr.LastSentAt.Add(h.resendCooldown()),
		ExpiresAt:   pr.ExpiresAt,
	})
}

// Resend handles POST /api/password-reset/resend-code.
func (h *PasswordResetHandler) Resend(ctx echo.Context) error {
	var req passwordResetResendBody
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}
	if req.ResetToken == uuid.Nil {
		return badRequest(ctx, "reset_token is required")
	}

	pr, err := h.resets.GetByID(ctx.Request().Context(), req.ResetToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apiError(ctx, http.StatusNotFound, "not_found", "Reset token not found")
		}
		return internalError(ctx, "Failed to load password reset")
	}

	now := h.now().UTC()
	if now.After(pr.ExpiresAt) {
		_ = h.resets.Delete(ctx.Request().Context(), pr.ID)
		return apiError(ctx, http.StatusGone, "expired", "Reset token expired")
	}

	cooldown := h.resendCooldown()
	resendAfter := pr.LastSentAt.Add(cooldown)
	if now.Before(resendAfter) {
		return tooManyRequests(ctx, resendAfter)
	}

	user, err := h.users.GetUserByID(ctx.Request().Context(), pr.UserID)
	if err != nil {
		return internalError(ctx, "Failed to load user")
	}

	code, err := generateCode()
	if err != nil {
		return internalError(ctx, "Failed to generate reset code")
	}
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return internalError(ctx, "Failed to hash reset code")
	}

	pr.CodeHash = string(codeHash)
	pr.Attempts = 0
	pr.LastSentAt = now
	pr.ExpiresAt = now.Add(h.codeTTL())
	if err := h.resets.Update(ctx.Request().Context(), pr); err != nil {
		return internalError(ctx, "Failed to update password reset")
	}

	if err := h.mailer.SendPasswordResetCode(ctx.Request().Context(), user.Email, code); err != nil {
		return internalError(ctx, "Failed to send reset email")
	}

	return ctx.JSON(http.StatusOK, passwordResetPendingResponse{
		ResetToken:  pr.ID,
		ResendAfter: pr.LastSentAt.Add(cooldown),
		ExpiresAt:   pr.ExpiresAt,
	})
}

// Confirm handles POST /api/password-reset/confirm.
func (h *PasswordResetHandler) Confirm(ctx echo.Context) error {
	var req passwordResetConfirmBody
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}
	if req.ResetToken == uuid.Nil || req.Code == "" || req.NewPassword == "" {
		return badRequest(ctx, "reset_token, code and new_password are required")
	}

	pr, err := h.resets.GetByID(ctx.Request().Context(), req.ResetToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apiError(ctx, http.StatusNotFound, "not_found", "Reset token not found")
		}
		return internalError(ctx, "Failed to load password reset")
	}

	now := h.now().UTC()
	if now.After(pr.ExpiresAt) {
		_ = h.resets.Delete(ctx.Request().Context(), pr.ID)
		return apiError(ctx, http.StatusGone, "expired", "Reset token expired")
	}
	if h.cfg.MaxAttempts > 0 && pr.Attempts >= h.cfg.MaxAttempts {
		_ = h.resets.Delete(ctx.Request().Context(), pr.ID)
		return apiError(ctx, http.StatusTooManyRequests, "too_many_attempts",
			"Too many attempts; restart password reset")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pr.CodeHash), []byte(req.Code)); err != nil {
		pr.Attempts++
		_ = h.resets.Update(ctx.Request().Context(), pr)
		return apiError(ctx, http.StatusUnauthorized, "invalid_code", "Invalid reset code")
	}

	user, err := h.users.GetUserByID(ctx.Request().Context(), pr.UserID)
	if err != nil {
		return internalError(ctx, "Failed to load user")
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return internalError(ctx, "Failed to hash password")
	}
	user.PasswordHash = string(pwdHash)
	if err := h.users.UpdateUser(ctx.Request().Context(), user); err != nil {
		return internalError(ctx, "Failed to update user")
	}

	// Invalidate all existing sessions and the reset token itself.
	_ = h.sessions.DeleteSessionsByUserID(ctx.Request().Context(), user.ID)
	_ = h.resets.Delete(ctx.Request().Context(), pr.ID)

	return ctx.NoContent(http.StatusNoContent)
}

func (h *PasswordResetHandler) codeTTL() time.Duration {
	if h.cfg.CodeTTL > 0 {
		return h.cfg.CodeTTL
	}
	return 15 * time.Minute
}

func (h *PasswordResetHandler) resendCooldown() time.Duration {
	if h.cfg.ResendCooldown > 0 {
		return h.cfg.ResendCooldown
	}
	return 60 * time.Second
}
