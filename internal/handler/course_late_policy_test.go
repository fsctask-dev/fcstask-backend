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

type mockCLPRepo struct{ mock.Mock }

func (m *mockCLPRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) (*model.CourseLatePolicy, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CourseLatePolicy), args.Error(1)
}
func (m *mockCLPRepo) Create(ctx context.Context, p *model.CourseLatePolicy) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockCLPRepo) Update(ctx context.Context, p *model.CourseLatePolicy) error {
	return m.Called(ctx, p).Error(0)
}

type mockRoleRepoH struct{ mock.Mock }

func (m *mockRoleRepoH) AssignRoleWithPermissions(ctx context.Context, userRole *model.UserRole, permissions []string) error {
	return m.Called(ctx, userRole, permissions).Error(0)
}
func (m *mockRoleRepoH) RevokeRoleWithPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	return m.Called(ctx, userID, courseID, roleID).Error(0)
}
func (m *mockRoleRepoH) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.UserRole, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.UserRole), args.Error(1)
}
func (m *mockRoleRepoH) GetRoleIDByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID, courseID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}
func (m *mockRoleRepoH) RoleBelongsToCourse(ctx context.Context, roleID, courseID uuid.UUID) (bool, error) {
	args := m.Called(ctx, roleID, courseID)
	return args.Bool(0), args.Error(1)
}
func (m *mockRoleRepoH) HasPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, roleID, permission)
	return args.Bool(0), args.Error(1)
}
func (m *mockRoleRepoH) AddPermission(ctx context.Context, p *model.CourseAdminPermission) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockRoleRepoH) AddPermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	return m.Called(ctx, roleID, permissions).Error(0)
}
func (m *mockRoleRepoH) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	return m.Called(ctx, roleID, permission).Error(0)
}
func (m *mockRoleRepoH) RemovePermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	return m.Called(ctx, roleID, permissions).Error(0)
}
func (m *mockRoleRepoH) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.CourseAdminPermission), args.Error(1)
}

func setupCourseLateHandler() (*handler.CourseLateHandler, *mockCLPRepo) {
	clpRepo := new(mockCLPRepo)
	roleRepo := new(mockRoleRepoH)
	roleID := uuid.New()
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, mock.Anything).Return(true, nil)
	svc := service.NewCourseLatePolicy(clpRepo, roleRepo)
	return handler.NewCourseLateHandler(svc), clpRepo
}

func TestCourseLateHandler_CreateOrUpdate_Linear_Success(t *testing.T) {
	e := echo.New()
	h, clpRepo := setupCourseLateHandler()
	courseID := uuid.New()
	user := &model.User{ID: uuid.New()}

	clpRepo.On("GetByCourseID", mock.Anything, courseID).Return(nil, nil)
	clpRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeLinear
	})).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{
		"policy_type":         "linear",
		"soft_penalty":        0.0,
		"hard_deadline_score": 0.5,
	})
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	assert.NoError(t, h.CreateOrUpdate(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCourseLateHandler_CreateOrUpdate_NoUser(t *testing.T) {
	e := echo.New()
	h, _ := setupCourseLateHandler()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(uuid.New().String())
	assert.NoError(t, h.CreateOrUpdate(c))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCourseLateHandler_CreateOrUpdate_InvalidCourseID(t *testing.T) {
	e := echo.New()
	h, _ := setupCourseLateHandler()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})
	c.SetParamNames("courseId")
	c.SetParamValues("bad-uuid")
	assert.NoError(t, h.CreateOrUpdate(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCourseLateHandler_CreateOrUpdate_InvalidPolicyType(t *testing.T) {
	e := echo.New()
	h, _ := setupCourseLateHandler()
	courseID := uuid.New()
	body, _ := json.Marshal(map[string]interface{}{"policy_type": "invalid"})
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, &model.User{ID: uuid.New()})
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())
	assert.NoError(t, h.CreateOrUpdate(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCourseLateHandler_CreateOrUpdate_Step_Success(t *testing.T) {
	e := echo.New()
	h, clpRepo := setupCourseLateHandler()
	courseID := uuid.New()
	user := &model.User{ID: uuid.New()}

	clpRepo.On("GetByCourseID", mock.Anything, courseID).Return(nil, nil)
	clpRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeStep
	})).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{"policy_type": "step", "step_percent": 0.1})
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	assert.NoError(t, h.CreateOrUpdate(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCourseLateHandler_CreateOrUpdate_Coefficient_Success(t *testing.T) {
	e := echo.New()
	h, clpRepo := setupCourseLateHandler()
	courseID := uuid.New()
	user := &model.User{ID: uuid.New()}

	clpRepo.On("GetByCourseID", mock.Anything, courseID).Return(nil, nil)
	clpRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeCoefficient
	})).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{"policy_type": "coefficient", "coefficient": 0.8})
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	assert.NoError(t, h.CreateOrUpdate(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCourseLateHandler_CreateOrUpdate_Update_Existing(t *testing.T) {
	e := echo.New()
	h, clpRepo := setupCourseLateHandler()
	courseID := uuid.New()
	user := &model.User{ID: uuid.New()}

	existing := &model.CourseLatePolicy{
		ID: uuid.New(), CourseID: courseID,
		PolicyType: model.PolicyTypeLinear, HardDeadlineScore: 0.3,
	}
	clpRepo.On("GetByCourseID", mock.Anything, courseID).Return(existing, nil)
	clpRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.HardDeadlineScore == 0.7
	})).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{
		"policy_type":         "linear",
		"hard_deadline_score": 0.7,
	})
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(handler.UserContextKey, user)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	assert.NoError(t, h.CreateOrUpdate(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}
