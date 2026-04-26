package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/oauth"
)

// --- mocks ---

type MockOAuthIdentityRepo struct {
	mock.Mock
}

func (m *MockOAuthIdentityRepo) Create(ctx context.Context, i *models.OAuthIdentity) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}
func (m *MockOAuthIdentityRepo) Update(ctx context.Context, i *models.OAuthIdentity) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}
func (m *MockOAuthIdentityRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.OAuthIdentity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OAuthIdentity), args.Error(1)
}
func (m *MockOAuthIdentityRepo) GetByProviderUID(ctx context.Context, p, uid string) (*models.OAuthIdentity, error) {
	args := m.Called(ctx, p, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OAuthIdentity), args.Error(1)
}
func (m *MockOAuthIdentityRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.OAuthIdentity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.OAuthIdentity), args.Error(1)
}
func (m *MockOAuthIdentityRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

var _ repo.IOAuthIdentityRepo = (*MockOAuthIdentityRepo)(nil)

type MockRegistrationSessionRepo struct {
	mock.Mock
}

func (m *MockRegistrationSessionRepo) Create(ctx context.Context, s *models.RegistrationSession) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}
func (m *MockRegistrationSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.RegistrationSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RegistrationSession), args.Error(1)
}
func (m *MockRegistrationSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockRegistrationSessionRepo) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

var _ repo.IRegistrationSessionRepo = (*MockRegistrationSessionRepo)(nil)

// fakeProvider always returns a fixed profile (or error).
type fakeProvider struct {
	name    string
	enabled bool
	profile *oauth.Profile
	err     error
}

func (f *fakeProvider) Name() string  { return f.name }
func (f *fakeProvider) Enabled() bool { return f.enabled }
func (f *fakeProvider) Exchange(_ context.Context, _ oauth.ExchangePayload) (*oauth.Profile, error) {
	return f.profile, f.err
}

// --- helpers ---

type oauthTestEnv struct {
	h        *OAuthHandler
	users    *MockUserRepository
	sessions *MockSessionRepository
	idents   *MockOAuthIdentityRepo
	regs     *MockRegistrationSessionRepo
}

func newOAuthEnv(t *testing.T, fp *fakeProvider) oauthTestEnv {
	t.Helper()
	users := new(MockUserRepository)
	sessions := new(MockSessionRepository)
	idents := new(MockOAuthIdentityRepo)
	regs := new(MockRegistrationSessionRepo)
	registry := oauth.NewRegistry(fp)
	h := NewOAuthHandler(users, sessions, idents, regs, registry)
	return oauthTestEnv{h, users, sessions, idents, regs}
}

func doExchange(t *testing.T, h *OAuthHandler, providerName string, body any) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/oauth/"+providerName+"/exchange", bytes.NewBuffer(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	assert.NoError(t, h.Exchange(ctx, providerName))
	return rec
}

// --- Exchange tests ---

func TestOAuthExchange_NewIdentity_CreatesRegistrationSession(t *testing.T) {
	fp := &fakeProvider{
		name:    "gitlab",
		enabled: true,
		profile: &oauth.Profile{
			ProviderUID: "42",
			Email:       "ivan@example.com",
			Username:    "ivan",
			FirstName:   "Ivan",
			LastName:    "Petrov",
			AccessToken: "AT",
		},
	}
	env := newOAuthEnv(t, fp)

	env.idents.On("GetByProviderUID", mock.Anything, "gitlab", "42").Return(nil, gorm.ErrRecordNotFound)

	regID := uuid.New()
	env.regs.On("Create", mock.Anything, mock.MatchedBy(func(s *models.RegistrationSession) bool {
		return s.Provider == "gitlab" && s.ProviderUID == "42" &&
			s.Email == "ivan@example.com" && s.AccessToken == "AT" &&
			!s.ExpiresAt.IsZero()
	})).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.RegistrationSession)
		s.ID = regID
	})

	rec := doExchange(t, env.h, "gitlab", map[string]any{"code": "x", "redirect_uri": "y"})
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "registration_required", resp["status"])
	assert.Equal(t, regID.String(), resp["registration_token"])
	suggested, ok := resp["suggested"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "ivan@example.com", suggested["email"])
	assert.Equal(t, "ivan", suggested["username"])

	// nothing persistent should have been written to oauth_identities
	env.idents.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	env.idents.AssertExpectations(t)
	env.regs.AssertExpectations(t)
}

