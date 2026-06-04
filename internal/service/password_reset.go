package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask-backend/internal/config"
	"fcstask-backend/internal/db/model"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/mailer"
	"fcstask-backend/internal/metrics"
)

// PasswordResetService owns the "I forgot my password" flow: request a code,
// resend it, and confirm it to set a new password.
type PasswordResetService struct {
	userRepo          repo.IUserRepo
	passwordResetRepo repo.IPasswordResetRepository

	mailer mailer.Mailer

	passwordResetConfig config.EmailRegistrationConfig

	metrics *metrics.PasswordResetMetrics
}

func NewPasswordResetService(
	userRepo repo.IUserRepo,
	passwordResetRepo repo.IPasswordResetRepository,
	m mailer.Mailer,
	passwordResetConfig config.EmailRegistrationConfig,
) *PasswordResetService {
	return &PasswordResetService{
		userRepo:            userRepo,
		passwordResetRepo:   passwordResetRepo,
		mailer:              m,
		passwordResetConfig: passwordResetConfig,
	}
}

func (s *PasswordResetService) WithMetrics(m *metrics.PasswordResetMetrics) *PasswordResetService {
	s.metrics = m
	return s
}

type PasswordResetConfirmInput struct {
	Email       string
	Code        string
	Password    string
	MaxAttempts int
}

func (s *PasswordResetService) PasswordResetRequest(ctx context.Context, email string) (result *model.PasswordReset, err error) {
	defer func() { s.metrics.IncRequest(errorOutcome(err)) }()

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("email not found")
		}
		return nil, Internal("Failed to get user by email", err)
	}

	if err := s.passwordResetRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return nil, Internal("Failed to delete previous password reset", err)
	}

	code, err := generateCode()
	if err != nil {
		return nil, Internal("Failed to generate reset code", err)
	}
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return nil, Internal("Failed to hash reset code", err)
	}

	now := time.Now()
	result = &models.PasswordReset{
		UserID:     user.ID,
		User:       user,
		CodeHash:   string(codeHash),
		LastSentAt: now,
		ExpiresAt:  now.Add(s.passwordResetConfig.TTL),
	}
	if err := mailer.SendPasswordReset(s.mailer, ctx, result, code); err != nil {
		return nil, Internal("failed to send password reset email", err)
	}
	if err := s.passwordResetRepo.Create(ctx, result); err != nil {
		return nil, Internal("failed to create password reset", err)
	}
	return result, nil
}

func (s *PasswordResetService) PasswordResetResend(ctx context.Context, email string) (result *model.PasswordReset, err error) {
	defer func() { s.metrics.IncResend(errorOutcome(err)) }()

	result, err = s.passwordResetRepo.GetByUserEmail(ctx, email)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.PasswordResetRequest(ctx, email)
		}
		return nil, Internal("Failed to get user by email", err)
	}

	// GetByUserEmail does not preload the user; load it so the email can be sent.
	user, err := s.userRepo.GetUserByID(ctx, result.UserID)
	if err != nil {
		return nil, Internal("Failed to get user", err)
	}
	result.User = user

	code, err := generateCode()
	if err != nil {
		return nil, Internal("Failed to generate reset code", err)
	}
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return nil, Internal("Failed to hash reset code", err)
	}

	now := time.Now()
	result.CodeHash = string(codeHash)
	result.LastSentAt = now
	result.ExpiresAt = now.Add(s.passwordResetConfig.TTL)
	result.Attempts = 0

	if err := mailer.SendPasswordReset(s.mailer, ctx, result, code); err != nil {
		return nil, Internal("failed to send password reset email", err)
	}
	if err := s.passwordResetRepo.Update(ctx, result); err != nil {
		return nil, Internal("failed to update password reset", err)
	}
	return result, nil
}

func (s *PasswordResetService) PasswordResetConfirm(ctx context.Context, input PasswordResetConfirmInput) (err error) {
	defer func() { s.metrics.IncConfirm(errorOutcome(err)) }()

	passwordReset, err := s.passwordResetRepo.GetByUserEmail(ctx, input.Email)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound("password reset not found")
		}
		return Internal("Failed to get user by email", err)
	}

	now := time.Now()
	if passwordReset.ExpiresAt.Before(now) {
		return Unauthorized("password reset has expired")
	}

	if passwordReset.Attempts > input.MaxAttempts {
		return Unauthorized("number of attempts exceeded")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordReset.CodeHash), []byte(input.Code)); err != nil {
		passwordReset.Attempts++
		if err := s.passwordResetRepo.Update(ctx, passwordReset); err != nil {
			return Internal("failed to update password reset", err)
		}
		return Unauthorized("password codes do not match")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return Internal("Failed to hash password", err)
	}
	// GetByUserEmail does not preload the user, so load it by ID before updating.
	user, err := s.userRepo.GetUserByID(ctx, passwordReset.UserID)
	if err != nil {
		return Internal("Failed to get user", err)
	}
	user.PasswordHash = string(hash)
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return Internal("Failed to update password", err)
	}

	if err := s.passwordResetRepo.Delete(ctx, passwordReset.ID); err != nil {
		return Internal("failed to delete password reset", err)
	}
	return nil
}
