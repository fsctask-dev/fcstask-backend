package handler_test

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
	"fcstask-backend/internal/handler"
	"fcstask-backend/internal/service"
)

type mockGradeUpdateService struct {
	mock.Mock
}

func (m *mockGradeUpdateService) UpdateGrade(ctx context.Context, userID uuid.UUID, input service.UpdateGradeInput) (*model.StudentTaskScore, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StudentTaskScore), args.Error(1)
}

func setupGradeUpdateHandler() (*handler.GradeUpdateHandler, *mockGradeUpdateService) {
	svc := new(mockGradeUpdateService)
	return handler.NewGradeUpdateHandler(svc), svc
}

func createTestContext(e *echo.Echo, method, path string, body []byte, params map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	paramNames := make([]string, 0, len(params))
	paramValues := make([]string, 0, len(params))
	for name, value := range params {
		paramNames = append(paramNames, name)
		paramValues = append(paramValues, value)
	}
	c.SetParamNames(paramNames...)
	c.SetParamValues(paramValues...)

	return c, rec
}

func TestGradeUpdateHandler_UpdateGrade_Success(t *testing.T) {
	e := echo.New()
	h, mockSvc := setupGradeUpdateHandler()

	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 95

	expectedScore := &model.StudentTaskScore{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     scoreValue,
	}

	mockSvc.On("UpdateGrade", mock.Anything, userID, mock.MatchedBy(func(input service.UpdateGradeInput) bool {
		return input.StudentID == studentID &&
			input.TaskID == taskID &&
			input.CourseID == courseID &&
			input.Score != nil &&
			*input.Score == scoreValue
	})).Return(expectedScore, nil)

	requestBody := map[string]interface{}{
		"studentId": studentID.String(),
		"score":     scoreValue,
	}
	body, _ := json.Marshal(requestBody)

	c, rec := createTestContext(e, http.MethodPost, "/", body, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: userID})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.StudentTaskScore
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, studentID, response.StudentID)
	assert.Equal(t, taskID, response.TaskID)
	assert.Equal(t, courseID, response.CourseID)
	assert.Equal(t, scoreValue, response.Score)

	mockSvc.AssertExpectations(t)
}

func TestGradeUpdateHandler_UpdateGrade_NoUserInContext(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()

	c, rec := createTestContext(e, http.MethodPost, "/", nil, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "User not found")
}

