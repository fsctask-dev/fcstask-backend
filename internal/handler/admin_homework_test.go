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

type MockAdminHomeworkService struct{ mock.Mock }

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
	return m.Called(ctx, userID, hwID).Error(0)
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
	return m.Called(ctx, userID, deadlineID).Error(0)
}
func (m *MockAdminHomeworkService) GetDeadlineByHomeworkID(ctx context.Context, userID, hwID uuid.UUID) (*model.Deadline, error) {
	args := m.Called(ctx, userID, hwID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deadline), args.Error(1)
}

func setupHomeworkHandler() (*handler.AdminHomeworkHandler, *MockAdminHomeworkService) {
	svc := new(MockAdminHomeworkService)
	h := handler.NewAdminHomeworkHandler(svc)
	return h, svc
}

func testUser() *model.User {
	return &model.User{ID: uuid.New()}
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

func TestCreateHomeworkHandler_Success(t *testing.T) {
	e := echo.New()
	h, svc := setupHomeworkHandler()

	courseID := uuid.New()
	soft := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	hard := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
	user := testUser()

	body, _ := json.Marshal(map[string]interface{}{
		"start_date":    "2025-01-01",
		"end_date":      "2025-06-01",
		"soft_deadline": soft.Format(time.RFC3339),
		"hard_deadline": hard.Format(time.RFC3339),
	})

	hw := &model.Homework{HwID: uuid.New(), CourseID: courseID}
	svc.On("CreateHomework", mock.Anything, user.ID, mock.MatchedBy(func(in service.CreateHomeworkInput) bool {
		return in.CourseID == courseID
	})).Return(hw, nil)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	assert.NoError(t, h.CreateHomework(c))
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateHomeworkHandler_NoUser(t *testing.T) {
	e := echo.New()
	h, _ := setupHomeworkHandler()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(uuid.New().String())
	assert.NoError(t, h.CreateHomework(c))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateHomeworkHandler_InvalidCourseID(t *testing.T) {
	e := echo.New()
	h, _ := setupHomeworkHandler()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, testUser())
	c.SetParamNames("courseId")
	c.SetParamValues("not-a-uuid")
	assert.NoError(t, h.CreateHomework(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetHomeworkHandler_Success(t *testing.T) {
	e := echo.New()
	h, svc := setupHomeworkHandler()
	user := testUser()
	hwID := uuid.New()
	hw := &model.Homework{HwID: hwID, CourseID: uuid.New()}
	svc.On("GetHomework", mock.Anything, user.ID, hwID).Return(hw, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	assert.NoError(t, h.GetHomework(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetHomeworkHandler_NotFound(t *testing.T) {
	e := echo.New()
	h, svc := setupHomeworkHandler()
	user := testUser()
	hwID := uuid.New()
	svc.On("GetHomework", mock.Anything, user.ID, hwID).Return(nil, service.NotFound("Homework not found"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	assert.NoError(t, h.GetHomework(c))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListHomeworkHandler_Success(t *testing.T) {
	e := echo.New()
	h, svc := setupHomeworkHandler()
	user := testUser()
	courseID := uuid.New()
	svc.On("ListHomework", mock.Anything, user.ID, courseID).Return([]model.Homework{{HwID: uuid.New()}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	assert.NoError(t, h.ListHomework(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	var result []model.Homework
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Len(t, result, 1)
}

func TestDeleteHomeworkHandler_Success(t *testing.T) {
	e := echo.New()
	h, svc := setupHomeworkHandler()
	user := testUser()
	hwID := uuid.New()
	svc.On("DeleteHomework", mock.Anything, user.ID, hwID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	assert.NoError(t, h.DeleteHomework(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestPublishHomeworkHandler_Success(t *testing.T) {
	e := echo.New()
	h, svc := setupHomeworkHandler()
	user := testUser()
	hwID := uuid.New()
	isPublic := true
	hw := &model.Homework{HwID: hwID, IsPublic: &isPublic}
	svc.On("PublishHomework", mock.Anything, user.ID, hwID, true).Return(hw, nil)

	body, _ := json.Marshal(map[string]bool{"is_public": true})
	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("hwId")
	c.SetParamValues(hwID.String())

	assert.NoError(t, h.PublishHomework(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateDeadlineHandler_Success(t *testing.T) {
	e := echo.New()
	h, svc := setupHomeworkHandler()
	user := testUser()
	dlID := uuid.New()
	dl := &model.Deadline{ID: dlID, Title: "Updated"}
	svc.On("UpdateDeadline", mock.Anything, user.ID, dlID, mock.Anything).Return(dl, nil)

	body, _ := json.Marshal(map[string]string{"title": "Updated"})
	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("deadlineId")
	c.SetParamValues(dlID.String())

	assert.NoError(t, h.UpdateDeadline(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDeleteDeadlineHandler_Success(t *testing.T) {
	e := echo.New()
	h, svc := setupHomeworkHandler()
	user := testUser()
	dlID := uuid.New()
	svc.On("DeleteDeadline", mock.Anything, user.ID, dlID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("deadlineId")
	c.SetParamValues(dlID.String())

	assert.NoError(t, h.DeleteDeadline(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
