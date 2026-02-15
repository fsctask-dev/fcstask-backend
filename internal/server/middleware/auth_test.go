package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask/internal/api"
	models "fcstask/internal/db/model"
	"fcstask/internal/db/repo"
	"fcstask/internal/server/handler"
)

// --- Mocks ---

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByTgUID(ctx context.Context, tgUID int64) (*models.User, error) {
	args := m.Called(ctx, tgUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetAllWithSessions(ctx context.Context, limit, offset int) ([]models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) CountUsersWithSessions(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

var _ repo.UserRepositoryInterface = (*MockUserRepository)(nil)

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetAllWithUser(ctx context.Context, limit, offset int) ([]models.Session, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockSessionRepository) CountAll(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSessionRepository) TouchAccessedAt(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) CleanOutdated(ctx context.Context, ttl time.Duration) (int64, error) {
	args := m.Called(ctx, ttl)
	return args.Get(0).(int64), args.Error(1)
}

var _ repo.SessionRepositoryInterface = (*MockSessionRepository)(nil)

// --- Tests ---

var (
	testSessionID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testUserID    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
)

func TestAuth_UnprotectedPath(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	mw := Auth(mockUserRepo, mockSessionRepo, []string{"/api/me"})

	called := false
	h := mw(func(ctx echo.Context) error {
		called = true
		return ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/api/signin", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/api/signin")

	err := h(ctx)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuth_MissingAuthHeader(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	mw := Auth(mockUserRepo, mockSessionRepo, []string{"/api/me"})

	h := mw(func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/api/me")

	err := h(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "unauthorized", resp.Error.Code)
	assert.Equal(t, "Missing or invalid Authorization header", resp.Error.Message)
}

func TestAuth_InvalidBearerPrefix(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	mw := Auth(mockUserRepo, mockSessionRepo, []string{"/api/me"})

	h := mw(func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/api/me")

	err := h(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Missing or invalid Authorization header", resp.Error.Message)
}

func TestAuth_InvalidTokenFormat(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	mw := Auth(mockUserRepo, mockSessionRepo, []string{"/api/me"})

	h := mw(func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer not-a-uuid")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/api/me")

	err := h(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid session token", resp.Error.Message)
}

func TestAuth_SessionNotFound(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	mockSessionRepo.On("GetByID", mock.Anything, testSessionID).Return(nil, errors.New("not found"))

	mw := Auth(mockUserRepo, mockSessionRepo, []string{"/api/me"})

	h := mw(func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+testSessionID.String())
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/api/me")

	err := h(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Session not found", resp.Error.Message)

	mockSessionRepo.AssertExpectations(t)
}

func TestAuth_UserNotFound(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	session := &models.Session{
		ID:     testSessionID,
		UserID: testUserID,
	}
	mockSessionRepo.On("GetByID", mock.Anything, testSessionID).Return(session, nil)
	mockUserRepo.On("GetByID", mock.Anything, testUserID).Return(nil, errors.New("not found"))

	mw := Auth(mockUserRepo, mockSessionRepo, []string{"/api/me"})

	h := mw(func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+testSessionID.String())
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/api/me")

	err := h(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "User not found", resp.Error.Message)

	mockSessionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAuth_Success(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	now := time.Now()
	session := &models.Session{
		ID:     testSessionID,
		UserID: testUserID,
	}
	user := &models.User{
		ID:        testUserID,
		Email:     "test@example.com",
		Username:  "testuser",
		UserID:    uuid.MustParse("55555555-5555-5555-5555-555555555555"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockSessionRepo.On("GetByID", mock.Anything, testSessionID).Return(session, nil)
	mockSessionRepo.On("TouchAccessedAt", mock.Anything, testSessionID).Return(nil)
	mockUserRepo.On("GetByID", mock.Anything, testUserID).Return(user, nil)

	mw := Auth(mockUserRepo, mockSessionRepo, []string{"/api/me"})

	var ctxUser *models.User
	h := mw(func(ctx echo.Context) error {
		u, ok := ctx.Get(handler.UserContextKey).(*models.User)
		if ok {
			ctxUser = u
		}
		return ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+testSessionID.String())
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/api/me")

	err := h(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, ctxUser)
	assert.Equal(t, "testuser", ctxUser.Username)
	assert.Equal(t, testUserID, ctxUser.ID)

	mockSessionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
