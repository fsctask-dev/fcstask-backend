package handler

import (
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
)

func intPtr(v int) *int { return &v }

func makeSessions(count int) []models.Session {
	sessions := make([]models.Session, count)
	now := time.Now()
	for i := 0; i < count; i++ {
		userID := uuid.New()
		sessions[i] = models.Session{
			ID:        uuid.New(),
			UserID:    userID,
			IP:        "127.0.0.1",
			UserAgent: "TestAgent",
			CreatedAt: now.Add(-time.Duration(i) * time.Minute),
			UpdatedAt: now.Add(-time.Duration(i) * time.Minute),
			User: models.User{
				ID:       userID,
				Email:    "user@example.com",
				Username: "testuser",
				UserID:   uuid.New(),
			},
		}
	}
	return sessions
}

func TestGetSessionsHandler_DefaultPagination(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	sessions := makeSessions(3)
	mockSessionRepo.On("CountAll", mock.Anything).Return(int64(3), nil)
	mockSessionRepo.On("GetAllWithUser", mock.Anything, 20, 0).Return(sessions, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetSessionsParams{}
	err := GetSessionsHandler(mockSessionRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.PaginatedSessionsResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, int64(3), resp.Total)
	assert.Equal(t, 20, resp.Limit)
	assert.Equal(t, 0, resp.Offset)
	assert.Len(t, resp.Items, 3)
	assert.Equal(t, "testuser", resp.Items[0].User.Username)

	mockSessionRepo.AssertExpectations(t)
}

func TestGetSessionsHandler_CustomPagination(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	sessions := makeSessions(2)
	mockSessionRepo.On("CountAll", mock.Anything).Return(int64(10), nil)
	mockSessionRepo.On("GetAllWithUser", mock.Anything, 5, 3).Return(sessions, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions?limit=5&offset=3", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetSessionsParams{Limit: intPtr(5), Offset: intPtr(3)}
	err := GetSessionsHandler(mockSessionRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.PaginatedSessionsResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, int64(10), resp.Total)
	assert.Equal(t, 5, resp.Limit)
	assert.Equal(t, 3, resp.Offset)
	assert.Len(t, resp.Items, 2)

	mockSessionRepo.AssertExpectations(t)
}

func TestGetSessionsHandler_EmptyResult(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	mockSessionRepo.On("CountAll", mock.Anything).Return(int64(0), nil)
	mockSessionRepo.On("GetAllWithUser", mock.Anything, 20, 0).Return([]models.Session{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetSessionsParams{}
	err := GetSessionsHandler(mockSessionRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.PaginatedSessionsResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, int64(0), resp.Total)
	assert.Empty(t, resp.Items)

	mockSessionRepo.AssertExpectations(t)
}

func TestGetSessionsHandler_LimitTooHigh(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions?limit=200", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetSessionsParams{Limit: intPtr(200)}
	err := GetSessionsHandler(mockSessionRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "bad_request", resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Limit")
}

func TestGetSessionsHandler_LimitZero(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions?limit=0", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetSessionsParams{Limit: intPtr(0)}
	err := GetSessionsHandler(mockSessionRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "bad_request", resp.Error.Code)
}

func TestGetSessionsHandler_NegativeOffset(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions?offset=-1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetSessionsParams{Offset: intPtr(-1)}
	err := GetSessionsHandler(mockSessionRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "bad_request", resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Offset")
}

func TestGetSessionsHandler_CountError(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	mockSessionRepo.On("CountAll", mock.Anything).Return(int64(0), errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetSessionsParams{}
	err := GetSessionsHandler(mockSessionRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "internal_error", resp.Error.Code)

	mockSessionRepo.AssertExpectations(t)
}

func TestGetSessionsHandler_GetAllError(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	mockSessionRepo.On("CountAll", mock.Anything).Return(int64(5), nil)
	mockSessionRepo.On("GetAllWithUser", mock.Anything, 20, 0).Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetSessionsParams{}
	err := GetSessionsHandler(mockSessionRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "internal_error", resp.Error.Code)

	mockSessionRepo.AssertExpectations(t)
}

// === GetUsersWithSessions ===

func TestGetUsersWithSessionsHandler_Success(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)

	now := time.Now()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sess1 := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	sess2 := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	users := []models.User{
		{
			ID:        userID,
			Email:     "john@example.com",
			Username:  "johndoe",
			UserID:    uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			CreatedAt: now,
			UpdatedAt: now,
			Sessions: []models.Session{
				{ID: sess1, UserID: userID, IP: "10.0.0.1", UserAgent: "Chrome", CreatedAt: now, UpdatedAt: now},
				{ID: sess2, UserID: userID, IP: "10.0.0.2", UserAgent: "Firefox", CreatedAt: now, UpdatedAt: now},
			},
		},
	}

	mockUserRepo.On("CountUsersWithSessions", mock.Anything).Return(int64(1), nil)
	mockUserRepo.On("GetAllWithSessions", mock.Anything, 20, 0).Return(users, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/sessions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetUsersWithSessionsParams{}
	err := GetUsersWithSessionsHandler(mockUserRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.PaginatedUsersWithSessionsResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Total)
	assert.Equal(t, 20, resp.Limit)
	assert.Equal(t, 0, resp.Offset)
	assert.Len(t, resp.Items, 1)

	item := resp.Items[0]
	assert.Equal(t, "johndoe", item.User.Username)
	assert.Len(t, item.Sessions, 2)
	assert.Equal(t, "10.0.0.1", item.Sessions[0].Ip)
	assert.Equal(t, "10.0.0.2", item.Sessions[1].Ip)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUsersWithSessionsHandler_CustomPagination(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)

	mockUserRepo.On("CountUsersWithSessions", mock.Anything).Return(int64(50), nil)
	mockUserRepo.On("GetAllWithSessions", mock.Anything, 10, 5).Return([]models.User{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/sessions?limit=10&offset=5", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetUsersWithSessionsParams{Limit: intPtr(10), Offset: intPtr(5)}
	err := GetUsersWithSessionsHandler(mockUserRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.PaginatedUsersWithSessionsResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, int64(50), resp.Total)
	assert.Equal(t, 10, resp.Limit)
	assert.Equal(t, 5, resp.Offset)
	assert.Empty(t, resp.Items)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUsersWithSessionsHandler_InvalidLimit(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/sessions?limit=200", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetUsersWithSessionsParams{Limit: intPtr(200)}
	err := GetUsersWithSessionsHandler(mockUserRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "bad_request", resp.Error.Code)
}

func TestGetUsersWithSessionsHandler_NegativeOffset(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/sessions?offset=-5", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetUsersWithSessionsParams{Offset: intPtr(-5)}
	err := GetUsersWithSessionsHandler(mockUserRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetUsersWithSessionsHandler_CountError(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)

	mockUserRepo.On("CountUsersWithSessions", mock.Anything).Return(int64(0), errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/v1/users/sessions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetUsersWithSessionsParams{}
	err := GetUsersWithSessionsHandler(mockUserRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUsersWithSessionsHandler_GetAllError(t *testing.T) {
	e := echo.New()
	mockUserRepo := new(MockUserRepository)

	mockUserRepo.On("CountUsersWithSessions", mock.Anything).Return(int64(5), nil)
	mockUserRepo.On("GetAllWithSessions", mock.Anything, 20, 0).Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/v1/users/sessions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetUsersWithSessionsParams{}
	err := GetUsersWithSessionsHandler(mockUserRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mockUserRepo.AssertExpectations(t)
}

func TestGetSessionsHandler_UserDataIncluded(t *testing.T) {
	e := echo.New()
	mockSessionRepo := new(MockSessionRepository)

	now := time.Now()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sessionID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	firstName := "John"
	lastName := "Doe"

	sessions := []models.Session{
		{
			ID:        sessionID,
			UserID:    userID,
			IP:        "192.168.1.1",
			UserAgent: "Mozilla/5.0",
			CreatedAt: now,
			UpdatedAt: now,
			User: models.User{
				ID:        userID,
				Email:     "john@example.com",
				Username:  "johndoe",
				FirstName: &firstName,
				LastName:  &lastName,
				UserID:    uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	mockSessionRepo.On("CountAll", mock.Anything).Return(int64(1), nil)
	mockSessionRepo.On("GetAllWithUser", mock.Anything, 20, 0).Return(sessions, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	params := api.GetSessionsParams{}
	err := GetSessionsHandler(mockSessionRepo, ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.PaginatedSessionsResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Len(t, resp.Items, 1)

	item := resp.Items[0]
	assert.Equal(t, "192.168.1.1", item.Ip)
	assert.Equal(t, "Mozilla/5.0", item.UserAgent)
	assert.Equal(t, "johndoe", item.User.Username)
	assert.Equal(t, "john@example.com", string(item.User.Email))
	assert.Equal(t, uuid.MustParse("44444444-4444-4444-4444-444444444444"), uuid.UUID(item.User.UserId))

	mockSessionRepo.AssertExpectations(t)
}
