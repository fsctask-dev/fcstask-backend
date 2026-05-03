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

	"fcstask-backend/internal/db/model"
)

type MockHomeworkRepo struct {
	mock.Mock
}

func (m *MockHomeworkRepo) Create(ctx context.Context, hw *model.Homework) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *MockHomeworkRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Homework, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Homework), args.Error(1)
}

func (m *MockHomeworkRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]*model.Homework, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Homework), args.Error(1)
}

func (m *MockHomeworkRepo) Update(ctx context.Context, hw *model.Homework) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *MockHomeworkRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockDeadlineRepo struct {
	mock.Mock
}

func (m *MockDeadlineRepo) Create(ctx context.Context, deadline *model.Deadline) error {
	args := m.Called(ctx, deadline)
	return args.Error(0)
}

func (m *MockDeadlineRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Deadline, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deadline), args.Error(1)
}

func (m *MockDeadlineRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]*model.Deadline, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Deadline), args.Error(1)
}

func (m *MockDeadlineRepo) Update(ctx context.Context, deadline *model.Deadline) error {
	args := m.Called(ctx, deadline)
	return args.Error(0)
}

func (m *MockDeadlineRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestAdminHomeworkHandler(t *testing.T) (*AdminHomeworkHandler, *MockHomeworkRepo, *MockDeadlineRepo, *echo.Echo) {
	e := echo.New()
	mockHomeworkRepo := new(MockHomeworkRepo)
	mockDeadlineRepo := new(MockDeadlineRepo)
	handler := NewAdminHomeworkHandler(mockHomeworkRepo, mockDeadlineRepo)
	return handler, mockHomeworkRepo, mockDeadlineRepo, e
}

func createContext(e *echo.Echo, method, path string, body interface{}, params map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	for key, value := range params {
		c.SetParamNames(key)
		c.SetParamValues(value)
	}

	return c, rec
}

func TestAdminCreateHomeworkHandler(t *testing.T) {
	tests := []struct {
		name           string
		courseID       string
		requestBody    interface{}
		setupMock      func(*MockHomeworkRepo)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "Success - with both dates",
			courseID: uuid.New().String(),
			requestBody: CreateHomeworkRequest{
				StartDate: stringPtr("2024-01-01"),
				EndDate:   stringPtr("2024-12-31"),
			},
			setupMock: func(m *MockHomeworkRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*model.Homework")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "Success - without dates",
			courseID: uuid.New().String(),
			requestBody: CreateHomeworkRequest{
				StartDate: nil,
				EndDate:   nil,
			},
			setupMock: func(m *MockHomeworkRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*model.Homework")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid course ID",
			courseID:       "invalid-uuid",
			requestBody:    CreateHomeworkRequest{},
			setupMock:      func(m *MockHomeworkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid course ID",
		},
		{
			name:     "Invalid start_date format",
			courseID: uuid.New().String(),
			requestBody: CreateHomeworkRequest{
				StartDate: stringPtr("01-01-2024"),
			},
			setupMock:      func(m *MockHomeworkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "start_date must be in format YYYY-MM-DD",
		},
		{
			name:     "End date before start date",
			courseID: uuid.New().String(),
			requestBody: CreateHomeworkRequest{
				StartDate: stringPtr("2024-12-31"),
				EndDate:   stringPtr("2024-01-01"),
			},
			setupMock:      func(m *MockHomeworkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "end_date must be after start_date",
		},
		{
			name:     "Repository create error",
			courseID: uuid.New().String(),
			requestBody: CreateHomeworkRequest{
				StartDate: stringPtr("2024-01-01"),
			},
			setupMock: func(m *MockHomeworkRepo) {
				m.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to create homework",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockHomeworkRepo, _, e := setupTestAdminHomeworkHandler(t)
			tt.setupMock(mockHomeworkRepo)

			c, rec := createContext(e, http.MethodPost, "/admin/courses/"+tt.courseID+"/homework", tt.requestBody, map[string]string{"courseId": tt.courseID})

			err := handler.AdminCreateHomeworkHandler(c)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["message"], tt.expectedError)
			}

			mockHomeworkRepo.AssertExpectations(t)
		})
	}
}

func TestAdminGetHomeworkHandler(t *testing.T) {
	hwID := uuid.New()
	courseID := uuid.New()

	tests := []struct {
		name           string
		hwID           string
		setupMock      func(*MockHomeworkRepo)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success",
			hwID: hwID.String(),
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid homework ID",
			hwID:           "invalid-uuid",
			setupMock:      func(m *MockHomeworkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid homework ID",
		},
		{
			name: "Homework not found",
			hwID: hwID.String(),
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Homework not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockHomeworkRepo, _, e := setupTestAdminHomeworkHandler(t)
			tt.setupMock(mockHomeworkRepo)

			c, rec := createContext(e, http.MethodGet, "/admin/courses/"+courseID.String()+"/homework/"+tt.hwID, nil, map[string]string{"hwId": tt.hwID})

			err := handler.AdminGetHomeworkHandler(c)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["message"], tt.expectedError)
			}

			mockHomeworkRepo.AssertExpectations(t)
		})
	}
}