func TestOAuthExchange_LinkedIdentity_SignsInWithoutRegistrationToken(t *testing.T) {
	fp := &fakeProvider{
		name:    "gitlab",
		enabled: true,
		profile: &oauth.Profile{ProviderUID: "42", Email: "ivan@example.com", AccessToken: "newAT"},
	}
	env := newOAuthEnv(t, fp)

	userID := uuid.New()
	identityID := uuid.New()
	sessionID := uuid.New()

	existing := &models.OAuthIdentity{
		ID:          identityID,
		Provider:    "gitlab",
		ProviderUID: "42",
		UserID:      userID,
	}
	user := &models.User{
		ID:        userID,
		Email:     "ivan@example.com",
		Username:  "ivan",
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	env.idents.On("GetByProviderUID", mock.Anything, "gitlab", "42").Return(existing, nil)
	env.idents.On("Update", mock.Anything, mock.MatchedBy(func(i *models.OAuthIdentity) bool {
		return i.ID == identityID && i.AccessToken == "newAT"
	})).Return(nil)
	env.users.On("GetByID", mock.Anything, userID).Return(user, nil)
	env.sessions.On("Create", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
		return s.UserID == userID
	})).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.Session)
		s.ID = sessionID
	})

	rec := doExchange(t, env.h, "gitlab", map[string]any{"code": "x", "redirect_uri": "y"})
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
	auth, ok := resp["auth"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, sessionID.String(), auth["session_token"])
	// crucial: no registration_token leaked on a successful sign-in
	_, hasToken := resp["registration_token"]
	assert.False(t, hasToken, "successful sign-in must not include registration_token")

	// no registration session should be created on a sign-in
	env.regs.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestOAuthExchange_UnknownProvider(t *testing.T) {
	fp := &fakeProvider{name: "gitlab", enabled: true, profile: &oauth.Profile{ProviderUID: "1"}}
	env := newOAuthEnv(t, fp)

	rec := doExchange(t, env.h, "facebook", map[string]any{"code": "x"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestOAuthExchange_DisabledProvider(t *testing.T) {
	fp := &fakeProvider{name: "gitlab", enabled: false, profile: &oauth.Profile{ProviderUID: "1"}}
	env := newOAuthEnv(t, fp)

	rec := doExchange(t, env.h, "gitlab", map[string]any{"code": "x"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestOAuthExchange_BadSignature(t *testing.T) {
	fp := &fakeProvider{name: "telegram", enabled: true, err: oauth.ErrSignatureMismatch}
	env := newOAuthEnv(t, fp)

	rec := doExchange(t, env.h, "telegram", map[string]any{"telegram_data": map[string]string{"id": "1"}})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- complete-signup tests ---

func doCompleteSignUp(t *testing.T, h *OAuthHandler, body any) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/oauth/complete-signup", bytes.NewBuffer(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	assert.NoError(t, h.CompleteSignUp(ctx))
	return rec
}

func TestOAuthCompleteSignUp_Success(t *testing.T) {
	fp := &fakeProvider{name: "gitlab", enabled: true}
	env := newOAuthEnv(t, fp)

	regID := uuid.New()
	regSession := &models.RegistrationSession{
		ID:           regID,
		Provider:     "gitlab",
		ProviderUID:  "42",
		Email:        "ivan@example.com",
		Username:     "ivan",
		AccessToken:  "AT",
		RefreshToken: "RT",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}
	createdUserID := uuid.New()
	createdSessionID := uuid.New()

	env.regs.On("GetByID", mock.Anything, regID).Return(regSession, nil)
	env.users.On("ExistsByEmail", mock.Anything, "ivan@example.com").Return(false, nil)
	env.users.On("ExistsByUsername", mock.Anything, "ivan").Return(false, nil)
	env.users.On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "ivan@example.com" && u.Username == "ivan" && u.PasswordHash == ""
	})).Return(nil).Run(func(args mock.Arguments) {
		u := args.Get(1).(*models.User)
		u.ID = createdUserID
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
	})
	env.idents.On("Create", mock.Anything, mock.MatchedBy(func(i *models.OAuthIdentity) bool {
		return i.UserID == createdUserID && i.Provider == "gitlab" &&
			i.ProviderUID == "42" && i.AccessToken == "AT" && i.RefreshToken == "RT"
	})).Return(nil)
	env.regs.On("Delete", mock.Anything, regID).Return(nil)
	env.sessions.On("Create", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
		return s.UserID == createdUserID
	})).Return(nil).Run(func(args mock.Arguments) {
		s := args.Get(1).(*models.Session)
		s.ID = createdSessionID
	})

	rec := doCompleteSignUp(t, env.h, map[string]any{
		"registration_token": regID.String(),
		"username":           "ivan",
		"email":              "ivan@example.com",
	})
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, createdSessionID.String(), resp["session_token"])

	env.users.AssertExpectations(t)
	env.sessions.AssertExpectations(t)
	env.idents.AssertExpectations(t)
	env.regs.AssertExpectations(t)
}

