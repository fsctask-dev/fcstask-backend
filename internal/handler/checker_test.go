package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestCheckerHandlerSubmitGrade_Success(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()

	body := map[string]interface{}{
		"student_id":   studentID,
		"task_id":      taskID,
		"course_id":    courseID,
		"status":       "passed",
		"submitted_at": time.Now().Format(time.RFC3339),
	}

	expected := &model.StudentTaskScore{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     100,
		IsPassed:  true,
	}

	c, rec := newEchoContext(http.MethodPost, "/api/grades", body, nil)
	svc.On("SubmitGrade", mock.Anything, mock.MatchedBy(func(inp service.SubmitGradeInput) bool {
		return inp.StudentID == studentID &&
			inp.TaskID == taskID &&
			inp.CourseID == courseID &&
			inp.Status == "passed"
	})).Return(expected, nil)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestCheckerHandlerSubmitGrade_MissingStudentID(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	body := map[string]interface{}{
		"task_id":   uuid.New(),
		"course_id": uuid.New(),
		"status":    "passed",
	}

	c, rec := newEchoContext(http.MethodPost, "/api/grades", body, nil)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCheckerHandlerSubmitGrade_MissingTaskID(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	body := map[string]interface{}{
		"student_id": uuid.New(),
		"course_id":  uuid.New(),
		"status":     "passed",
	}

	c, rec := newEchoContext(http.MethodPost, "/api/grades", body, nil)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCheckerHandlerSubmitGrade_MissingCourseID(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	body := map[string]interface{}{
		"student_id": uuid.New(),
		"task_id":    uuid.New(),
		"status":     "passed",
	}

	c, rec := newEchoContext(http.MethodPost, "/api/grades", body, nil)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCheckerHandlerSubmitGrade_SubmittedAtDefaultsToNow(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()

	body := map[string]interface{}{
		"student_id": studentID,
		"task_id":    taskID,
		"course_id":  courseID,
		"status":     "fail",
	}

	expected := &model.StudentTaskScore{Score: 0, IsPassed: false}

	c, rec := newEchoContext(http.MethodPost, "/api/grades", body, nil)
	svc.On("SubmitGrade", mock.Anything, mock.MatchedBy(func(inp service.SubmitGradeInput) bool {
		return !inp.SubmittedAt.IsZero()
	})).Return(expected, nil)

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestCheckerHandlerSubmitGrade_ServiceError(t *testing.T) {
	svc := new(MockCheckerService)
	h := handler.NewCheckerHandler(svc)

	body := map[string]interface{}{
		"student_id": uuid.New(),
		"task_id":    uuid.New(),
		"course_id":  uuid.New(),
		"status":     "passed",
	}

	c, rec := newEchoContext(http.MethodPost, "/api/grades", body, nil)
	svc.On("SubmitGrade", mock.Anything, mock.Anything).
		Return(nil, service.NotFound("Task not found"))

	err := h.SubmitGrade(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}
