package handler

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
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

type SignUpHandler struct {
	users    repo.IUserRepo
	sessions repo.SessionRepositoryInterface
	regs     repo.IEmailRegistrationRepo
	mailer   mailer.Mailer
	cfg      config.MailerConfig
	now      func() time.Time
}

func NewSignUpHandler(
	users repo.IUserRepo,
	sessions repo.SessionRepositoryInterface,
	regs repo.IEmailRegistrationRepo,
	m mailer.Mailer,
	cfg config.MailerConfig,
) *SignUpHandler {
	return &SignUpHandler{
		users:    users,
		sessions: sessions,
		regs:     regs,
		mailer:   m,
		cfg:      cfg,
		now:      time.Now,
	}
}

// Request/response shapes are local — see oauth.go for the reason we don't reuse api.*.

type signUpSubmitRequest struct {
	Email     string  `json:"email"`
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

type signUpPendingResponse struct {
	VerificationToken uuid.UUID `json:"verification_token"`
	ResendAfter       time.Time `json:"resend_after"`
	ExpiresAt         time.Time `json:"expires_at"`
}

type signUpVerifyRequest struct {
	VerificationToken uuid.UUID `json:"verification_token"`
	Code              string    `json:"code"`
}

type signUpResendRequest struct {
	VerificationToken uuid.UUID `json:"verification_token"`
}

// Submit handles POST /api/signup. The User row is NOT created here — only an
// EmailRegistration with a hashed code. The frontend gets a verification_token
// to use when calling /api/signup/verify.
func (h *SignUpHandler) Submit(ctx echo.Context) error {
	var req signUpSubmitRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return badRequest(ctx, "Username, password and email are required")
	}

	if exists, err := h.users.ExistsByEmail(ctx.Request().Context(), req.Email); err != nil {
		return internalError(ctx, "Failed to check email uniqueness")
	} else if exists {
		return conflict(ctx, "User with this email already exists")
	}
	if exists, err := h.users.ExistsByUsername(ctx.Request().Context(), req.Username); err != nil {
		return internalError(ctx, "Failed to check username uniqueness")
	} else if exists {
		return conflict(ctx, "User with this username already exists")
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return internalError(ctx, "Failed to hash password")
	}

	code, err := generateCode()
	if err != nil {
		return internalError(ctx, "Failed to generate verification code")
	}
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return internalError(ctx, "Failed to hash verification code")
	}

	now := h.now().UTC()
	reg := &models.EmailRegistration{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(pwdHash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		CodeHash:     string(codeHash),
		LastSentAt:   now,
		ExpiresAt:    now.Add(h.codeTTL()),
	}

	// Drop any prior pending registration for this email so a fresh /signup
	// always wins (mirrors the OAuth registration_session pattern).
	_ = h.regs.DeleteByEmail(ctx.Request().Context(), req.Email)

	if err := h.regs.Create(ctx.Request().Context(), reg); err != nil {
		return internalError(ctx, "Failed to create registration")
	}

	if err := h.mailer.SendVerificationCode(ctx.Request().Context(), req.Email, req.Username, code); err != nil {
		_ = h.regs.Delete(ctx.Request().Context(), reg.ID)
		return internalError(ctx, "Failed to send verification email")
	}

	return ctx.JSON(http.StatusAccepted, signUpPendingResponse{
		VerificationToken: reg.ID,
		ResendAfter:       reg.LastSentAt.Add(h.resendCooldown()),
		ExpiresAt:         reg.ExpiresAt,
	})
}

