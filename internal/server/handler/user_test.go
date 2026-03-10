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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"fcstask/internal/api"
	models "fcstask/internal/db/model"
	"fcstask/internal/db/repo"
)

// MockUserRepository мок для репозитория пользователей
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
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

func (m *MockUserRepository) GetByUserID(ctx context.Context, userID int64) (*models.User, error) {
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

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
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

// Проверяем что мок реализует интерфейс
var _ repo.UserRepositoryInterface = (*MockUserRepository)(nil)

// TestCreateUserHandler_Success тест успешного создания пользователя
func TestCreateUserHandler_Success(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовые данные
	reqBody := api.CreateUserRequest{
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		UserId:    12345,
	}

	reqJSON, _ := json.Marshal(reqBody)

	// Ожидания мока
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Email == string(reqBody.Email) &&
			user.Username == reqBody.Username &&
			user.UserID == reqBody.UserId
	})).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*models.User)
		user.ID = 1
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	})

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем хендлер
	err := CreateUserHandler(mockRepo, ctx)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.User
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, reqBody.Email, resp.Email)
	assert.Equal(t, reqBody.Username, resp.Username)
	assert.Equal(t, reqBody.UserId, resp.UserId)
	assert.Equal(t, int64(1), resp.Id)

	mockRepo.AssertExpectations(t)
}

// TestCreateUserHandler_InvalidRequest тест с невалидным запросом
func TestCreateUserHandler_InvalidRequest(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Невалидный JSON
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем хендлер
	err := CreateUserHandler(mockRepo, ctx)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid request body", resp["error"])

	mockRepo.AssertExpectations(t)
}

// TestCreateUserHandler_MissingRequiredFields тест с отсутствующими обязательными полями
func TestCreateUserHandler_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		reqBody  api.CreateUserRequest
		expected string
	}{
		{
			name: "missing email",
			reqBody: api.CreateUserRequest{
				Username: "testuser",
				UserId:   12345,
			},
			expected: "Email, username and user_id are required",
		},
		{
			name: "missing username",
			reqBody: api.CreateUserRequest{
				Email:  "test@example.com",
				UserId: 12345,
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

			// Выполняем хендлер
			err := CreateUserHandler(mockRepo, ctx)

			// Проверяем результат
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var resp map[string]string
			json.Unmarshal(rec.Body.Bytes(), &resp)
			assert.Equal(t, tc.expected, resp["error"])

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestCreateUserHandler_DatabaseError тест с ошибкой базы данных
func TestCreateUserHandler_DatabaseError(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовые данные
	reqBody := api.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		UserId:   12345,
	}

	reqJSON, _ := json.Marshal(reqBody)

	// Ожидания мока - ошибка при создании
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))

	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем хендлер
	err := CreateUserHandler(mockRepo, ctx)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to create user", resp["error"])

	mockRepo.AssertExpectations(t)
}

// TestGetUserByIDHandler_Success тест успешного получения пользователя по ID
func TestGetUserByIDHandler_Success(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовый пользователь
	now := time.Now()
	testUser := &models.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		UserID:    12345,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Ожидания мока
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(testUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем хендлер
	err := GetUserByIDHandler(mockRepo, ctx, 1)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.User
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, testUser.Email, string(resp.Email))
	assert.Equal(t, testUser.Username, resp.Username)
	assert.Equal(t, testUser.UserID, resp.UserId)

	mockRepo.AssertExpectations(t)
}

// TestGetUserByIDHandler_NotFound тест когда пользователь не найден
func TestGetUserByIDHandler_NotFound(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Ожидания мока - пользователь не найден
	mockRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/999", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем хендлер
	err := GetUserByIDHandler(mockRepo, ctx, 999)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "User not found", resp["error"])

	mockRepo.AssertExpectations(t)
}

// TestGetUserByIDHandler_DatabaseError тест с ошибкой базы данных
func TestGetUserByIDHandler_DatabaseError(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Ожидания мока - ошибка базы данных
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/v1/users/1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем хендлер
	err := GetUserByIDHandler(mockRepo, ctx, 1)

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Failed to get user", resp["error"])

	mockRepo.AssertExpectations(t)
}

// TestGetUserByUsernameHandler_Success тест успешного получения пользователя по username
func TestGetUserByUsernameHandler_Success(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовый пользователь
	now := time.Now()
	testUser := &models.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		UserID:    12345,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Ожидания мока
	mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/username/testuser", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем хендлер
	err := GetUserByUsernameHandler(mockRepo, ctx, "testuser")

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.User
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, testUser.Username, resp.Username)
	assert.Equal(t, testUser.Email, string(resp.Email))

	mockRepo.AssertExpectations(t)
}

// TestGetUserByEmailHandler_Success тест успешного получения пользователя по email
func TestGetUserByEmailHandler_Success(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Тестовый пользователь
	now := time.Now()
	testUser := &models.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		UserID:    12345,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Ожидания мока
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/email/test@example.com", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем хендлер
	err := GetUserByEmailHandler(mockRepo, ctx, "test@example.com")

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.User
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, testUser.Email, string(resp.Email))
	assert.Equal(t, testUser.Username, resp.Username)

	mockRepo.AssertExpectations(t)
}

// TestGetUserByEmailHandler_NotFound тест когда пользователь не найден по email
func TestGetUserByEmailHandler_NotFound(t *testing.T) {
	// Настройка
	e := echo.New()
	mockRepo := new(MockUserRepository)

	// Ожидания мока - пользователь не найден
	mockRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, gorm.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodGet, "/v1/users/email/notfound@example.com", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Выполняем хендлер
	err := GetUserByEmailHandler(mockRepo, ctx, "notfound@example.com")

	// Проверяем результат
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "User not found", resp["error"])

	mockRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}
