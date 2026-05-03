package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
)

type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) Create(ctx context.Context, task *model.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskRepo) GetByHwID(ctx context.Context, hwID uuid.UUID) ([]*model.Task, error) {
	args := m.Called(ctx, hwID)
	return args.Get(0).([]*model.Task), args.Error(1)
}

func (m *MockTaskRepo) Update(ctx context.Context, task *model.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepo) SetScore(ctx context.Context, id uuid.UUID, score int) error {
	args := m.Called(ctx, id, score)
	return args.Error(0)
}

type MockHomeworkRepo struct {
	mock.Mock
}

func (m *MockHomeworkRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Homework, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Homework), args.Error(1)
}

func setupTest() (*echo.Echo, *MockTaskRepo, *MockHomeworkRepo) {
	e := echo.New()
	mockTaskRepo := new(MockTaskRepo)
	mockHomeworkRepo := new(MockHomeworkRepo)
	return e, mockTaskRepo, mockHomeworkRepo
}

func TestAdminCreateTaskHandler(t *testing.T) {
	tests := []struct {
		name           string
		hwID           string
		requestBody    interface{}
		setupMocks     func(*MockHomeworkRepo, *MockTaskRepo, uuid.UUID)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful task creation",
			hwID: uuid.New().String(),
			requestBody: CreateTaskRequest{
				RepoURL: stringPtr("https://github.com/test/repo"),
				TaskURL: stringPtr("https://task.com/1"),
			},
			setupMocks: func(hwRepo *MockHomeworkRepo, taskRepo *MockTaskRepo, hwID uuid.UUID) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{ID: hwID}, nil)
				taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Task")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid homework ID",
			hwID:           "invalid-uuid",
			requestBody:    CreateTaskRequest{},
			setupMocks:     func(hwRepo *MockHomeworkRepo, taskRepo *MockTaskRepo, hwID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "homework not found",
			hwID: uuid.New().String(),
			requestBody: CreateTaskRequest{
				RepoURL: stringPtr("https://github.com/test/repo"),
			},
			setupMocks: func(hwRepo *MockHomeworkRepo, taskRepo *MockTaskRepo, hwID uuid.UUID) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid request body",
			hwID:           uuid.New().String(),
			requestBody:    "invalid json",
			setupMocks:     func(hwRepo *MockHomeworkRepo, taskRepo *MockTaskRepo, hwID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "create task fails",
			hwID: uuid.New().String(),
			requestBody: CreateTaskRequest{
				RepoURL: stringPtr("https://github.com/test/repo"),
			},
			setupMocks: func(hwRepo *MockHomeworkRepo, taskRepo *MockTaskRepo, hwID uuid.UUID) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{ID: hwID}, nil)
				taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Task")).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockTaskRepo, mockHomeworkRepo := setupTest()
			handler := NewAdminTaskHandler(mockTaskRepo, mockHomeworkRepo)

			hwID, _ := uuid.Parse(tt.hwID)
			tt.setupMocks(mockHomeworkRepo, mockTaskRepo, hwID)

			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/admin/courses/courseId/homework/"+tt.hwID+"/tasks", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("hwId")
			c.SetParamValues(tt.hwID)

			err := handler.AdminCreateTaskHandler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			mockHomeworkRepo.AssertExpectations(t)
			mockTaskRepo.AssertExpectations(t)
		})
	}
}

