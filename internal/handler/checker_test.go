package handler_test

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
	"fcstask-backend/internal/handler"
	"fcstask-backend/internal/service"
)

type MockCheckerService struct {
	mock.Mock
}

func (m *MockCheckerService) SubmitGrade(ctx context.Context, input service.SubmitGradeInput) (*model.StudentTaskScore, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StudentTaskScore), args.Error(1)
}

func TestCheckerHandler_SubmitGrade_Success(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	submittedAt := time.Now().UTC()

	reqBody := map[string]interface{}{
		"student_id":   studentID,
		"task_id":      taskID,
		"course_id":    courseID,
		"raw_score":    85,
		"is_passed":    true,
		"submitted_at": submittedAt,
	}

	expected := &model.StudentTaskScore{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     85,
		IsPassed:  true,
	}

	e := echo.New()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc.On("SubmitGrade", mock.Anything, mock.MatchedBy(func(input service.SubmitGradeInput) bool {
		return input.StudentID == studentID &&
			input.TaskID == taskID &&
			input.CourseID == courseID &&
			input.RawScore == 85 &&
			input.IsPassed == true &&
			input.SubmittedAt.Equal(submittedAt)
	})).Return(expected, nil)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.StudentTaskScore
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, studentID, response.StudentID)
	assert.Equal(t, 85, response.Score)

	svc.AssertExpectations(t)
}

func TestCheckerHandler_SubmitGrade_DefaultSubmittedAt(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()

	reqBody := map[string]interface{}{
		"student_id": studentID,
		"task_id":    taskID,
		"course_id":  courseID,
		"raw_score":  90,
		"is_passed":  true,
	}

	expected := &model.StudentTaskScore{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     90,
		IsPassed:  true,
	}

	e := echo.New()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc.On("SubmitGrade", mock.Anything, mock.MatchedBy(func(input service.SubmitGradeInput) bool {
		return input.StudentID == studentID &&
			input.TaskID == taskID &&
			input.CourseID == courseID &&
			input.RawScore == 90 &&
			!input.SubmittedAt.IsZero()
	})).Return(expected, nil)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	svc.AssertExpectations(t)
}

func TestCheckerHandler_SubmitGrade_MissingStudentID(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	reqBody := map[string]interface{}{
		"task_id":   uuid.New(),
		"course_id": uuid.New(),
		"raw_score": 85,
		"is_passed": true,
	}

	e := echo.New()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	svc.AssertNotCalled(t, "SubmitGrade")
}

func TestCheckerHandler_SubmitGrade_MissingTaskID(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	reqBody := map[string]interface{}{
		"student_id": uuid.New(),
		"course_id":  uuid.New(),
		"raw_score":  85,
		"is_passed":  true,
	}

	e := echo.New()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	svc.AssertNotCalled(t, "SubmitGrade")
}

func TestCheckerHandler_SubmitGrade_MissingCourseID(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	reqBody := map[string]interface{}{
		"student_id": uuid.New(),
		"task_id":    uuid.New(),
		"raw_score":  85,
		"is_passed":  true,
	}

	e := echo.New()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	svc.AssertNotCalled(t, "SubmitGrade")
}

func TestCheckerHandler_SubmitGrade_ServiceError(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()

	reqBody := map[string]interface{}{
		"student_id": studentID,
		"task_id":    taskID,
		"course_id":  courseID,
		"raw_score":  85,
		"is_passed":  true,
	}

	e := echo.New()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	svc.On("SubmitGrade", mock.Anything, mock.Anything).Return(nil, service.NotFound("Task not found"))

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	svc.AssertExpectations(t)
}

func TestCheckerHandler_SubmitGrade_InvalidRequestBody(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	svc.AssertNotCalled(t, "SubmitGrade")
}
