package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask-backend/internal/config"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

// --- mocks ---

type MockEmailRegistrationRepo struct {
	mock.Mock
}

func (m *MockEmailRegistrationRepo) Create(ctx context.Context, r *models.EmailRegistration) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *MockEmailRegistrationRepo) Update(ctx context.Context, r *models.EmailRegistration) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *MockEmailRegistrationRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.EmailRegistration, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EmailRegistration), args.Error(1)
}
func (m *MockEmailRegistrationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockEmailRegistrationRepo) DeleteByEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}
func (m *MockEmailRegistrationRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

var _ repo.IEmailRegistrationRepo = (*MockEmailRegistrationRepo)(nil)

// fakeMailer captures the codes the handler told us to send so tests can
// fish them back out and complete the flow.
type fakeMailer struct {
	mu                   sync.Mutex
	verificationCodes    []string
	passwordResetCodes   []string
	verifyErr            error
	passwordResetErr     error
}

func (f *fakeMailer) SendVerificationCode(_ context.Context, _, _, code string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.verifyErr != nil {
		return f.verifyErr
	}
	f.verificationCodes = append(f.verificationCodes, code)
	return nil
}

func (f *fakeMailer) SendPasswordResetCode(_ context.Context, _, code string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.passwordResetErr != nil {
		return f.passwordResetErr
	}
	f.passwordResetCodes = append(f.passwordResetCodes, code)
	return nil
}

func (f *fakeMailer) lastVerification() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.verificationCodes) == 0 {
		return ""
	}
	return f.verificationCodes[len(f.verificationCodes)-1]
}

// --- helpers ---

type signUpEnv struct {
	h      *SignUpHandler
	users  *MockUserRepository
	sess   *MockSessionRepository
	regs   *MockEmailRegistrationRepo
	mailer *fakeMailer
	now    time.Time
}

func newSignUpEnv(t *testing.T) signUpEnv {
	t.Helper()
	users := new(MockUserRepository)
	sess := new(MockSessionRepository)
	regs := new(MockEmailRegistrationRepo)
	mailer := &fakeMailer{}
	cfg := config.MailerConfig{
		CodeTTL:        15 * time.Minute,
		ResendCooldown: 60 * time.Second,
		MaxAttempts:    5,
	}
	now := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	h := NewSignUpHandler(users, sess, regs, mailer, cfg)
	h.now = func() time.Time { return now }
	return signUpEnv{h, users, sess, regs, mailer, now}
}

func postJSON(t *testing.T, path string, body any, fn func(echo.Context) error) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	assert.NoError(t, fn(ctx))
	return rec
}

// --- Submit ---

func TestSignUp_Submit_CreatesPendingRegistrationAndSendsCode(t *testing.T) {
	env := newSignUpEnv(t)

	env.users.On("ExistsByEmail", mock.Anything, "ivan@example.com").Return(false, nil)
	env.users.On("ExistsByUsername", mock.Anything, "ivan").Return(false, nil)
	env.regs.On("DeleteByEmail", mock.Anything, "ivan@example.com").Return(nil)
	regID := uuid.New()
	env.regs.On("Create", mock.Anything, mock.MatchedBy(func(r *models.EmailRegistration) bool {
		return r.Email == "ivan@example.com" && r.Username == "ivan" &&
			r.PasswordHash != "" && r.CodeHash != "" && !r.LastSentAt.IsZero()
	})).Return(nil).Run(func(args mock.Arguments) {
		r := args.Get(1).(*models.EmailRegistration)
		r.ID = regID
	})

	rec := postJSON(t, "/api/signup", map[string]any{
		"email":    "ivan@example.com",
		"username": "ivan",
		"password": "pass1234",
	}, env.h.Submit)
	assert.Equal(t, http.StatusAccepted, rec.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, regID.String(), resp["verification_token"])
	assert.NotEmpty(t, resp["resend_after"])
	assert.NotEmpty(t, resp["expires_at"])
	assert.NotEmpty(t, env.mailer.lastVerification())
}

func TestSignUp_Submit_EmailConflict(t *testing.T) {
	env := newSignUpEnv(t)
	env.users.On("ExistsByEmail", mock.Anything, "taken@example.com").Return(true, nil)

	rec := postJSON(t, "/api/signup", map[string]any{
		"email":    "taken@example.com",
		"username": "ivan",
		"password": "pass1234",
	}, env.h.Submit)
	assert.Equal(t, http.StatusConflict, rec.Code)
	env.regs.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSignUp_Submit_MissingFields(t *testing.T) {
	env := newSignUpEnv(t)
	rec := postJSON(t, "/api/signup", map[string]any{
		"email": "ivan@example.com",
	}, env.h.Submit)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Verify ---

func TestSignUp_Verify_Success(t *testing.T) {
	env := newSignUpEnv(t)

	code := "123456"
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.MinCost)
	assert.NoError(t, err)

	regID := uuid.New()
	reg := &models.EmailRegistration{
		ID:           regID,
		Email:        "ivan@example.com",
		Username:     "ivan",
		PasswordHash: "$2a$04$pretendpasswordhash",
		CodeHash:     string(codeHash),
		LastSentAt:   env.now,
		ExpiresAt:    env.now.Add(15 * time.Minute),
	}
	env.regs.On("GetByID", mock.Anything, regID).Return(reg, nil)

	createdUserID := uuid.New()
	createdSessID := uuid.New()
	env.users.On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "ivan@example.com" && u.Username == "ivan" && u.PasswordHash == reg.PasswordHash
	})).Return(nil).Run(func(args mock.Arguments) {
		u := args.Get(1).(*models.User)
		u.ID = createdUserID
	})
	env.regs.On("Delete", mock.Anything, regID).Return(nil)
	env.sess.On("Create", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
		return s.UserID == createdUserID
	})).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.Session)
		s.ID = createdSessID
	})

	rec := postJSON(t, "/api/signup/verify", map[string]any{
		"verification_token": regID.String(),
		"code":               code,
	}, env.h.Verify)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, createdSessID.String(), resp["session_token"])
}

