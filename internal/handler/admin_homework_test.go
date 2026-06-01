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

type MockAdminHomeworkService struct {
	mock.Mock
}

func (m *MockAdminHomeworkService) CreateHomework(ctx context.Context, userID uuid.UUID, input service.CreateHomeworkInput) (*model.Homework, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Homework), args.Error(1)
}

func (m *MockAdminHomeworkService) GetHomework(ctx context.Context, userID, hwID uuid.UUID) (*model.Homework, error) {
	args := m.Called(ctx, userID, hwID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Homework), args.Error(1)
}

func (m *MockAdminHomeworkService) ListHomework(ctx context.Context, userID, courseID uuid.UUID) ([]model.Homework, error) {
	args := m.Called(ctx, userID, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Homework), args.Error(1)
}

func (m *MockAdminHomeworkService) UpdateHomework(ctx context.Context, userID, hwID uuid.UUID, input service.UpdateHomeworkInput) (*model.Homework, error) {
	args := m.Called(ctx, userID, hwID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Homework), args.Error(1)
}

func (m *MockAdminHomeworkService) DeleteHomework(ctx context.Context, userID, hwID uuid.UUID) error {
	args := m.Called(ctx, userID, hwID)
	return args.Error(0)
}

func (m *MockAdminHomeworkService) PublishHomework(ctx context.Context, userID, hwID uuid.UUID, isPublic bool) (*model.Homework, error) {
	args := m.Called(ctx, userID, hwID, isPublic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Homework), args.Error(1)
}

func (m *MockAdminHomeworkService) SetDeadline(ctx context.Context, userID uuid.UUID, input service.SetDeadlineInput) (*model.Deadline, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deadline), args.Error(1)
}

func (m *MockAdminHomeworkService) UpdateDeadline(ctx context.Context, userID, deadlineID uuid.UUID, input service.UpdateDeadlineInput) (*model.Deadline, error) {
	args := m.Called(ctx, userID, deadlineID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deadline), args.Error(1)
}

func (m *MockAdminHomeworkService) DeleteDeadline(ctx context.Context, userID, deadlineID uuid.UUID) error {
	args := m.Called(ctx, userID, deadlineID)
	return args.Error(0)
}

func (m *MockAdminHomeworkService) GetDeadlineByHomeworkID(ctx context.Context, userID, hwID uuid.UUID) (*model.Deadline, error) {
	args := m.Called(ctx, userID, hwID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deadline), args.Error(1)
}

func newEchoContext(method, path string, body interface{}, params map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	for k, v := range params {
		c.SetParamNames(k)
		c.SetParamValues(v)
	}
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})
	return c, rec
}

func newEchoContextMultiParam(method, path string, body interface{}, paramNames []string, paramValues []string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames(paramNames...)
	c.SetParamValues(paramValues...)
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})
	return c, rec
}

func TestHandlerCreateHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	hwID := uuid.New()
	title := "Week 1"
	description := "Test description"
	position := 1
	startDate := "2025-01-01"
	endDate := "2025-06-01"
	softDeadline := time.Now().UTC()
	hardDeadline := time.Now().UTC().Add(24 * time.Hour)

	body := map[string]interface{}{
		"title":         title,
		"description":   description,
		"position":      position,
		"start_date":    startDate,
		"end_date":      endDate,
		"soft_deadline": softDeadline,
		"hard_deadline": hardDeadline,
	}

	expected := &model.Homework{HwID: hwID, CourseID: courseID, Title: title}

	c, rec := newEchoContext(http.MethodPost, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("CreateHomework", mock.Anything, mock.Anything, mock.MatchedBy(func(input service.CreateHomeworkInput) bool {
		return input.CourseID == courseID &&
			input.Title == title &&
			input.Description == description &&
			input.Position == position &&
			input.StartDate == startDate &&
			input.EndDate == endDate &&
			input.SoftDeadline.Equal(softDeadline) &&
			input.HardDeadline.Equal(hardDeadline)
	})).Return(expected, nil)

	err := h.CreateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerCreateHomework_InvalidCourseID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodPost, "/", nil, map[string]string{"courseId": "not-a-uuid"})

	err := h.CreateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerCreateHomework_ServiceError(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	body := map[string]interface{}{
		"start_date": "2025-01-01",
		"end_date":   "2025-06-01",
	}

	c, rec := newEchoContext(http.MethodPost, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("CreateHomework", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.BadRequest("end date must be after start_date"))

	err := h.CreateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerGetHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	expected := &model.Homework{HwID: hwID}

	c, rec := newEchoContextMultiParam(http.MethodGet, "/", nil,
		[]string{"courseId", "hwId"},
		[]string{uuid.New().String(), hwID.String()},
	)
	svc.On("GetHomework", mock.Anything, mock.Anything, hwID).Return(expected, nil)

	err := h.GetHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerGetHomework_InvalidID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"hwId": "invalid"})

	err := h.GetHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerGetHomework_NotFound(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"hwId": hwID.String()})
	svc.On("GetHomework", mock.Anything, mock.Anything, hwID).Return(nil, service.NotFound("Homework not found"))

	err := h.GetHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	expected := []model.Homework{{HwID: uuid.New(), CourseID: courseID}}

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"courseId": courseID.String()})
	svc.On("ListHomework", mock.Anything, mock.Anything, courseID).Return(expected, nil)

	err := h.ListHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListHomework_InvalidCourseID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"courseId": "bad"})

	err := h.ListHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerUpdateHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	newEnd := "2026-01-01"
	body := map[string]interface{}{"end_date": newEnd}
	expected := &model.Homework{HwID: hwID}

	c, rec := newEchoContextMultiParam(http.MethodPatch, "/", body,
		[]string{"courseId", "hwId"},
		[]string{uuid.New().String(), hwID.String()},
	)
	svc.On("UpdateHomework", mock.Anything, mock.Anything, hwID, service.UpdateHomeworkInput{EndDate: newEnd}).Return(expected, nil)

	err := h.UpdateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerUpdateHomework_InvalidID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodPatch, "/", nil, map[string]string{"hwId": "bad"})

	err := h.UpdateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerUpdateHomework_ServiceError(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	body := map[string]interface{}{"end_date": "2025-01-01"}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"hwId": hwID.String()})
	svc.On("UpdateHomework", mock.Anything, mock.Anything, hwID, mock.Anything).Return(nil, service.NotFound("Homework not found"))

	err := h.UpdateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerDeleteHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"hwId": hwID.String()})
	svc.On("DeleteHomework", mock.Anything, mock.Anything, hwID).Return(nil)

	err := h.DeleteHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerDeleteHomework_InvalidID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"hwId": "bad"})

	err := h.DeleteHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerDeleteHomework_NotFound(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"hwId": hwID.String()})
	svc.On("DeleteHomework", mock.Anything, mock.Anything, hwID).Return(service.NotFound("Homework not found"))

	err := h.DeleteHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerPublishHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	isPublic := true
	body := map[string]interface{}{"is_public": isPublic}
	expected := &model.Homework{HwID: hwID, IsPublic: &isPublic}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"hwId": hwID.String()})
	svc.On("PublishHomework", mock.Anything, mock.Anything, hwID, isPublic).Return(expected, nil)

	err := h.PublishHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerPublishHomework_InvalidID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodPatch, "/", nil, map[string]string{"hwId": "bad"})

	err := h.PublishHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerSetDeadline_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	hwID := uuid.New()
	body := map[string]interface{}{
		"title":    "Deadline 1",
		"due_date": "2025-12-31T23:59:59Z",
	}
	expected := &model.Deadline{Title: "Deadline 1", CourseID: courseID}

	c, rec := newEchoContextMultiParam(http.MethodPut, "/", body,
		[]string{"courseId", "hwId"},
		[]string{courseID.String(), hwID.String()},
	)
	svc.On("SetDeadline", mock.Anything, mock.Anything, mock.MatchedBy(func(inp service.SetDeadlineInput) bool {
		return inp.CourseID == courseID && inp.HomeworkID == hwID && inp.Title == "Deadline 1"
	})).Return(expected, nil)

	err := h.SetDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerSetDeadline_InvalidCourseID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContextMultiParam(http.MethodPut, "/", nil,
		[]string{"courseId", "hwId"},
		[]string{"bad", uuid.New().String()},
	)

	err := h.SetDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerSetDeadline_InvalidHomeworkID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContextMultiParam(http.MethodPut, "/", nil,
		[]string{"courseId", "hwId"},
		[]string{uuid.New().String(), "bad"},
	)

	err := h.SetDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerSetDeadline_ServiceError(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	hwID := uuid.New()
	body := map[string]interface{}{"title": "", "due_date": "2025-12-31T23:59:59Z"}

	c, rec := newEchoContextMultiParam(http.MethodPut, "/", body,
		[]string{"courseId", "hwId"},
		[]string{courseID.String(), hwID.String()},
	)
	svc.On("SetDeadline", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.BadRequest("title is required"))

	err := h.SetDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerUpdateDeadline_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	dlID := uuid.New()
	newTitle := "Updated Title"
	body := map[string]interface{}{"title": newTitle}
	expected := &model.Deadline{ID: dlID, Title: newTitle}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"deadlineId": dlID.String()})
	svc.On("UpdateDeadline", mock.Anything, mock.Anything, dlID, service.UpdateDeadlineInput{Title: newTitle}).Return(expected, nil)

	err := h.UpdateDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerUpdateDeadline_InvalidID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodPatch, "/", nil, map[string]string{"deadlineId": "bad"})

	err := h.UpdateDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerUpdateDeadline_NotFound(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	dlID := uuid.New()
	body := map[string]interface{}{"title": "New"}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"deadlineId": dlID.String()})
	svc.On("UpdateDeadline", mock.Anything, mock.Anything, dlID, mock.Anything).Return(nil, service.NotFound("Deadline not found"))

	err := h.UpdateDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerDeleteDeadline_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	dlID := uuid.New()
	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"deadlineId": dlID.String()})
	svc.On("DeleteDeadline", mock.Anything, mock.Anything, dlID).Return(nil)

	err := h.DeleteDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerDeleteDeadline_InvalidID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"deadlineId": "bad"})

	err := h.DeleteDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerDeleteDeadline_NotFound(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	dlID := uuid.New()
	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"deadlineId": dlID.String()})
	svc.On("DeleteDeadline", mock.Anything, mock.Anything, dlID).Return(service.NotFound("Deadline not found"))

	err := h.DeleteDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerCreateHomework_MissingRequiredFields(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	body := map[string]interface{}{
		"start_date": "2025-01-01",
		"end_date":   "2025-06-01",
	}

	c, rec := newEchoContext(http.MethodPost, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("CreateHomework", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.BadRequest("title is required"))

	err := h.CreateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerCreateHomework_InvalidSoftDeadline(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	body := map[string]interface{}{
		"title":         "Week 1",
		"start_date":    "2025-01-01",
		"end_date":      "2025-06-01",
		"soft_deadline": "invalid",
		"hard_deadline": time.Now().UTC(),
	}

	c, rec := newEchoContext(http.MethodPost, "/", body, map[string]string{"courseId": courseID.String()})
	err := h.CreateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "bad_request")
}

func TestHandlerCreateHomework_ServiceDeadlineValidationError(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	softDeadline := time.Now().UTC()
	hardDeadline := softDeadline.Add(-24 * time.Hour)

	body := map[string]interface{}{
		"title":         "Week 1",
		"start_date":    "2025-01-01",
		"end_date":      "2025-06-01",
		"soft_deadline": softDeadline,
		"hard_deadline": hardDeadline,
	}

	c, rec := newEchoContext(http.MethodPost, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("CreateHomework", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.BadRequest("hard_deadline must be after soft_deadline"))

	err := h.CreateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertExpectations(t)
}