func TestOAuthCompleteSignUp_TokenExpired(t *testing.T) {
	fp := &fakeProvider{name: "gitlab", enabled: true}
	env := newOAuthEnv(t, fp)

	regID := uuid.New()
	env.regs.On("GetByID", mock.Anything, regID).Return(&models.RegistrationSession{
		ID: regID, ExpiresAt: time.Now().Add(-time.Minute),
	}, nil)
	env.regs.On("Delete", mock.Anything, regID).Return(nil)

	rec := doCompleteSignUp(t, env.h, map[string]any{
		"registration_token": regID.String(),
		"username":           "ivan",
		"email":              "ivan@example.com",
	})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestOAuthCompleteSignUp_TokenNotFound(t *testing.T) {
	fp := &fakeProvider{name: "gitlab", enabled: true}
	env := newOAuthEnv(t, fp)

	id := uuid.New()
	env.regs.On("GetByID", mock.Anything, id).Return(nil, gorm.ErrRecordNotFound)

	rec := doCompleteSignUp(t, env.h, map[string]any{
		"registration_token": id.String(),
		"username":           "ivan",
		"email":              "ivan@example.com",
	})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestOAuthCompleteSignUp_EmailTaken(t *testing.T) {
	fp := &fakeProvider{name: "gitlab", enabled: true}
	env := newOAuthEnv(t, fp)

	regID := uuid.New()
	env.regs.On("GetByID", mock.Anything, regID).Return(&models.RegistrationSession{
		ID: regID, ExpiresAt: time.Now().Add(10 * time.Minute),
	}, nil)
	env.users.On("ExistsByEmail", mock.Anything, "ivan@example.com").Return(true, nil)

	rec := doCompleteSignUp(t, env.h, map[string]any{
		"registration_token": regID.String(),
		"username":           "ivan",
		"email":              "ivan@example.com",
	})
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestOAuthCompleteSignUp_MissingFields(t *testing.T) {
	fp := &fakeProvider{name: "gitlab", enabled: true}
	env := newOAuthEnv(t, fp)

	rec := doCompleteSignUp(t, env.h, map[string]any{
		"registration_token": uuid.New().String(),
		"username":           "",
		"email":              "x@y.z",
	})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
