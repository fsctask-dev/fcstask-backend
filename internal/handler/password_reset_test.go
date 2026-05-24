package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask-backend/internal/config"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

// --- mocks ---

type MockPasswordResetRepo struct {
	mock.Mock
}

func (m *MockPasswordResetRepo) Create(ctx context.Context, p *models.PasswordReset) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}
func (m *MockPasswordResetRepo) Update(ctx context.Context, p *models.PasswordReset) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}
func (m *MockPasswordResetRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.PasswordReset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PasswordReset), args.Error(1)
}
func (m *MockPasswordResetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockPasswordResetRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
func (m *MockPasswordResetRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

var _ repo.IPasswordResetRepo = (*MockPasswordResetRepo)(nil)

// --- helpers ---

type pwResetEnv struct {
	h      *PasswordResetHandler
	users  *MockUserRepository
	sess   *MockSessionRepository
	resets *MockPasswordResetRepo
	mailer *fakeMailer
	now    time.Time
}

func newPwResetEnv(t *testing.T) pwResetEnv {
	t.Helper()
	users := new(MockUserRepository)
	sess := new(MockSessionRepository)
	resets := new(MockPasswordResetRepo)
	mailer := &fakeMailer{}
	cfg := config.MailerConfig{
		CodeTTL:        15 * time.Minute,
		ResendCooldown: 60 * time.Second,
		MaxAttempts:    5,
	}
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	h := NewPasswordResetHandler(users, sess, resets, mailer, cfg)
	h.now = func() time.Time { return now }
	return pwResetEnv{h, users, sess, resets, mailer, now}
}

// --- Request ---

func TestPasswordReset_Request_Success(t *testing.T) {
	env := newPwResetEnv(t)
	userID := uuid.New()
	user := &models.User{ID: userID, Email: "ivan@example.com"}
	env.users.On("GetUserByEmail", mock.Anything, "ivan@example.com").Return(user, nil)
	env.resets.On("DeleteByUserID", mock.Anything, userID).Return(nil)
	resetID := uuid.New()
	env.resets.On("Create", mock.Anything, mock.MatchedBy(func(p *models.PasswordReset) bool {
		return p.UserID == userID && p.CodeHash != ""
	})).Return(nil).Run(func(args mock.Arguments) {
		p := args.Get(1).(*models.PasswordReset)
		p.ID = resetID
	})

	rec := postJSON(t, "/api/password-reset/request", map[string]any{
		"email": "ivan@example.com",
	}, env.h.Request)
	assert.Equal(t, http.StatusAccepted, rec.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, resetID.String(), resp["reset_token"])
	assert.Len(t, env.mailer.passwordResetCodes, 1)
}

func TestPasswordReset_Request_UserNotFound(t *testing.T) {
	env := newPwResetEnv(t)
	env.users.On("GetUserByEmail", mock.Anything, "ghost@example.com").Return(nil, gorm.ErrRecordNotFound)

	rec := postJSON(t, "/api/password-reset/request", map[string]any{
		"email": "ghost@example.com",
	}, env.h.Request)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	env.resets.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// --- Resend ---

func TestPasswordReset_Resend_RespectsCooldown(t *testing.T) {
	env := newPwResetEnv(t)
	resetID := uuid.New()
	env.resets.On("GetByID", mock.Anything, resetID).Return(&models.PasswordReset{
		ID:         resetID,
		LastSentAt: env.now.Add(-10 * time.Second),
		ExpiresAt:  env.now.Add(10 * time.Minute),
	}, nil)

	rec := postJSON(t, "/api/password-reset/resend-code", map[string]any{
		"reset_token": resetID.String(),
	}, env.h.Resend)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Empty(t, env.mailer.passwordResetCodes)
}

func TestPasswordReset_Resend_Success(t *testing.T) {
	env := newPwResetEnv(t)
	resetID := uuid.New()
	userID := uuid.New()
	pr := &models.PasswordReset{
		ID:         resetID,
		UserID:     userID,
		LastSentAt: env.now.Add(-2 * time.Minute),
		ExpiresAt:  env.now.Add(10 * time.Minute),
		Attempts:   2,
	}
	env.resets.On("GetByID", mock.Anything, resetID).Return(pr, nil)
	env.users.On("GetUserByID", mock.Anything, userID).Return(&models.User{ID: userID, Email: "ivan@example.com"}, nil)
	env.resets.On("Update", mock.Anything, mock.MatchedBy(func(p *models.PasswordReset) bool {
		return p.Attempts == 0 && p.LastSentAt.Equal(env.now)
	})).Return(nil)

	rec := postJSON(t, "/api/password-reset/resend-code", map[string]any{
		"reset_token": resetID.String(),
	}, env.h.Resend)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, env.mailer.passwordResetCodes, 1)
}

// --- Confirm ---

func TestPasswordReset_Confirm_Success_InvalidatesSessions(t *testing.T) {
	env := newPwResetEnv(t)
	resetID := uuid.New()
	userID := uuid.New()

	code := "654321"
	codeHash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.MinCost)
	pr := &models.PasswordReset{
		ID:        resetID,
		UserID:    userID,
		CodeHash:  string(codeHash),
		ExpiresAt: env.now.Add(15 * time.Minute),
	}
	env.resets.On("GetByID", mock.Anything, resetID).Return(pr, nil)
	user := &models.User{ID: userID, Email: "ivan@example.com", PasswordHash: "old"}
	env.users.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	env.users.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.PasswordHash != "" && u.PasswordHash != "old"
	})).Return(nil)
	env.sess.On("DeleteSessionsByUserID", mock.Anything, userID).Return(nil)
	env.resets.On("Delete", mock.Anything, resetID).Return(nil)

	rec := postJSON(t, "/api/password-reset/confirm", map[string]any{
		"reset_token":  resetID.String(),
		"code":         code,
		"new_password": "newPass456",
	}, env.h.Confirm)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	env.sess.AssertExpectations(t)
}