func TestAdminListHomeworkHandler(t *testing.T) {
	courseID := uuid.New()
	hw1ID := uuid.New()
	hw2ID := uuid.New()

	tests := []struct {
		name           string
		courseID       string
		setupMock      func(*MockHomeworkRepo)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:     "Success - multiple homework",
			courseID: courseID.String(),
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByCourseID", mock.Anything, courseID).Return([]*model.Homework{
					{ID: hw1ID, CourseID: courseID},
					{ID: hw2ID, CourseID: courseID},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:     "Success - no homework",
			courseID: courseID.String(),
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByCourseID", mock.Anything, courseID).Return([]*model.Homework{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "Invalid course ID",
			courseID:       "invalid-uuid",
			setupMock:      func(m *MockHomeworkRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Repository error",
			courseID: courseID.String(),
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByCourseID", mock.Anything, courseID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockHomeworkRepo, _, e := setupTestAdminHomeworkHandler(t)
			tt.setupMock(mockHomeworkRepo)
			c, rec := createContext(e, http.MethodGet, "/admin/courses/"+tt.courseID+"/homework", nil, map[string]string{"courseId": tt.courseID})
			err := handler.AdminListHomeworkHandler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusOK {
				var response []*model.Homework
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, tt.expectedCount)
			}

			mockHomeworkRepo.AssertExpectations(t)
		})
	}
}

func TestAdminUpdateHomeworkHandler(t *testing.T) {
	hwID := uuid.New()
	courseID := uuid.New()

	tests := []struct {
		name           string
		hwID           string
		requestBody    interface{}
		setupMock      func(*MockHomeworkRepo)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success - update both dates",
			hwID: hwID.String(),
			requestBody: UpdateHomeworkRequest{
				StartDate: stringPtr("2024-06-01"),
				EndDate:   stringPtr("2024-12-31"),
			},
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*model.Homework")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Success - update only start date",
			hwID:        hwID.String(),
			requestBody: UpdateHomeworkRequest{StartDate: stringPtr("2024-06-01")},
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*model.Homework")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid homework ID",
			hwID:           "invalid-uuid",
			requestBody:    UpdateHomeworkRequest{},
			setupMock:      func(m *MockHomeworkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid homework ID",
		},
		{
			name: "Homework not found",
			hwID: hwID.String(),
			requestBody: UpdateHomeworkRequest{
				StartDate: stringPtr("2024-06-01"),
			},
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Homework not found",
		},
		{
			name: "Invalid date format",
			hwID: hwID.String(),
			requestBody: UpdateHomeworkRequest{
				StartDate: stringPtr("06-01-2024"),
			},
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "start_date must be in format YYYY-MM-DD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockHomeworkRepo, _, e := setupTestAdminHomeworkHandler(t)
			tt.setupMock(mockHomeworkRepo)

			c, rec := createContext(e, http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+tt.hwID, tt.requestBody, map[string]string{"hwId": tt.hwID})

			err := handler.AdminUpdateHomeworkHandler(c)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["message"], tt.expectedError)
			}

			mockHomeworkRepo.AssertExpectations(t)
		})
	}
}

func TestAdminDeleteHomeworkHandler(t *testing.T) {
	hwID := uuid.New()
	courseID := uuid.New()

	tests := []struct {
		name           string
		hwID           string
		setupMock      func(*MockHomeworkRepo)
		expectedStatus int
	}{
		{
			name: "Success",
			hwID: hwID.String(),
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
				m.On("Delete", mock.Anything, hwID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Invalid homework ID",
			hwID: "invalid-uuid",
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Homework not found",
			hwID: hwID.String(),
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Delete error",
			hwID: hwID.String(),
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
				m.On("Delete", mock.Anything, hwID).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockHomeworkRepo, _, e := setupTestAdminHomeworkHandler(t)
			tt.setupMock(mockHomeworkRepo)

			c, rec := createContext(e, http.MethodDelete, "/admin/courses/"+courseID.String()+"/homework/"+tt.hwID, nil, map[string]string{"hwId": tt.hwID})

			err := handler.AdminDeleteHomeworkHandler(c)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockHomeworkRepo.AssertExpectations(t)
		})
	}
}