func TestGradeUpdateHandler_UpdateGrade_InvalidCourseID(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	c, rec := createTestContext(e, http.MethodPost, "/", nil, map[string]string{
		"course_id": "invalid-uuid",
		"hw_id":     uuid.New().String(),
		"task_id":   uuid.New().String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid course_id")
}

func TestGradeUpdateHandler_UpdateGrade_MissingCourseID(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	c, rec := createTestContext(e, http.MethodPost, "/", nil, map[string]string{
		"hw_id":   uuid.New().String(),
		"task_id": uuid.New().String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGradeUpdateHandler_UpdateGrade_InvalidTaskID(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	courseID := uuid.New()
	hwID := uuid.New()

	c, rec := createTestContext(e, http.MethodPost, "/", nil, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   "invalid-uuid",
	})
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid task_id")
}

func TestGradeUpdateHandler_UpdateGrade_MissingTaskID(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	courseID := uuid.New()
	hwID := uuid.New()

	c, rec := createTestContext(e, http.MethodPost, "/", nil, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGradeUpdateHandler_UpdateGrade_InvalidJSON(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()

	c, rec := createTestContext(e, http.MethodPost, "/", []byte("invalid json"), map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid request body")
}

func TestGradeUpdateHandler_UpdateGrade_MissingStudentID(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	userID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	scoreValue := 85

	requestBody := map[string]interface{}{
		"score": scoreValue,
	}
	body, _ := json.Marshal(requestBody)

	c, rec := createTestContext(e, http.MethodPost, "/", body, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: userID})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid student id")
}

func TestGradeUpdateHandler_UpdateGrade_StudentIDEmpty(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	userID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	scoreValue := 85

	requestBody := map[string]interface{}{
		"studentId": "",
		"score":     scoreValue,
	}
	body, _ := json.Marshal(requestBody)

	c, rec := createTestContext(e, http.MethodPost, "/", body, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: userID})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid request body")
}

func TestGradeUpdateHandler_UpdateGrade_StudentIDNullUUID(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	userID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	scoreValue := 85

	requestBody := map[string]interface{}{
		"studentId": "00000000-0000-0000-0000-000000000000",
		"score":     scoreValue,
	}
	body, _ := json.Marshal(requestBody)

	c, rec := createTestContext(e, http.MethodPost, "/", body, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: userID})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid student id")
}

func TestGradeUpdateHandler_UpdateGrade_MissingScore(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	userID := uuid.New()
	studentID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()

	requestBody := map[string]interface{}{
		"studentId": studentID.String(),
	}
	body, _ := json.Marshal(requestBody)

	c, rec := createTestContext(e, http.MethodPost, "/", body, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: userID})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid score")
}

func TestGradeUpdateHandler_UpdateGrade_ScoreNull(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	userID := uuid.New()
	studentID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()

	requestBody := map[string]interface{}{
		"studentId": studentID.String(),
		"score":     nil,
	}
	body, _ := json.Marshal(requestBody)

	c, rec := createTestContext(e, http.MethodPost, "/", body, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: userID})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid score")
}

func TestGradeUpdateHandler_UpdateGrade_ServiceReturnsNotFound(t *testing.T) {
	e := echo.New()
	h, mockSvc := setupGradeUpdateHandler()

	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 95

	mockSvc.On("UpdateGrade", mock.Anything, userID, mock.Anything).Return(nil, service.NotFound("task not found"))

	requestBody := map[string]interface{}{
		"studentId": studentID.String(),
		"score":     scoreValue,
	}
	body, _ := json.Marshal(requestBody)

	c, rec := createTestContext(e, http.MethodPost, "/", body, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: userID})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestGradeUpdateHandler_UpdateGrade_ServiceReturnsInternalError(t *testing.T) {
	e := echo.New()
	h, mockSvc := setupGradeUpdateHandler()

	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 95

	mockSvc.On("UpdateGrade", mock.Anything, userID, mock.Anything).Return(nil, service.Internal("Failed to update grade", nil))

	requestBody := map[string]interface{}{
		"studentId": studentID.String(),
		"score":     scoreValue,
	}
	body, _ := json.Marshal(requestBody)

	c, rec := createTestContext(e, http.MethodPost, "/", body, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: userID})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestGradeUpdateHandler_UpdateGrade_ServiceCalledWithCorrectInput(t *testing.T) {
	e := echo.New()
	h, mockSvc := setupGradeUpdateHandler()

	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 75

	var capturedInput service.UpdateGradeInput

	mockSvc.On("UpdateGrade", mock.Anything, userID, mock.MatchedBy(func(input service.UpdateGradeInput) bool {
		capturedInput = input
		return input.StudentID == studentID &&
			input.TaskID == taskID &&
			input.CourseID == courseID &&
			input.Score != nil &&
			*input.Score == scoreValue
	})).Return(&model.StudentTaskScore{}, nil)

	requestBody := map[string]interface{}{
		"studentId": studentID.String(),
		"score":     scoreValue,
	}
	body, _ := json.Marshal(requestBody)

	c, rec := createTestContext(e, http.MethodPost, "/", body, map[string]string{
		"course_id": courseID.String(),
		"hw_id":     hwID.String(),
		"task_id":   taskID.String(),
	})
	c.Set(handler.UserContextKey, &model.User{ID: userID})

	err := h.UpdateGrade(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, studentID, capturedInput.StudentID)
	assert.Equal(t, taskID, capturedInput.TaskID)
	assert.Equal(t, courseID, capturedInput.CourseID)
	assert.Equal(t, scoreValue, *capturedInput.Score)

	mockSvc.AssertExpectations(t)
}
