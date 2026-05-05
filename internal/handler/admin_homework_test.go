package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/handler"
)

type MockHomeworkRepo struct {
	mock.Mock
}

func (m *MockHomeworkRepo) Create(ctx context.Context, hw *model.Homework) error {
	return m.Called(ctx, hw).Error(0)
}

func (m *MockHomeworkRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Homework, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Homework), args.Error(1)
}

func (m *MockHomeworkRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.Homework, error) {
	args := m.Called(ctx, courseID)
	return args.Get(0).([]model.Homework), args.Error(1)
}

func (m *MockHomeworkRepo) Update(ctx context.Context, hw *model.Homework) error {
	return m.Called(ctx, hw).Error(0)
}

func (m *MockHomeworkRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type MockDeadlineRepo struct {
	mock.Mock
}

func (m *MockDeadlineRepo) Create(ctx context.Context, deadline *model.Deadline) error {
	return m.Called(ctx, deadline).Error(0)
}

func (m *MockDeadlineRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Deadline, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deadline), args.Error(1)
}

func (m *MockDeadlineRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.Deadline, error) {
	args := m.Called(ctx, courseID)
	return args.Get(0).([]model.Deadline), args.Error(1)
}

func (m *MockDeadlineRepo) Update(ctx context.Context, deadline *model.Deadline) error {
	return m.Called(ctx, deadline).Error(0)
}

