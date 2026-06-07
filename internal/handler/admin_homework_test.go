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

func TestAdminHomeworkHandler_CreateHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	title := "Week 1"
	description := "Test description"
	position := 1
	startDate := "2025-01-01"
	endDate := "2025-06-01"

	body := map[string]interface{}{
		"title":       title,
		"description": description,
		"position":    position,
		"start_date":  startDate,
		"end_date":    endDate,
	}

	expected := &model.Homework{HwID: uuid.New(), CourseID: courseID, Title: title}

	c, rec := newEchoContext(http.MethodPost, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("CreateHomework", mock.Anything, mock.Anything, mock.MatchedBy(func(input service.CreateHomeworkInput) bool {
		return input.CourseID == courseID &&
			input.Title == title &&
			input.Description == description &&
			input.Position == position &&
			input.StartDate == startDate &&
			input.EndDate == endDate
	})).Return(expected, nil)

	err := h.CreateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_CreateHomework_InvalidCourseID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodPost, "/", nil, map[string]string{"courseId": "not-a-uuid"})

	err := h.CreateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid course ID")
}

func TestAdminHomeworkHandler_CreateHomework_MissingCourseID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodPost, "/", nil, map[string]string{})

	err := h.CreateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminHomeworkHandler_CreateHomework_ServiceError(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	body := map[string]interface{}{
		"title":      "Test",
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

func TestAdminHomeworkHandler_GetHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	expected := &model.Homework{HwID: hwID, Title: "Test Homework"}

	c, rec := newEchoContextMultiParam(http.MethodGet, "/", nil,
		[]string{"courseId", "hwId"},
		[]string{uuid.New().String(), hwID.String()},
	)
	svc.On("GetHomework", mock.Anything, mock.Anything, hwID).Return(expected, nil)

	err := h.GetHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.Homework
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, hwID, response.HwID)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_GetHomework_InvalidHwID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"hwId": "invalid"})

	err := h.GetHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid homework ID")
}

func TestAdminHomeworkHandler_GetHomework_NotFound(t *testing.T) {
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

func TestAdminHomeworkHandler_ListHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	expected := []model.Homework{
		{HwID: uuid.New(), CourseID: courseID, Title: "HW1"},
		{HwID: uuid.New(), CourseID: courseID, Title: "HW2"},
	}

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"courseId": courseID.String()})
	svc.On("ListHomework", mock.Anything, mock.Anything, courseID).Return(expected, nil)

	err := h.ListHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response []model.Homework
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_ListHomework_EmptyList(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	courseID := uuid.New()
	expected := []model.Homework{}

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"courseId": courseID.String()})
	svc.On("ListHomework", mock.Anything, mock.Anything, courseID).Return(expected, nil)

	err := h.ListHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "[]\n", rec.Body.String())
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_ListHomework_InvalidCourseID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"courseId": "invalid"})

	err := h.ListHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid course ID")
}

