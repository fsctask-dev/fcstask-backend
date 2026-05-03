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

type MockCourseRepository struct {
	mock.Mock
}

func (m *MockCourseRepository) GetByID(ctx context.Context, id string) (*model.Course, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

func (m *MockCourseRepository) Update(ctx context.Context, course *model.Course) error {
	args := m.Called(ctx, course)
	return args.Error(0)
}

func (m *MockCourseRepository) Create(ctx context.Context, course *model.Course) error {
	args := m.Called(ctx, course)
	return args.Error(0)
}

func (m *MockCourseRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCourseRepository) List(ctx context.Context, filter interface{}) ([]*model.Course, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*model.Course), args.Error(1)
}

func TestAdminEditCourseHandler_Success(t *testing.T) {
	// Setup
	e := echo.New()
	courseID := uuid.New()
	course := &model.Course{
		ID:          courseID.String(),
		Name:        "Old Name",
		Description: ptrString("Old Description"),
		Type:        model.CourseTypePrivate,
		Status:      "draft",
	}

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(course, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Course")).Return(nil)

	handler := NewAdminCourseHandler(mockRepo)

	reqBody := UpdateCourseInfoRequest{
		Name:        ptrString("New Name"),
		Description: ptrString("New Description"),
		Type:        ptrType(model.CourseTypePublic),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String(), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminEditCourseHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.Course
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", response.Name)
	assert.Equal(t, "New Description", *response.Description)
	assert.Equal(t, model.CourseTypePublic, response.Type)

	mockRepo.AssertExpectations(t)
}

func TestAdminEditCourseHandler_InvalidCourseID(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepository)
	handler := NewAdminCourseHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/invalid-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("invalid-id")

	err := handler.AdminEditCourseHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Invalid course ID", response["error"])
}

func TestAdminEditCourseHandler_CourseNotFound(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(nil, assert.AnError)

	handler := NewAdminCourseHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminEditCourseHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminEditCourseHandler_InvalidType(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()
	course := &model.Course{
		ID:   courseID.String(),
		Name: "Test Course",
	}

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(course, nil)

	handler := NewAdminCourseHandler(mockRepo)

	reqBody := map[string]string{"type": "invalid"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String(), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminEditCourseHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Type must be 'public' or 'private'", response["error"])
}

func TestAdminEditCourseHandler_InvalidStartDateFormat(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()
	course := &model.Course{ID: courseID.String()}

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(course, nil)

	handler := NewAdminCourseHandler(mockRepo)

	reqBody := UpdateCourseInfoRequest{
		StartDate: ptrString("2023-13-01"),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String(), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminEditCourseHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminEditCourseHandler_InvalidDateRange(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()
	course := &model.Course{
		ID:        courseID.String(),
		StartDate: nil,
		EndDate:   nil,
	}

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(course, nil)

	handler := NewAdminCourseHandler(mockRepo)

	reqBody := UpdateCourseInfoRequest{
		StartDate: ptrString("2024-01-10"),
		EndDate:   ptrString("2024-01-05"),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String(), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminEditCourseHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "end_date must be after start_date", response["error"])
}

func TestAdminEditCourseHandler_UpdateError(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()
	course := &model.Course{ID: courseID.String(), Name: "Test"}

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(course, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Course")).Return(assert.AnError)

	handler := NewAdminCourseHandler(mockRepo)

	reqBody := UpdateCourseInfoRequest{Name: ptrString("New Name")}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String(), bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminEditCourseHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAdminUpdateCourseStatusHandler_Success(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()
	course := &model.Course{
		ID:     courseID.String(),
		Status: "draft",
	}

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(course, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Course")).Return(nil)

	handler := NewAdminCourseHandler(mockRepo)

	reqBody := UpdateCourseStatusRequest{Status: "published"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/status", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminUpdateCourseStatusHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response model.Course
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "published", response.Status)

	mockRepo.AssertExpectations(t)
}

func TestAdminUpdateCourseStatusHandler_InvalidCourseID(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepository)
	handler := NewAdminCourseHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/invalid-id/status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("invalid-id")

	err := handler.AdminUpdateCourseStatusHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminUpdateCourseStatusHandler_CourseNotFound(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(nil, assert.AnError)

	handler := NewAdminCourseHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminUpdateCourseStatusHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminUpdateCourseStatusHandler_MissingStatus(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()
	course := &model.Course{ID: courseID.String(), Status: "draft"}

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(course, nil)

	handler := NewAdminCourseHandler(mockRepo)

	reqBody := UpdateCourseStatusRequest{Status: ""}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/status", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminUpdateCourseStatusHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Status is required", response["error"])
}

func TestAdminUpdateCourseStatusHandler_InvalidStatus(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()
	course := &model.Course{ID: courseID.String(), Status: "draft"}

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(course, nil)

	handler := NewAdminCourseHandler(mockRepo)

	reqBody := UpdateCourseStatusRequest{Status: "invalid_status"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/status", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminUpdateCourseStatusHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Invalid status value", response["error"])
}

func TestAdminUpdateCourseStatusHandler_UpdateError(t *testing.T) {
	e := echo.New()
	courseID := uuid.New()
	course := &model.Course{ID: courseID.String(), Status: "draft"}

	mockRepo := new(MockCourseRepository)
	mockRepo.On("GetByID", mock.Anything, courseID.String()).Return(course, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Course")).Return(assert.AnError)

	handler := NewAdminCourseHandler(mockRepo)

	reqBody := UpdateCourseStatusRequest{Status: "published"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/status", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := handler.AdminUpdateCourseStatusHandler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func ptrString(s string) *string {
	return &s
}

func ptrType(t model.CourseType) *model.CourseType {
	return &t
}