func (m *MockDeadlineRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestAdminCreateHomework_Success(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	courseID := uuid.New()
	body := map[string]interface{}{
		"course_id":  courseID,
		"start_date": "2025-01-01",
		"end_date":   "2025-06-01",
	}
	req, rec := newRequest(http.MethodPost, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	hwRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Homework")).Return(nil)

	err := h.AdminCreateHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestAdminCreateHomework_InvalidCourseID(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	req, rec := newRequest(http.MethodPost, "/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("bad-uuid")

	err := h.AdminCreateHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminCreateHomework_InvalidStartDate(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	courseID := uuid.New()
	body := map[string]interface{}{"start_date": "not-a-date"}
	req, rec := newRequest(http.MethodPost, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := h.AdminCreateHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminCreateHomework_EndBeforeStart(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	courseID := uuid.New()
	body := map[string]interface{}{
		"start_date": "2025-12-01",
		"end_date":   "2025-01-01",
	}
	req, rec := newRequest(http.MethodPost, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := h.AdminCreateHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminGetHomework_Success(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	hwID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	hw := &model.Homework{HwID: hwID}
	hwRepo.On("GetByID", mock.Anything, hwID).Return(hw, nil)

	err := h.AdminGetHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminGetHomework_InvalidID(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	req, rec := newRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("hwId")
	c.SetParamValues("bad-uuid")

	err := h.AdminGetHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminGetHomework_NotFound(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	hwID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	hwRepo.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)

	err := h.AdminGetHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminListHomework_Success(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	courseID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	hwRepo.On("GetByCourseID", mock.Anything, courseID).Return([]model.Homework{{HwID: uuid.New()}}, nil)

	err := h.AdminListHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminUpdateHomework_Success(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	hwID := uuid.New()
	body := map[string]interface{}{"end_date": "2026-01-01"}
	req, rec := newRequest(http.MethodPatch, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	hw := &model.Homework{HwID: hwID}
	hwRepo.On("GetByID", mock.Anything, hwID).Return(hw, nil)
	hwRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Homework")).Return(nil)

	err := h.AdminUpdateHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminUpdateHomework_EndBeforeStart(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	hwID := uuid.New()
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	body := map[string]interface{}{"end_date": "2025-01-01"}
	req, rec := newRequest(http.MethodPatch, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	hw := &model.Homework{HwID: hwID, StartDate: &start}
	hwRepo.On("GetByID", mock.Anything, hwID).Return(hw, nil)

	err := h.AdminUpdateHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminDeleteHomework_Success(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	hwID := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	hw := &model.Homework{HwID: hwID}
	hwRepo.On("GetByID", mock.Anything, hwID).Return(hw, nil)
	hwRepo.On("Delete", mock.Anything, hwID).Return(nil)

	err := h.AdminDeleteHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAdminDeleteHomework_NotFound(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	hwID := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	hwRepo.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)

	err := h.AdminDeleteHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminPublishHomework_Success(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	hwID := uuid.New()
	body := map[string]interface{}{"is_public": true}
	req, rec := newRequest(http.MethodPatch, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	hw := &model.Homework{HwID: hwID}
	hwRepo.On("GetByID", mock.Anything, hwID).Return(hw, nil)
	hwRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Homework")).Return(nil)

	err := h.AdminPublishHomeworkHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result model.Homework
	_ = json.NewDecoder(rec.Body).Decode(&result)
	assert.NotNil(t, result.IsPublic)
	assert.True(t, *result.IsPublic)
}

func TestAdminSetHomeworkDeadline_Success(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	body := map[string]interface{}{
		"title":    "Deadline 1",
		"due_date": "2025-12-31T23:59:59Z",
	}
	req, rec := newRequest(http.MethodPut, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), hwID.String())

	hw := &model.Homework{HwID: hwID}
	hwRepo.On("GetByID", mock.Anything, hwID).Return(hw, nil)
	dlRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Deadline")).Return(nil)

	err := h.AdminSetHomeworkDeadlineHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestAdminSetHomeworkDeadline_MissingTitle(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	body := map[string]interface{}{"due_date": "2025-12-31T23:59:59Z"}
	req, rec := newRequest(http.MethodPut, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), hwID.String())

	hw := &model.Homework{HwID: hwID}
	hwRepo.On("GetByID", mock.Anything, hwID).Return(hw, nil)

	err := h.AdminSetHomeworkDeadlineHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminSetHomeworkDeadline_InvalidDueDate(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	body := map[string]interface{}{
		"title":    "X",
		"due_date": "not-a-date",
	}
	req, rec := newRequest(http.MethodPut, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), hwID.String())

	hw := &model.Homework{HwID: hwID}
	hwRepo.On("GetByID", mock.Anything, hwID).Return(hw, nil)

	err := h.AdminSetHomeworkDeadlineHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminUpdateDeadline_Success(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	dlID := uuid.New()
	newTitle := "Updated Title"
	body := map[string]interface{}{"title": newTitle}
	req, rec := newRequest(http.MethodPatch, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("deadlineId")
	c.SetParamValues(dlID.String())

	dl := &model.Deadline{ID: dlID, Title: "Old Title"}
	dlRepo.On("GetByID", mock.Anything, dlID).Return(dl, nil)
	dlRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Deadline")).Return(nil)

	err := h.AdminUpdateDeadlineHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminUpdateDeadline_EmptyTitle(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	dlID := uuid.New()
	body := map[string]interface{}{"title": ""}
	req, rec := newRequest(http.MethodPatch, "/", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("deadlineId")
	c.SetParamValues(dlID.String())

	dl := &model.Deadline{ID: dlID, Title: "Old Title"}
	dlRepo.On("GetByID", mock.Anything, dlID).Return(dl, nil)

	err := h.AdminUpdateDeadlineHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminDeleteDeadline_Success(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	dlID := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("deadlineId")
	c.SetParamValues(dlID.String())

	dl := &model.Deadline{ID: dlID}
	dlRepo.On("GetByID", mock.Anything, dlID).Return(dl, nil)
	dlRepo.On("Delete", mock.Anything, dlID).Return(nil)

	err := h.AdminDeleteDeadlineHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAdminDeleteDeadline_NotFound(t *testing.T) {
	e := setupEcho()
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	h := handler.NewAdminHomeworkHandler(hwRepo, dlRepo)

	dlID := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("deadlineId")
	c.SetParamValues(dlID.String())

	dlRepo.On("GetByID", mock.Anything, dlID).Return(nil, assert.AnError)

	err := h.AdminDeleteDeadlineHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

var _ = httptest.NewRecorder
