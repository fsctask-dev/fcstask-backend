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

func TestGradeUpdateHandler_UpdateGrade_Success(t *testing.T) {
	e := echo.New()
	h, mockSvc := setupGradeUpdateHandler()

	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	scoreValue := 95

	expectedScore := &model.StudentTaskScore{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     scoreValue,
		IsPassed:  true,
	}

	mockSvc.On("UpdateGrade", mock.Anything, userID, mock.MatchedBy(func(input service.UpdateGradeInput) bool {
		return input.StudentID == studentID &&
			input.TaskID == taskID &&
			input.CourseID == courseID &&
			*input.Score == scoreValue
	})).Return(expectedScore, nil)

	body, _ := json.Marshal(map[string]interface{}{
		"studentId": studentID.String(),
		"taskId":    taskID.String(),
		"courseId":  courseID.String(),
		"score":     scoreValue,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: userID})
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), uuid.New().String(), taskID.String())

	err := h.UpdateGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.StudentTaskScore
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, studentID, response.StudentID)
	assert.Equal(t, scoreValue, response.Score)

	mockSvc.AssertExpectations(t)
}

func TestGradeUpdateHandler_UpdateGrade_NoUser(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(uuid.New().String(), uuid.New().String(), uuid.New().String())

	err := h.UpdateGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGradeUpdateHandler_UpdateGrade_InvalidRequestBody(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(uuid.New().String(), uuid.New().String(), uuid.New().String())

	err := h.UpdateGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGradeUpdateHandler_UpdateGrade_MissingStudentID(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"taskId":   uuid.New().String(),
		"courseId": uuid.New().String(),
		"score":    85,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(uuid.New().String(), uuid.New().String(), uuid.New().String())

	err := h.UpdateGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid student id")
}

func TestGradeUpdateHandler_UpdateGrade_MissingTaskID(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"studentId": uuid.New().String(),
		"courseId":  uuid.New().String(),
		"score":     85,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(uuid.New().String(), uuid.New().String(), uuid.New().String())

	err := h.UpdateGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid task id")
}

func TestGradeUpdateHandler_UpdateGrade_MissingCourseID(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"studentId": uuid.New().String(),
		"taskId":    uuid.New().String(),
		"score":     85,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(uuid.New().String(), uuid.New().String(), uuid.New().String())

	err := h.UpdateGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid course id")
}

func TestGradeUpdateHandler_UpdateGrade_MissingScore(t *testing.T) {
	e := echo.New()
	h, _ := setupGradeUpdateHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"studentId": uuid.New().String(),
		"taskId":    uuid.New().String(),
		"courseId":  uuid.New().String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(uuid.New().String(), uuid.New().String(), uuid.New().String())

	err := h.UpdateGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid score")
}

func TestGradeUpdateHandler_UpdateGrade_ServiceError(t *testing.T) {
	e := echo.New()
	h, mockSvc := setupGradeUpdateHandler()

	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	scoreValue := 95

	mockSvc.On("UpdateGrade", mock.Anything, userID, mock.Anything).Return(nil, assert.AnError)

	body, _ := json.Marshal(map[string]interface{}{
		"studentId": studentID.String(),
		"taskId":    taskID.String(),
		"courseId":  courseID.String(),
		"score":     scoreValue,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: userID})
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), uuid.New().String(), taskID.String())

	err := h.UpdateGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestGradeUpdateHandler_UpdateGrade_ZeroScore(t *testing.T) {
	e := echo.New()
	h, mockSvc := setupGradeUpdateHandler()

	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	scoreValue := 0

	expectedScore := &model.StudentTaskScore{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     scoreValue,
		IsPassed:  false,
	}

	mockSvc.On("UpdateGrade", mock.Anything, userID, mock.Anything).Return(expectedScore, nil)

	body, _ := json.Marshal(map[string]interface{}{
		"studentId": studentID.String(),
		"taskId":    taskID.String(),
		"courseId":  courseID.String(),
		"score":     scoreValue,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: userID})
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), uuid.New().String(), taskID.String())

	err := h.UpdateGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.StudentTaskScore
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Score)

	mockSvc.AssertExpectations(t)
}