func TestAdminListTasksHandler(t *testing.T) {
	tests := []struct {
		name           string
		hwID           string
		setupMocks     func(*MockHomeworkRepo, *MockTaskRepo, uuid.UUID)
		expectedStatus int
	}{
		{
			name: "successful list",
			hwID: uuid.New().String(),
			setupMocks: func(hwRepo *MockHomeworkRepo, taskRepo *MockTaskRepo, hwID uuid.UUID) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{ID: hwID}, nil)
				taskRepo.On("GetByHwID", mock.Anything, hwID).Return([]*model.Task{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid homework ID",
			hwID:           "invalid-uuid",
			setupMocks:     func(hwRepo *MockHomeworkRepo, taskRepo *MockTaskRepo, hwID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "homework not found",
			hwID: uuid.New().String(),
			setupMocks: func(hwRepo *MockHomeworkRepo, taskRepo *MockTaskRepo, hwID uuid.UUID) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "get tasks fails",
			hwID: uuid.New().String(),
			setupMocks: func(hwRepo *MockHomeworkRepo, taskRepo *MockTaskRepo, hwID uuid.UUID) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{ID: hwID}, nil)
				taskRepo.On("GetByHwID", mock.Anything, hwID).Return([]*model.Task{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockTaskRepo, mockHomeworkRepo := setupTest()
			handler := NewAdminTaskHandler(mockTaskRepo, mockHomeworkRepo)

			hwID, _ := uuid.Parse(tt.hwID)
			tt.setupMocks(mockHomeworkRepo, mockTaskRepo, hwID)

			req := httptest.NewRequest(http.MethodGet, "/admin/courses/courseId/homework/"+tt.hwID+"/tasks", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("hwId")
			c.SetParamValues(tt.hwID)

			err := handler.AdminListTasksHandler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAdminGetTaskHandler(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name           string
		taskID         string
		setupMocks     func(*MockTaskRepo)
		expectedStatus int
	}{
		{
			name:   "successful get",
			taskID: taskID.String(),
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{ID: taskID}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid task ID",
			taskID:         "invalid-uuid",
			setupMocks:     func(taskRepo *MockTaskRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "task not found",
			taskID: taskID.String(),
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockTaskRepo, mockHomeworkRepo := setupTest()
			handler := NewAdminTaskHandler(mockTaskRepo, mockHomeworkRepo)

			tt.setupMocks(mockTaskRepo)

			req := httptest.NewRequest(http.MethodGet, "/admin/courses/courseId/homework/hwId/tasks/"+tt.taskID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("taskId")
			c.SetParamValues(tt.taskID)

			err := handler.AdminGetTaskHandler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAdminUpdateTaskHandler(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name           string
		taskID         string
		requestBody    interface{}
		setupMocks     func(*MockTaskRepo)
		expectedStatus int
	}{
		{
			name:   "successful update",
			taskID: taskID.String(),
			requestBody: UpdateTaskRequest{
				RepoURL: stringPtr("https://github.com/updated/repo"),
				TaskURL: stringPtr("https://updated.com"),
			},
			setupMocks: func(taskRepo *MockTaskRepo) {
				task := &model.Task{ID: taskID, RepoURL: stringPtr("old"), TaskURL: stringPtr("old")}
				taskRepo.On("GetByID", mock.Anything, taskID).Return(task, nil)
				taskRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Task")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid task ID",
			taskID:         "invalid-uuid",
			requestBody:    UpdateTaskRequest{},
			setupMocks:     func(taskRepo *MockTaskRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "task not found",
			taskID: taskID.String(),
			requestBody: UpdateTaskRequest{
				RepoURL: stringPtr("test"),
			},
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid request body",
			taskID:         taskID.String(),
			requestBody:    "invalid json",
			setupMocks:     func(taskRepo *MockTaskRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockTaskRepo, mockHomeworkRepo := setupTest()
			handler := NewAdminTaskHandler(mockTaskRepo, mockHomeworkRepo)

			tt.setupMocks(mockTaskRepo)

			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/admin/courses/courseId/homework/hwId/tasks/"+tt.taskID, bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("taskId")
			c.SetParamValues(tt.taskID)

			err := handler.AdminUpdateTaskHandler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAdminDeleteTaskHandler(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name           string
		taskID         string
		setupMocks     func(*MockTaskRepo)
		expectedStatus int
	}{
		{
			name:   "successful delete",
			taskID: taskID.String(),
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{ID: taskID}, nil)
				taskRepo.On("Delete", mock.Anything, taskID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid task ID",
			taskID:         "invalid-uuid",
			setupMocks:     func(taskRepo *MockTaskRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "task not found",
			taskID: taskID.String(),
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "delete fails",
			taskID: taskID.String(),
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{ID: taskID}, nil)
				taskRepo.On("Delete", mock.Anything, taskID).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockTaskRepo, mockHomeworkRepo := setupTest()
			handler := NewAdminTaskHandler(mockTaskRepo, mockHomeworkRepo)

			tt.setupMocks(mockTaskRepo)

			req := httptest.NewRequest(http.MethodDelete, "/admin/courses/courseId/homework/hwId/tasks/"+tt.taskID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("taskId")
			c.SetParamValues(tt.taskID)

			err := handler.AdminDeleteTaskHandler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAdminSetTaskScoreHandler(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name           string
		taskID         string
		requestBody    interface{}
		setupMocks     func(*MockTaskRepo)
		expectedStatus int
	}{
		{
			name:   "successful set score",
			taskID: taskID.String(),
			requestBody: map[string]int{
				"score": 95,
			},
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{ID: taskID}, nil)
				taskRepo.On("SetScore", mock.Anything, taskID, 95).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid task ID",
			taskID:         "invalid-uuid",
			requestBody:    map[string]int{"score": 100},
			setupMocks:     func(taskRepo *MockTaskRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "task not found",
			taskID: taskID.String(),
			requestBody: map[string]int{
				"score": 100,
			},
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "negative score",
			taskID: taskID.String(),
			requestBody: map[string]int{
				"score": -5,
			},
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{ID: taskID}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request body",
			taskID:         taskID.String(),
			requestBody:    "invalid json",
			setupMocks:     func(taskRepo *MockTaskRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "set score fails",
			taskID: taskID.String(),
			requestBody: map[string]int{
				"score": 100,
			},
			setupMocks: func(taskRepo *MockTaskRepo) {
				taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{ID: taskID}, nil)
				taskRepo.On("SetScore", mock.Anything, taskID, 100).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockTaskRepo, mockHomeworkRepo := setupTest()
			handler := NewAdminTaskHandler(mockTaskRepo, mockHomeworkRepo)

			tt.setupMocks(mockTaskRepo)

			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/admin/courses/courseId/homework/hwId/tasks/"+tt.taskID+"/score", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("taskId")
			c.SetParamValues(tt.taskID)

			err := handler.AdminSetTaskScoreHandler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
