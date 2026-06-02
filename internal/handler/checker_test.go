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

type MockCheckerService struct{ mock.Mock }

func (m *MockCheckerService) SubmitGrade(ctx context.Context, input service.SubmitGradeInput) (*model.StudentTaskScore, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StudentTaskScore), args.Error(1)
}

func setupCheckerHandler() (*handler.CheckerHandler, *MockCheckerService) {
	svc := new(MockCheckerService)
	return handler.NewCheckerHandler(svc), svc
}

func submitGradeBody(studentID, taskID, courseID uuid.UUID, status string, submittedAt time.Time) []byte {
	b, _ := json.Marshal(map[string]interface{}{
		"student_id":   studentID,
		"task_id":      taskID,
		"course_id":    courseID,
		"status":       status,
		"submitted_at": submittedAt,
	})
	return b
}

func TestCheckerHandler_SubmitGrade_Success(t *testing.T) {
	e := echo.New()
	h, svc := setupCheckerHandler()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	score := &model.StudentTaskScore{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     100,
		IsPassed:  true,
	}
	svc.On("SubmitGrade", mock.Anything, mock.MatchedBy(func(in service.SubmitGradeInput) bool {
		return in.StudentID == studentID && in.TaskID == taskID && in.Status == "passed"
	})).Return(score, nil)

	body := submitGradeBody(studentID, taskID, courseID, "passed", now)
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.SubmitGrade(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var result model.StudentTaskScore
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, 100, result.Score)
	assert.True(t, result.IsPassed)
}

func TestCheckerHandler_SubmitGrade_MissingStudentID(t *testing.T) {
	e := echo.New()
	h, _ := setupCheckerHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"task_id":   uuid.New(),
		"course_id": uuid.New(),
		"status":    "passed",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.SubmitGrade(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCheckerHandler_SubmitGrade_MissingTaskID(t *testing.T) {
	e := echo.New()
	h, _ := setupCheckerHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"student_id": uuid.New(),
		"course_id":  uuid.New(),
		"status":     "passed",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.SubmitGrade(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCheckerHandler_SubmitGrade_MissingCourseID(t *testing.T) {
	e := echo.New()
	h, _ := setupCheckerHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"student_id": uuid.New(),
		"task_id":    uuid.New(),
		"status":     "passed",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.SubmitGrade(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCheckerHandler_MissingSubmittedAt(t *testing.T) {
	e := echo.New()
	h, _ := setupCheckerHandler()
	body, _ := json.Marshal(map[string]interface{}{
		"student_id": uuid.New(),
		"task_id":    uuid.New(),
		"course_id":  uuid.New(),
		"status":     "passed",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	assert.NoError(t, h.SubmitGrade(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCheckerHandler_SubmitGrade_ServiceError(t *testing.T) {
	e := echo.New()
	h, svc := setupCheckerHandler()

	svc.On("SubmitGrade", mock.Anything, mock.Anything).Return(nil, service.NotFound("Task not found"))

	body := submitGradeBody(uuid.New(), uuid.New(), uuid.New(), "passed", time.Now())
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.SubmitGrade(c))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCheckerHandler_SubmitGrade_FailStatus(t *testing.T) {
	e := echo.New()
	h, svc := setupCheckerHandler()

	score := &model.StudentTaskScore{Score: 0, IsPassed: false}
	svc.On("SubmitGrade", mock.Anything, mock.MatchedBy(func(in service.SubmitGradeInput) bool {
		return in.Status == "fail"
	})).Return(score, nil)

	body := submitGradeBody(uuid.New(), uuid.New(), uuid.New(), "fail", time.Now())
	req := httptest.NewRequest(http.MethodPost, "/api/grades", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.SubmitGrade(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	var result model.StudentTaskScore
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.False(t, result.IsPassed)
	assert.Equal(t, 0, result.Score)
}