func TestSignUp_Verify_WrongCode_IncrementsAttempts(t *testing.T) {
	env := newSignUpEnv(t)

	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	regID := uuid.New()
	reg := &models.EmailRegistration{
		ID:        regID,
		CodeHash:  string(codeHash),
		ExpiresAt: env.now.Add(15 * time.Minute),
		Attempts:  0,
	}
	env.regs.On("GetByID", mock.Anything, regID).Return(reg, nil)
	env.regs.On("Update", mock.Anything, mock.MatchedBy(func(r *models.EmailRegistration) bool {
		return r.Attempts == 1
	})).Return(nil)

	rec := postJSON(t, "/api/signup/verify", map[string]any{
		"verification_token": regID.String(),
		"code":               "999999",
	}, env.h.Verify)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	env.users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSignUp_Verify_TooManyAttempts(t *testing.T) {
	env := newSignUpEnv(t)
	regID := uuid.New()
	env.regs.On("GetByID", mock.Anything, regID).Return(&models.EmailRegistration{
		ID:        regID,
		Attempts:  5,
		ExpiresAt: env.now.Add(time.Minute),
	}, nil)
	env.regs.On("Delete", mock.Anything, regID).Return(nil)

	rec := postJSON(t, "/api/signup/verify", map[string]any{
		"verification_token": regID.String(),
		"code":               "123456",
	}, env.h.Verify)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestSignUp_Verify_Expired(t *testing.T) {
	env := newSignUpEnv(t)
	regID := uuid.New()
	env.regs.On("GetByID", mock.Anything, regID).Return(&models.EmailRegistration{
		ID:        regID,
		ExpiresAt: env.now.Add(-time.Second),
	}, nil)
	env.regs.On("Delete", mock.Anything, regID).Return(nil)

	rec := postJSON(t, "/api/signup/verify", map[string]any{
		"verification_token": regID.String(),
		"code":               "123456",
	}, env.h.Verify)
	assert.Equal(t, http.StatusGone, rec.Code)
}

func TestSignUp_Verify_TokenNotFound(t *testing.T) {
	env := newSignUpEnv(t)
	regID := uuid.New()
	env.regs.On("GetByID", mock.Anything, regID).Return(nil, gorm.ErrRecordNotFound)

	rec := postJSON(t, "/api/signup/verify", map[string]any{
		"verification_token": regID.String(),
		"code":               "123456",
	}, env.h.Verify)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- Resend ---

func TestSignUp_Resend_RespectsCooldown(t *testing.T) {
	env := newSignUpEnv(t)
	regID := uuid.New()
	// last sent 30s ago — cooldown is 60s, so should reject
	env.regs.On("GetByID", mock.Anything, regID).Return(&models.EmailRegistration{
		ID:         regID,
		LastSentAt: env.now.Add(-30 * time.Second),
		ExpiresAt:  env.now.Add(10 * time.Minute),
	}, nil)

	rec := postJSON(t, "/api/signup/resend-code", map[string]any{
		"verification_token": regID.String(),
	}, env.h.Resend)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Retry-After"))
	assert.Empty(t, env.mailer.verificationCodes)
}

func TestSignUp_Resend_AfterCooldown_SendsNewCode(t *testing.T) {
	env := newSignUpEnv(t)
	regID := uuid.New()
	reg := &models.EmailRegistration{
		ID:         regID,
		Email:      "ivan@example.com",
		Username:   "ivan",
		LastSentAt: env.now.Add(-2 * time.Minute),
		ExpiresAt:  env.now.Add(10 * time.Minute),
		Attempts:   3,
	}
	env.regs.On("GetByID", mock.Anything, regID).Return(reg, nil)
	env.regs.On("Update", mock.Anything, mock.MatchedBy(func(r *models.EmailRegistration) bool {
		// new code should reset attempts and set LastSentAt to "now"
		return r.Attempts == 0 && r.LastSentAt.Equal(env.now) && r.CodeHash != ""
	})).Return(nil)

	rec := postJSON(t, "/api/signup/resend-code", map[string]any{
		"verification_token": regID.String(),
	}, env.h.Resend)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, env.mailer.verificationCodes, 1)
}

func TestSignUp_Resend_Expired(t *testing.T) {
	env := newSignUpEnv(t)
	regID := uuid.New()
	env.regs.On("GetByID", mock.Anything, regID).Return(&models.EmailRegistration{
		ID:        regID,
		ExpiresAt: env.now.Add(-time.Minute),
	}, nil)
	env.regs.On("Delete", mock.Anything, regID).Return(nil)

	rec := postJSON(t, "/api/signup/resend-code", map[string]any{
		"verification_token": regID.String(),
	}, env.h.Resend)
	assert.Equal(t, http.StatusGone, rec.Code)
}