func TestAdminHomeworkHandler_UpdateHomework_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	newTitle := "Updated Title"
	newEndDate := "2026-01-01"
	body := map[string]interface{}{
		"title":    newTitle,
		"end_date": newEndDate,
	}
	expected := &model.Homework{HwID: hwID, Title: newTitle}

	c, rec := newEchoContextMultiParam(http.MethodPatch, "/", body,
		[]string{"courseId", "hwId"},
		[]string{uuid.New().String(), hwID.String()},
	)
	svc.On("UpdateHomework", mock.Anything, mock.Anything, hwID, mock.MatchedBy(func(input service.UpdateHomeworkInput) bool {
		return input.Title != nil && *input.Title == newTitle &&
			input.EndDate == newEndDate
	})).Return(expected, nil)

	err := h.UpdateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_UpdateHomework_UpdatePositionToZero(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	pos := 0
	newTitle := "Week 0"
	body := map[string]interface{}{
		"title":    newTitle,
		"position": pos,
	}
	expected := &model.Homework{HwID: hwID, Title: newTitle, Position: pos}

	c, rec := newEchoContextMultiParam(http.MethodPatch, "/", body,
		[]string{"courseId", "hwId"},
		[]string{uuid.New().String(), hwID.String()},
	)
	svc.On("UpdateHomework", mock.Anything, mock.Anything, hwID, mock.MatchedBy(func(input service.UpdateHomeworkInput) bool {
		return input.Title != nil && *input.Title == newTitle &&
			input.Position != nil && *input.Position == 0
	})).Return(expected, nil)

	err := h.UpdateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_UpdateHomework_ClearDescription(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	emptyDesc := ""
	body := map[string]interface{}{
		"description": emptyDesc,
	}
	expected := &model.Homework{HwID: hwID, Description: nil}

	c, rec := newEchoContextMultiParam(http.MethodPatch, "/", body,
		[]string{"courseId", "hwId"},
		[]string{uuid.New().String(), hwID.String()},
	)
	svc.On("UpdateHomework", mock.Anything, mock.Anything, hwID, mock.MatchedBy(func(input service.UpdateHomeworkInput) bool {
		return input.Description != nil && *input.Description == ""
	})).Return(expected, nil)

	err := h.UpdateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_UpdateHomework_InvalidHwID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodPatch, "/", nil, map[string]string{"hwId": "invalid"})

	err := h.UpdateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminHomeworkHandler_UpdateHomework_NotFound(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	body := map[string]interface{}{"title": "New Title"}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"hwId": hwID.String()})
	svc.On("UpdateHomework", mock.Anything, mock.Anything, hwID, mock.Anything).Return(nil, service.NotFound("Homework not found"))

	err := h.UpdateHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_DeleteHomework_Success(t *testing.T) {
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

func TestAdminHomeworkHandler_DeleteHomework_InvalidHwID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"hwId": "invalid"})

	err := h.DeleteHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminHomeworkHandler_DeleteHomework_NotFound(t *testing.T) {
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

func TestAdminHomeworkHandler_PublishHomework_Success(t *testing.T) {
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

func TestAdminHomeworkHandler_PublishHomework_SetFalse(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	isPublic := false
	body := map[string]interface{}{"is_public": isPublic}
	expected := &model.Homework{HwID: hwID, IsPublic: &isPublic}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"hwId": hwID.String()})
	svc.On("PublishHomework", mock.Anything, mock.Anything, hwID, isPublic).Return(expected, nil)

	err := h.PublishHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_PublishHomework_InvalidHwID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodPatch, "/", nil, map[string]string{"hwId": "invalid"})

	err := h.PublishHomework(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminHomeworkHandler_GetDeadlineByHomeworkID_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	expected := &model.Deadline{ID: uuid.New(), HomeworkID: hwID, Title: "Deadline"}

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"hwId": hwID.String()})
	svc.On("GetDeadlineByHomeworkID", mock.Anything, mock.Anything, hwID).Return(expected, nil)

	err := h.GetDeadlineByHomeworkID(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_GetDeadlineByHomeworkID_InvalidHwID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"hwId": "invalid"})

	err := h.GetDeadlineByHomeworkID(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminHomeworkHandler_GetDeadlineByHomeworkID_NotFound(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"hwId": hwID.String()})
	svc.On("GetDeadlineByHomeworkID", mock.Anything, mock.Anything, hwID).Return(nil, service.NotFound("Deadline not found"))

	err := h.GetDeadlineByHomeworkID(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_SetDeadline_MissingCourseID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	hwID := uuid.New()
	body := map[string]interface{}{
		"title": "Deadline",
	}

	c, rec := newEchoContextMultiParam(http.MethodPut, "/", body,
		[]string{"hwId"},
		[]string{hwID.String()},
	)

	err := h.SetDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "course_id is required")
}

func TestAdminHomeworkHandler_SetDeadline_InvalidHomeworkID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	body := map[string]interface{}{
		"course_id": uuid.New().String(),
		"title":     "Deadline",
	}

	c, rec := newEchoContextMultiParam(http.MethodPut, "/", body,
		[]string{"hwId"},
		[]string{"invalid"},
	)

	err := h.SetDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid homework ID")
}

func TestAdminHomeworkHandler_UpdateDeadline_Success(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	dlID := uuid.New()
	newTitle := "Updated Title"
	newSoftDeadline := time.Now().UTC().Add(48 * time.Hour)

	body := map[string]interface{}{
		"title":         newTitle,
		"soft_deadline": newSoftDeadline,
	}
	expected := &model.Deadline{ID: dlID, Title: newTitle}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"deadlineId": dlID.String()})
	svc.On("UpdateDeadline", mock.Anything, mock.Anything, dlID, mock.MatchedBy(func(input service.UpdateDeadlineInput) bool {
		return input.Title == newTitle && input.SoftDeadline.Equal(newSoftDeadline)
	})).Return(expected, nil)

	err := h.UpdateDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_UpdateDeadline_PartialUpdate(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	dlID := uuid.New()
	newHardDeadline := time.Now().UTC().Add(72 * time.Hour)

	body := map[string]interface{}{
		"hard_deadline": newHardDeadline,
	}
	expected := &model.Deadline{ID: dlID}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"deadlineId": dlID.String()})
	svc.On("UpdateDeadline", mock.Anything, mock.Anything, dlID, mock.MatchedBy(func(input service.UpdateDeadlineInput) bool {
		return input.HardDeadline.Equal(newHardDeadline) && input.Title == "" && input.Description == ""
	})).Return(expected, nil)

	err := h.UpdateDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_UpdateDeadline_InvalidDeadlineID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodPatch, "/", nil, map[string]string{"deadlineId": "invalid"})

	err := h.UpdateDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminHomeworkHandler_UpdateDeadline_NotFound(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	dlID := uuid.New()
	body := map[string]interface{}{"title": "New Title"}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"deadlineId": dlID.String()})
	svc.On("UpdateDeadline", mock.Anything, mock.Anything, dlID, mock.Anything).Return(nil, service.NotFound("Deadline not found"))

	err := h.UpdateDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestAdminHomeworkHandler_DeleteDeadline_Success(t *testing.T) {
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

func TestAdminHomeworkHandler_DeleteDeadline_InvalidDeadlineID(t *testing.T) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)

	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"deadlineId": "invalid"})

	err := h.DeleteDeadline(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminHomeworkHandler_DeleteDeadline_NotFound(t *testing.T) {
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