// Verify handles POST /api/signup/verify.
func (h *SignUpHandler) Verify(ctx echo.Context) error {
	var req signUpVerifyRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}
	if req.VerificationToken == uuid.Nil || req.Code == "" {
		return badRequest(ctx, "verification_token and code are required")
	}

	reg, err := h.regs.GetByID(ctx.Request().Context(), req.VerificationToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apiError(ctx, http.StatusNotFound, "not_found", "Verification token not found")
		}
		return internalError(ctx, "Failed to load registration")
	}

	now := h.now().UTC()
	if now.After(reg.ExpiresAt) {
		_ = h.regs.Delete(ctx.Request().Context(), reg.ID)
		return apiError(ctx, http.StatusGone, "expired", "Verification token expired")
	}

	if h.cfg.MaxAttempts > 0 && reg.Attempts >= h.cfg.MaxAttempts {
		_ = h.regs.Delete(ctx.Request().Context(), reg.ID)
		return apiError(ctx, http.StatusTooManyRequests, "too_many_attempts",
			"Too many attempts; restart signup")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(reg.CodeHash), []byte(req.Code)); err != nil {
		reg.Attempts++
		_ = h.regs.Update(ctx.Request().Context(), reg)
		return apiError(ctx, http.StatusUnauthorized, "invalid_code", "Invalid verification code")
	}

	user := &models.User{
		Email:        reg.Email,
		Username:     reg.Username,
		PasswordHash: reg.PasswordHash,
		FirstName:    reg.FirstName,
		LastName:     reg.LastName,
		UserID:       uuid.New(),
	}
	if err := h.users.Create(ctx.Request().Context(), user); err != nil {
		if col := uniqueConstraintColumn(err); col != "" {
			return conflict(ctx, "User with this "+col+" already exists")
		}
		return internalError(ctx, "Failed to create user")
	}

	_ = h.regs.Delete(ctx.Request().Context(), reg.ID)

	session := &models.Session{
		UserID:    user.ID,
		IP:        ctx.RealIP(),
		UserAgent: ctx.Request().UserAgent(),
	}
	if err := h.sessions.Create(ctx.Request().Context(), session); err != nil {
		return internalError(ctx, "Failed to create session")
	}

	return ctx.JSON(http.StatusCreated, buildAuthResponse(user, session))
}

// Resend handles POST /api/signup/resend-code.
func (h *SignUpHandler) Resend(ctx echo.Context) error {
	var req signUpResendRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "Invalid request body")
	}
	if req.VerificationToken == uuid.Nil {
		return badRequest(ctx, "verification_token is required")
	}

	reg, err := h.regs.GetByID(ctx.Request().Context(), req.VerificationToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apiError(ctx, http.StatusNotFound, "not_found", "Verification token not found")
		}
		return internalError(ctx, "Failed to load registration")
	}

	now := h.now().UTC()
	if now.After(reg.ExpiresAt) {
		_ = h.regs.Delete(ctx.Request().Context(), reg.ID)
		return apiError(ctx, http.StatusGone, "expired", "Verification token expired")
	}

	cooldown := h.resendCooldown()
	resendAfter := reg.LastSentAt.Add(cooldown)
	if now.Before(resendAfter) {
		return tooManyRequests(ctx, resendAfter)
	}

	code, err := generateCode()
	if err != nil {
		return internalError(ctx, "Failed to generate verification code")
	}
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return internalError(ctx, "Failed to hash verification code")
	}

	reg.CodeHash = string(codeHash)
	reg.Attempts = 0
	reg.LastSentAt = now
	reg.ExpiresAt = now.Add(h.codeTTL())
	if err := h.regs.Update(ctx.Request().Context(), reg); err != nil {
		return internalError(ctx, "Failed to update registration")
	}

	if err := h.mailer.SendVerificationCode(ctx.Request().Context(), reg.Email, reg.Username, code); err != nil {
		return internalError(ctx, "Failed to send verification email")
	}

	return ctx.JSON(http.StatusOK, signUpPendingResponse{
		VerificationToken: reg.ID,
		ResendAfter:       reg.LastSentAt.Add(cooldown),
		ExpiresAt:         reg.ExpiresAt,
	})
}

func (h *SignUpHandler) codeTTL() time.Duration {
	if h.cfg.CodeTTL > 0 {
		return h.cfg.CodeTTL
	}
	return 15 * time.Minute
}

func (h *SignUpHandler) resendCooldown() time.Duration {
	if h.cfg.ResendCooldown > 0 {
		return h.cfg.ResendCooldown
	}
	return 60 * time.Second
}

// generateCode returns a uniformly distributed 6-digit numeric code.
func generateCode() (string, error) {
	max := big.NewInt(1_000_000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func tooManyRequests(ctx echo.Context, retryAt time.Time) error {
	secondsLeft := int(time.Until(retryAt).Round(time.Second).Seconds())
	if secondsLeft < 0 {
		secondsLeft = 0
	}
	ctx.Response().Header().Set("Retry-After", strconv.Itoa(secondsLeft))
	return ctx.JSONBlob(http.StatusTooManyRequests, mustJSON(map[string]any{
		"error": map[string]any{
			"code":         "rate_limited",
			"message":      "Try again later",
			"retry_after":  secondsLeft,
			"resend_after": retryAt,
		},
	}))
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
