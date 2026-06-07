// internal/handler/stats_handler_test.go
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

// MockStatsService мок для сервиса статистики
type MockStatsService struct {
	mock.Mock
}

func (m *MockStatsService) GetStats(ctx context.Context, userID uuid.UUID) (*models.PlatformStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlatformStats), args.Error(1)
}

// TestStatsHandler_GetStats_Success тест успешного получения статистики
func TestStatsHandler_GetStats_Success(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockStatsService)

	user := &models.User{ID: uuid.New()}
	expected := &models.PlatformStats{
		TotalCourses:   10,
		PublicCourses:  6,
		PrivateCourses: 4,
		TotalUsers:     42,
	}

	mockSvc.On("GetStats", mock.Anything, user.ID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(UserContextKey, user)

	err := NewStatsHandler(mockSvc).GetStats(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp models.PlatformStats
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, int64(10), resp.TotalCourses)
	assert.Equal(t, int64(42), resp.TotalUsers)

	mockSvc.AssertExpectations(t)
}

// TestStatsHandler_GetStats_Unauthorized тест без авторизации
func TestStatsHandler_GetStats_Unauthorized(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockStatsService)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// user не установлен

	err := NewStatsHandler(mockSvc).GetStats(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestStatsHandler_GetStats_Forbidden тест без прав
func TestStatsHandler_GetStats_Forbidden(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockStatsService)

	user := &models.User{ID: uuid.New()}
	mockSvc.On("GetStats", mock.Anything, user.ID).Return(nil, service.Forbidden("You don't have permission to access this resource"))

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(UserContextKey, user)

	err := NewStatsHandler(mockSvc).GetStats(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	mockSvc.AssertExpectations(t)
}

// TestStatsHandler_GetStats_InternalError тест с ошибкой сервиса
func TestStatsHandler_GetStats_InternalError(t *testing.T) {
	e := echo.New()
	mockSvc := new(MockStatsService)

	user := &models.User{ID: uuid.New()}
	mockSvc.On("GetStats", mock.Anything, user.ID).Return(nil, service.Internal("Failed to get platform stats", nil))

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(UserContextKey, user)

	err := NewStatsHandler(mockSvc).GetStats(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mockSvc.AssertExpectations(t)
}
