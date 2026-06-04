package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/openframebox/gomail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/config"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/service"
)

// MockSessionRepository мок для репозитория сессий
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockSessionRepository) TouchSessionAccessedAt(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSessionsWithUser(ctx context.Context, limit, offset int) ([]models.Session, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockSessionRepository) CountSessions(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSessionRepository) CleanOutdatedSessions(ctx context.Context, ttl time.Duration) (int64, error) {
	args := m.Called(ctx, ttl)
	return args.Get(0).(int64), args.Error(1)
}

var _ repo.ISessionRepository = (*MockSessionRepository)(nil)

var (
	testUserUUID1    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testUserUUID2    = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testUserUUID99   = uuid.MustParse("99999999-9999-9999-9999-999999999999")
	testSessionUUID1 = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testSessionUUID2 = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	testSessionUUID3 = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
)

// stubEmailRegRepo is a minimal IEmailRegistrationRepo for handler tests.
type stubEmailRegRepo struct {
	reg       *models.EmailRegistration
	createErr error
}

func (r *stubEmailRegRepo) Create(ctx context.Context, reg *models.EmailRegistration) error {
	if r.createErr != nil {
		return r.createErr
	}
	if reg.ID == uuid.Nil {
		reg.ID = uuid.New()
	}
	r.reg = reg
	return nil
}

func (r *stubEmailRegRepo) Update(ctx context.Context, reg *models.EmailRegistration) error {
	r.reg = reg
	return nil
}

func (r *stubEmailRegRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.EmailRegistration, error) {
	if r.reg == nil || r.reg.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	return r.reg, nil
}

func (r *stubEmailRegRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.reg = nil
	return nil
}

func (r *stubEmailRegRepo) DeleteByEmail(ctx context.Context, email string) error { return nil }

func (r *stubEmailRegRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

// stubMailer satisfies mailer.Mailer for handler tests.
type stubMailer struct{ err error }

func (m stubMailer) Send(ctx context.Context, to gomail.Address, subject, body string) error {
	return m.err
}

// stubOAuthRepo is a no-op IOAuthIdentityRepo; SignUpVerify consults it after
// creating the user. GetByEmailRegistrationID returns a nil error so the verify
// path treats it as "nothing to relink".
type stubOAuthRepo struct{}

func (stubOAuthRepo) Create(ctx context.Context, i *models.OAuthIdentity) error { return nil }
func (stubOAuthRepo) Update(ctx context.Context, i *models.OAuthIdentity) error { return nil }
func (stubOAuthRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.OAuthIdentity, error) {
	return nil, gorm.ErrRecordNotFound
}
func (stubOAuthRepo) GetByProviderUID(ctx context.Context, provider, uid string) (*models.OAuthIdentity, error) {
	return nil, gorm.ErrRecordNotFound
}
func (stubOAuthRepo) GetProviderForUserID(ctx context.Context, userID uuid.UUID, provider string) (*models.OAuthIdentity, error) {
	return nil, gorm.ErrRecordNotFound
}
func (stubOAuthRepo) GetByEmailRegistrationID(ctx context.Context, id uuid.UUID) (*models.OAuthIdentity, error) {
	return &models.OAuthIdentity{}, nil
}
func (stubOAuthRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.OAuthIdentity, error) {
	return nil, nil
}
func (stubOAuthRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// === SignUp ===

func TestAuthHandler_SignUp_Success(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	emailRepo := &stubEmailRegRepo{}

	reqBody := api.SignUpRequest{
		Email:    "new@example.com",
		Username: "newuser",
		Password: "secret123",
	}
	reqJSON, _ := json.Marshal(reqBody)

	mockUserRepo.On("ExistsUserByEmail", mock.Anything, string(reqBody.Email)).Return(false, nil)
	mockUserRepo.On("ExistsUserByUsername", mock.Anything, reqBody.Username).Return(false, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	svc := service.NewAuthService(mockUserRepo, mockSessionRepo, emailRepo, nil, stubMailer{}, config.EmailRegistrationConfig{})
	err := NewAuthHandler(svc).SignUp(ctx)

	assert.NoError(t, err)
	// Registration is pending email verification: 202, no user created yet.
	assert.Equal(t, http.StatusAccepted, rec.Code)

	var resp api.SignUpPendingResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NotEqual(t, uuid.Nil, uuid.UUID(resp.VerificationToken))
	assert.NotNil(t, emailRepo.reg)
	assert.Equal(t, "new@example.com", emailRepo.reg.Email)
	assert.Equal(t, "newuser", emailRepo.reg.Username)
	assert.NotEmpty(t, emailRepo.reg.PasswordHash)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthHandler_SignUp_InvalidJSON(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignUp(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "bad_request", resp.Error.Code)
}

func TestAuthHandler_SignUp_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		reqBody api.SignUpRequest
	}{
		{
			name:    "missing password",
			reqBody: api.SignUpRequest{Email: "a@b.com", Username: "user"},
		},
		{
			name:    "missing email",
			reqBody: api.SignUpRequest{Username: "user", Password: "pass"},
		},
		{
			name:    "missing username",
			reqBody: api.SignUpRequest{Email: "a@b.com", Password: "pass"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			mockUserRepo := new(MockUserRepository)
			mockSessionRepo := new(MockSessionRepository)

			reqJSON, _ := json.Marshal(tc.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignUp(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var resp api.Error
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, "bad_request", resp.Error.Code)
		})
	}
}

func TestAuthHandler_SignUp_EmailConflict(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	reqBody := api.SignUpRequest{Email: "taken@example.com", Username: "newuser", Password: "pass"}
	reqJSON, _ := json.Marshal(reqBody)

	mockUserRepo.On("ExistsUserByEmail", mock.Anything, "taken@example.com").Return(true, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignUp(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "conflict", resp.Error.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthHandler_SignUp_UsernameConflict(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	reqBody := api.SignUpRequest{Email: "new@example.com", Username: "taken", Password: "pass"}
	reqJSON, _ := json.Marshal(reqBody)

	mockUserRepo.On("ExistsUserByEmail", mock.Anything, "new@example.com").Return(false, nil)
	mockUserRepo.On("ExistsUserByUsername", mock.Anything, "taken").Return(true, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignUp(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "conflict", resp.Error.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthHandler_SignUp_RegistrationCreateError(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	emailRepo := &stubEmailRegRepo{createErr: errors.New("db error")}

	reqBody := api.SignUpRequest{Email: "new@example.com", Username: "newuser", Password: "pass"}
	reqJSON, _ := json.Marshal(reqBody)

	mockUserRepo.On("ExistsUserByEmail", mock.Anything, "new@example.com").Return(false, nil)
	mockUserRepo.On("ExistsUserByUsername", mock.Anything, "newuser").Return(false, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	svc := service.NewAuthService(mockUserRepo, mockSessionRepo, emailRepo, nil, stubMailer{}, config.EmailRegistrationConfig{})
	err := NewAuthHandler(svc).SignUp(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "internal_error", resp.Error.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthHandler_SignUp_MailerError(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	emailRepo := &stubEmailRegRepo{}

	reqBody := api.SignUpRequest{Email: "new@example.com", Username: "newuser", Password: "pass"}
	reqJSON, _ := json.Marshal(reqBody)

	mockUserRepo.On("ExistsUserByEmail", mock.Anything, "new@example.com").Return(false, nil)
	mockUserRepo.On("ExistsUserByUsername", mock.Anything, "newuser").Return(false, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	svc := service.NewAuthService(mockUserRepo, mockSessionRepo, emailRepo, nil, stubMailer{err: errors.New("smtp down")}, config.EmailRegistrationConfig{})
	err := NewAuthHandler(svc).SignUp(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	// The verification email failed, so no pending registration is persisted.
	assert.Nil(t, emailRepo.reg)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthHandler_SignUpVerify_Success(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	const code = "123456"
	codeHash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.MinCost)
	pwHash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	regID := uuid.New()
	emailRepo := &stubEmailRegRepo{reg: &models.EmailRegistration{
		ID:           regID,
		Email:        "new@example.com",
		Username:     "newuser",
		PasswordHash: string(pwHash),
		CodeHash:     string(codeHash),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}}

	mockUserRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "new@example.com" && u.Username == "newuser"
	})).Return(nil).Run(func(args mock.Arguments) {
		args.Get(1).(*models.User).ID = testUserUUID1
	})
	mockSessionRepo.On("CreateSession", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
		return s.UserID == testUserUUID1
	})).Return(nil).Run(func(args mock.Arguments) {
		args.Get(1).(*models.Session).ID = testSessionUUID1
	})

	reqJSON, _ := json.Marshal(api.SignUpVerifyRequest{VerificationToken: regID, Code: code})
	req := httptest.NewRequest(http.MethodPost, "/api/signup/verify", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	svc := service.NewAuthService(
		mockUserRepo, mockSessionRepo, emailRepo, stubOAuthRepo{}, stubMailer{},
		config.EmailRegistrationConfig{},
	)
	err := NewAuthHandler(svc).SignUpVerify(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.AuthResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "new@example.com", string(resp.User.Email))
	assert.Equal(t, testSessionUUID1, uuid.UUID(resp.SessionToken))

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthHandler_SignUpVerify_WrongCode(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	regID := uuid.New()
	emailRepo := &stubEmailRegRepo{reg: &models.EmailRegistration{
		ID:        regID,
		CodeHash:  string(codeHash),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}}

	reqJSON, _ := json.Marshal(api.SignUpVerifyRequest{VerificationToken: regID, Code: "000000"})
	req := httptest.NewRequest(http.MethodPost, "/api/signup/verify", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	svc := service.NewAuthService(mockUserRepo, mockSessionRepo, emailRepo, nil, stubMailer{}, config.EmailRegistrationConfig{})
	err := NewAuthHandler(svc).SignUpVerify(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_SignUpResendCode_Success(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	regID := uuid.New()
	emailRepo := &stubEmailRegRepo{reg: &models.EmailRegistration{
		ID:        regID,
		Email:     "new@example.com",
		Username:  "newuser",
		ExpiresAt: time.Now().Add(time.Minute),
	}}

	reqJSON, _ := json.Marshal(api.SignUpResendRequest{VerificationToken: regID})
	req := httptest.NewRequest(http.MethodPost, "/api/signup/resend-code", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	svc := service.NewAuthService(mockUserRepo, mockSessionRepo, emailRepo, nil, stubMailer{}, config.EmailRegistrationConfig{})
	err := NewAuthHandler(svc).SignUpResendCode(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.SignUpPendingResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, regID, uuid.UUID(resp.VerificationToken))
}

// === SignIn ===

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	assert.NoError(t, err)
	return string(h)
}

func TestAuthHandler_SignIn_SuccessWithEmail(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	now := time.Now()
	email := "test@example.com"
	testUser := &models.User{
		ID:           testUserUUID1,
		Email:        email,
		Username:     "testuser",
		PasswordHash: hashPassword(t, "correctpass"),
		UserID:       testUserUUID1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	body := map[string]interface{}{
		"email":    email,
		"password": "correctpass",
	}
	reqJSON, _ := json.Marshal(body)

	mockUserRepo.On("GetUserByEmail", mock.Anything, email).Return(testUser, nil)
	mockSessionRepo.On("CreateSession", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
		return s.UserID == testUserUUID1
	})).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.Session)
		s.ID = testSessionUUID2
	})

	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AuthResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "testuser", resp.User.Username)
	assert.Equal(t, testSessionUUID2, uuid.UUID(resp.SessionToken))

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthHandler_SignIn_SuccessWithUsername(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	now := time.Now()
	testUser := &models.User{
		ID:           testUserUUID2,
		Email:        "user@example.com",
		Username:     "myuser",
		PasswordHash: hashPassword(t, "mypass"),
		UserID:       testUserUUID2,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	body := map[string]interface{}{
		"username": "myuser",
		"password": "mypass",
	}
	reqJSON, _ := json.Marshal(body)

	mockUserRepo.On("GetUserByUsername", mock.Anything, "myuser").Return(testUser, nil)
	mockSessionRepo.On("CreateSession", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.Session)
		s.ID = testSessionUUID3
	})

	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AuthResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "myuser", resp.User.Username)
	assert.Equal(t, testSessionUUID3, uuid.UUID(resp.SessionToken))

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthHandler_SignIn_InvalidJSON(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_SignIn_MissingPassword(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	body := map[string]interface{}{"email": "a@b.com"}
	reqJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "bad_request", resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Password")
}

func TestAuthHandler_SignIn_MissingEmailAndUsername(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	body := map[string]interface{}{"password": "pass"}
	reqJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Contains(t, resp.Error.Message, "Email or username")
}

func TestAuthHandler_SignIn_UserNotFound(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	body := map[string]interface{}{"email": "no@example.com", "password": "pass"}
	reqJSON, _ := json.Marshal(body)

	mockUserRepo.On("GetUserByEmail", mock.Anything, "no@example.com").Return(nil, gorm.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "unauthorized", resp.Error.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthHandler_SignIn_WrongPassword(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	testUser := &models.User{
		ID:           testUserUUID1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: hashPassword(t, "correctpass"),
		UserID:       testUserUUID1,
	}

	body := map[string]interface{}{"email": "test@example.com", "password": "wrongpass"}
	reqJSON, _ := json.Marshal(body)

	mockUserRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "unauthorized", resp.Error.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthHandler_SignIn_DatabaseError(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	body := map[string]interface{}{"email": "test@example.com", "password": "pass"}
	reqJSON, _ := json.Marshal(body)

	mockUserRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthHandler_SignIn_CreateSessionError(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	testUser := &models.User{
		ID:           testUserUUID1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: hashPassword(t, "pass"),
		UserID:       testUserUUID1,
	}

	body := map[string]interface{}{"email": "test@example.com", "password": "pass"}
	reqJSON, _ := json.Marshal(body)

	mockUserRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
	mockSessionRepo.On("CreateSession", mock.Anything, mock.Anything).Return(errors.New("session error"))

	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

// === GetMe ===

func TestAuthHandler_GetMe_Success(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	now := time.Now()
	testUser := &models.User{
		ID:        testUserUUID1,
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		UserID:    testUserUUID1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set(UserContextKey, testUser)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).GetMe(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.MeResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, "JD", resp.Initials)
	assert.Equal(t, "user", resp.Role)
}

func TestAuthHandler_GetMe_InitialsFromUsername(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	testUser := &models.User{
		ID:       testUserUUID1,
		Username: "alice",
		UserID:   testUserUUID1,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set(UserContextKey, testUser)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).GetMe(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.MeResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "AL", resp.Initials)
}

func TestAuthHandler_GetMe_NoUser(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(mockUserRepo, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).GetMe(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "unauthorized", resp.Error.Code)
}

// === SignOut ===

func TestAuthHandler_SignOut_Success(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	session := &models.Session{
		ID:     testSessionUUID1,
		UserID: testUserUUID1,
	}

	mockSessionRepo.On("DeleteSession", mock.Anything, testSessionUUID1).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/signout", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set(SessionContextKey, session)

	err := NewAuthHandler(service.NewAuthService(nil, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignOut(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	mockSessionRepo.AssertExpectations(t)
}

func TestAuthHandler_SignOut_NoSession(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	req := httptest.NewRequest(http.MethodPost, "/api/signout", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := NewAuthHandler(service.NewAuthService(nil, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignOut(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "unauthorized", resp.Error.Code)
}

func TestAuthHandler_SignOut_DeleteError(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	session := &models.Session{
		ID:     testSessionUUID1,
		UserID: testUserUUID1,
	}

	mockSessionRepo.On("DeleteSession", mock.Anything, testSessionUUID1).Return(errors.New("db error"))

	req := httptest.NewRequest(http.MethodPost, "/api/signout", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set(SessionContextKey, session)

	err := NewAuthHandler(service.NewAuthService(nil, mockSessionRepo, nil, nil, nil, config.EmailRegistrationConfig{})).SignOut(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "internal_error", resp.Error.Code)

	mockSessionRepo.AssertExpectations(t)
}

// === buildInitials ===

func TestBuildInitials(t *testing.T) {
	tests := []struct {
		name      string
		firstName *string
		lastName  *string
		username  string
		expected  string
	}{
		{
			name:      "both names",
			firstName: stringPtr("John"),
			lastName:  stringPtr("Doe"),
			username:  "johndoe",
			expected:  "JD",
		},
		{
			name:      "first name only",
			firstName: stringPtr("Alice"),
			username:  "alice",
			expected:  "A",
		},
		{
			name:     "last name only",
			lastName: stringPtr("Smith"),
			username: "smith",
			expected: "S",
		},
		{
			name:     "no names, long username",
			username: "charlie",
			expected: "CH",
		},
		{
			name:     "no names, single char username",
			username: "x",
			expected: "X",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user := &models.User{
				Username:  tc.username,
				FirstName: tc.firstName,
				LastName:  tc.lastName,
			}
			initials, _, err := service.NewAuthService(nil, nil, nil, nil, nil, config.EmailRegistrationConfig{}).GetMe(context.Background(), user)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, initials)
		})
	}
}