func TestPasswordReset_Confirm_WrongCode_IncrementsAttempts(t *testing.T) {
	env := newPwResetEnv(t)
	resetID := uuid.New()
	codeHash, _ := bcrypt.GenerateFromPassword([]byte("111111"), bcrypt.MinCost)
	pr := &models.PasswordReset{
		ID:        resetID,
		UserID:    uuid.New(),
		CodeHash:  string(codeHash),
		ExpiresAt: env.now.Add(15 * time.Minute),
	}
	env.resets.On("GetByID", mock.Anything, resetID).Return(pr, nil)
	env.resets.On("Update", mock.Anything, mock.MatchedBy(func(p *models.PasswordReset) bool {
		return p.Attempts == 1
	})).Return(nil)

	rec := postJSON(t, "/api/password-reset/confirm", map[string]any{
		"reset_token":  resetID.String(),
		"code":         "999999",
		"new_password": "newPass",
	}, env.h.Confirm)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	env.users.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestPasswordReset_Confirm_Expired(t *testing.T) {
	env := newPwResetEnv(t)
	resetID := uuid.New()
	env.resets.On("GetByID", mock.Anything, resetID).Return(&models.PasswordReset{
		ID:        resetID,
		ExpiresAt: env.now.Add(-time.Second),
	}, nil)
	env.resets.On("Delete", mock.Anything, resetID).Return(nil)

	rec := postJSON(t, "/api/password-reset/confirm", map[string]any{
		"reset_token":  resetID.String(),
		"code":         "123456",
		"new_password": "x",
	}, env.h.Confirm)
	assert.Equal(t, http.StatusGone, rec.Code)
}

func TestPasswordReset_Confirm_TooManyAttempts(t *testing.T) {
	env := newPwResetEnv(t)
	resetID := uuid.New()
	env.resets.On("GetByID", mock.Anything, resetID).Return(&models.PasswordReset{
		ID:        resetID,
		Attempts:  5,
		ExpiresAt: env.now.Add(time.Minute),
	}, nil)
	env.resets.On("Delete", mock.Anything, resetID).Return(nil)

	rec := postJSON(t, "/api/password-reset/confirm", map[string]any{
		"reset_token":  resetID.String(),
		"code":         "123456",
		"new_password": "x",
	}, env.h.Confirm)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}
