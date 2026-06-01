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
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"fcstask-backend/internal/api"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/service"
)

// MockUserRepository мок для репозитория пользователей
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUserID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByTgUID(ctx context.Context, tgUID int64) (*models.User, error) {
	args := m.Called(ctx, tgUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsUserByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CountUsers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetUsersWithSessions(ctx context.Context, limit, offset int) ([]models.User, error) {
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

// Проверяем что мок реализует интерфейс
var _ repo.IUserRepo = (*MockUserRepository)(nil)

var testUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
var testUserID2 = uuid.MustParse("99999999-9999-9999-9999-999999999999")
var testInternalUserID = uuid.MustParse("55555555-5555-5555-5555-555555555555")

// TestUserHandler_CreateUser_Success тест успешного создания пользователя
func TestUserHandler_CreateUser_Success(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовые данные
	reqBody := api.CreateUserRequest{
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		UserId:    openapi_types.UUID(testInternalUserID),
	}

	reqJSON, _ := json.Marshal(reqBody)

	// Ожидания мока
	mockRepo.On("ExistsUserByEmail", mock.Anything, string(reqBody.Email)).Return(false, nil)
	mockRepo.On("ExistsUserByUsername", mock.Anything, reqBody.Username).Return(false, nil)
	mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Email == string(reqBody.Email) &&
			user.Username == reqBody.Username &&
			user.UserID == testInternalUserID
	})).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*models.User)
		user.ID = testUserID
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	})

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем handler
	err := NewUserHandler(service.NewUserService(mockRepo)).CreateUser(ctx)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.User
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, reqBody.Email, resp.Email)
	assert.Equal(t, reqBody.Username, resp.Username)
	assert.Equal(t, reqBody.UserId, resp.UserId)
	assert.Equal(t, testUserID, resp.Id)

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_CreateUser_InvalidRequest тест с невалидным запросом
func TestUserHandler_CreateUser_InvalidRequest(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Невалидный JSON
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем handler
	err := NewUserHandler(service.NewUserService(mockRepo)).CreateUser(ctx)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "bad_request", resp.Error.Code)
	assert.Equal(t, "Invalid request body", resp.Error.Message)

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_CreateUser_MissingRequiredFields тест с отсутствующими обязательными полями
func TestUserHandler_CreateUser_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		reqBody  api.CreateUserRequest
		expected string
	}{
		{
			name: "missing email",
			reqBody: api.CreateUserRequest{
				Username: "testuser",
				UserId:   openapi_types.UUID(testInternalUserID),
			},
			expected: "Email, username and user_id are required",
		},
		{
			name: "missing username",
			reqBody: api.CreateUserRequest{
				Email:  "test@example.com",
				UserId: openapi_types.UUID(testInternalUserID),
			},
			expected: "Email, username and user_id are required",
		},
		{
			name: "missing user_id",
			reqBody: api.CreateUserRequest{
				Email:    "test@example.com",
				Username: "testuser",
			},
			expected: "Email, username and user_id are required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Настройка
			e := echo.New()
			mockRepo := new(MockUserRepository)

			reqJSON, _ := json.Marshal(tc.reqBody)

			req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			// Выполняем handler
			err := NewUserHandler(service.NewUserService(mockRepo)).CreateUser(ctx)

			// Проверяем результат
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var resp api.Error
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, "bad_request", resp.Error.Code)
			assert.Equal(t, tc.expected, resp.Error.Message)

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUserHandler_CreateUser_DatabaseError тест с ошибкой базы данных
func TestUserHandler_CreateUser_DatabaseError(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовые данные
	reqBody := api.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		UserId:   openapi_types.UUID(testInternalUserID),
	}

	reqJSON, _ := json.Marshal(reqBody)

	// Ожидания мока - ошибка при создании
	mockRepo.On("ExistsUserByEmail", mock.Anything, string(reqBody.Email)).Return(false, nil)
	mockRepo.On("ExistsUserByUsername", mock.Anything, reqBody.Username).Return(false, nil)
	mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(errors.New("database error"))

	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем handler
	err := NewUserHandler(service.NewUserService(mockRepo)).CreateUser(ctx)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "internal_error", resp.Error.Code)
	assert.Equal(t, "Failed to create user", resp.Error.Message)

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_GetUserByID_Success тест успешного получения пользователя по ID
func TestUserHandler_GetUserByID_Success(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовый пользователь
	now := time.Now()
	testUser := &models.User{
		ID:        testUserID,
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		UserID:    testInternalUserID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Ожидания мока
	mockRepo.On("GetUserByID", mock.Anything, testUserID).Return(testUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/"+testUserID.String(), nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем handler
	err := NewUserHandler(service.NewUserService(mockRepo)).GetUserByID(ctx, openapi_types.UUID(testUserID))

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.User
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, testUserID, resp.Id)
	assert.Equal(t, testUser.Email, string(resp.Email))
	assert.Equal(t, testUser.Username, resp.Username)
	assert.Equal(t, openapi_types.UUID(testUser.UserID), resp.UserId)

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_GetUserByID_NotFound тест когда пользователь не найден
func TestUserHandler_GetUserByID_NotFound(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Ожидания мока - пользователь не найден
	mockRepo.On("GetUserByID", mock.Anything, testUserID2).Return(nil, gorm.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/"+testUserID2.String(), nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем handler
	err := NewUserHandler(service.NewUserService(mockRepo)).GetUserByID(ctx, openapi_types.UUID(testUserID2))

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "not_found", resp.Error.Code)
	assert.Equal(t, "User not found", resp.Error.Message)

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_GetUserByID_DatabaseError тест с ошибкой базы данных
func TestUserHandler_GetUserByID_DatabaseError(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Ожидания мока - ошибка базы данных
	mockRepo.On("GetUserByID", mock.Anything, testUserID).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/v1/users/"+testUserID.String(), nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем handler
	err := NewUserHandler(service.NewUserService(mockRepo)).GetUserByID(ctx, openapi_types.UUID(testUserID))

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "internal_error", resp.Error.Code)
	assert.Equal(t, "Failed to get user", resp.Error.Message)

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_GetUserByUsername_Success тест успешного получения пользователя по username
func TestUserHandler_GetUserByUsername_Success(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовый пользователь
	now := time.Now()
	testUser := &models.User{
		ID:        testUserID,
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		UserID:    testInternalUserID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Ожидания мока
	mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(testUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/username/testuser", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем handler
	err := NewUserHandler(service.NewUserService(mockRepo)).GetUserByUsername(ctx, "testuser")

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.User
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, testUserID, resp.Id)
	assert.Equal(t, testUser.Username, resp.Username)
	assert.Equal(t, testUser.Email, string(resp.Email))

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_GetUserByEmail_Success тест успешного получения пользователя по email
func TestUserHandler_GetUserByEmail_Success(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовый пользователь
	now := time.Now()
	testUser := &models.User{
		ID:        testUserID,
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		UserID:    testInternalUserID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Ожидания мока
	mockRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/email/test@example.com", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем handler
	err := NewUserHandler(service.NewUserService(mockRepo)).GetUserByEmail(ctx, openapi_types.Email("test@example.com"))

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.User
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, testUserID, resp.Id)
	assert.Equal(t, testUser.Email, string(resp.Email))
	assert.Equal(t, testUser.Username, resp.Username)

	mockRepo.AssertExpectations(t)
}

// TestUserHandler_GetUserByEmail_NotFound тест когда пользователь не найден по email
func TestUserHandler_GetUserByEmail_NotFound(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Ожидания мока - пользователь не найден
	mockRepo.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return(nil, gorm.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/email/notfound@example.com", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем handler
	err := NewUserHandler(service.NewUserService(mockRepo)).GetUserByEmail(ctx, openapi_types.Email("notfound@example.com"))

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp api.Error
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "not_found", resp.Error.Code)
	assert.Equal(t, "User not found", resp.Error.Message)

	mockRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}