func TestAdminPublishHomeworkHandler(t *testing.T) {
	hwID := uuid.New()
	courseID := uuid.New()
	isPublic := true

	tests := []struct {
		name           string
		hwID           string
		requestBody    interface{}
		setupMock      func(*MockHomeworkRepo)
		expectedStatus int
	}{
		{
			name: "Success - publish",
			hwID: hwID.String(),
			requestBody: PublishHomeworkRequest{
				IsPublic: isPublic,
			},
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*model.Homework")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid homework ID",
			hwID: "invalid-uuid",
			requestBody: PublishHomeworkRequest{
				IsPublic: isPublic,
			},
			setupMock:      func(m *MockHomeworkRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Homework not found",
			hwID: hwID.String(),
			requestBody: PublishHomeworkRequest{
				IsPublic: isPublic,
			},
			setupMock: func(m *MockHomeworkRepo) {
				m.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockHomeworkRepo, _, e := setupTestAdminHomeworkHandler(t)
			tt.setupMock(mockHomeworkRepo)

			c, rec := createContext(e, http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+tt.hwID+"/publish", tt.requestBody, map[string]string{"hwId": tt.hwID})

			err := handler.AdminPublishHomeworkHandler(c)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockHomeworkRepo.AssertExpectations(t)
		})
	}
}

func TestAdminSetHomeworkDeadlineHandler(t *testing.T) {
	courseID := uuid.New()
	hwID := uuid.New()
	userID := uuid.New()

	dueDate := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name           string
		courseID       string
		hwID           string
		requestBody    interface{}
		setupMock      func(*MockHomeworkRepo, *MockDeadlineRepo)
		expectedStatus int
	}{
		{
			name:     "Success",
			courseID: courseID.String(),
			hwID:     hwID.String(),
			requestBody: SetHomeworkDeadlineRequest{
				Title:       "Final Deadline",
				Description: stringPtr("Submit final project"),
				DueDate:     dueDate,
			},
			setupMock: func(hwRepo *MockHomeworkRepo, deadlineRepo *MockDeadlineRepo) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
				deadlineRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Deadline")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "Invalid course ID",
			courseID: "invalid-uuid",
			hwID:     hwID.String(),
			requestBody: SetHomeworkDeadlineRequest{
				Title:   "Deadline",
				DueDate: dueDate,
			},
			setupMock:      func(hwRepo *MockHomeworkRepo, deadlineRepo *MockDeadlineRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Invalid homework ID",
			courseID: courseID.String(),
			hwID:     "invalid-uuid",
			requestBody: SetHomeworkDeadlineRequest{
				Title:   "Deadline",
				DueDate: dueDate,
			},
			setupMock:      func(hwRepo *MockHomeworkRepo, deadlineRepo *MockDeadlineRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Homework not found",
			courseID: courseID.String(),
			hwID:     hwID.String(),
			requestBody: SetHomeworkDeadlineRequest{
				Title:   "Deadline",
				DueDate: dueDate,
			},
			setupMock: func(hwRepo *MockHomeworkRepo, deadlineRepo *MockDeadlineRepo) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "Missing title",
			courseID: courseID.String(),
			hwID:     hwID.String(),
			requestBody: SetHomeworkDeadlineRequest{
				Title:   "",
				DueDate: dueDate,
			},
			setupMock: func(hwRepo *MockHomeworkRepo, deadlineRepo *MockDeadlineRepo) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Invalid due date format",
			courseID: courseID.String(),
			hwID:     hwID.String(),
			requestBody: SetHomeworkDeadlineRequest{
				Title:   "Deadline",
				DueDate: "2024-01-01",
			},
			setupMock: func(hwRepo *MockHomeworkRepo, deadlineRepo *MockDeadlineRepo) {
				hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{
					ID:       hwID,
					CourseID: courseID,
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockHomeworkRepo, mockDeadlineRepo, e := setupTestAdminHomeworkHandler(t)
			tt.setupMock(mockHomeworkRepo, mockDeadlineRepo)

			c, rec := createContext(e, http.MethodPut, "/admin/courses/"+tt.courseID+"/homework/"+tt.hwID+"/deadline", tt.requestBody, map[string]string{
				"courseId": tt.courseID,
				"hwId":     tt.hwID,
			})

			if tt.expectedStatus == http.StatusCreated {
				user := &model.User{ID: userID}
				c.Set(UserContextKey, user)
			}

			err := handler.AdminSetHomeworkDeadlineHandler(c)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockHomeworkRepo.AssertExpectations(t)
			mockDeadlineRepo.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
