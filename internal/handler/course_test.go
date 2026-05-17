package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

// Моки
type mockCourseRepo struct {
	mock.Mock
}

func (m *mockCourseRepo) GetCoursesByUserID(ctx context.Context, userID uuid.UUID, status string) ([]model.Course, error) {
	args := m.Called(ctx, userID, status)
	return args.Get(0).([]model.Course), args.Error(1)
}

func (m *mockCourseRepo) GetCourseByID(ctx context.Context, courseID string) (*model.Course, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

func (m *mockCourseRepo) CreateCourse(ctx context.Context, course model.Course) (*model.Course, error) {
	args := m.Called(ctx, course)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

func (m *mockCourseRepo) UpdateCourse(ctx context.Context, courseID string, course model.Course) (*model.Course, error) {
	args := m.Called(ctx, courseID, course)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

func (m *mockCourseRepo) DeleteCourse(ctx context.Context, courseID string) error {
	return nil
}
func (m *mockCourseRepo) GetCourseBoard(ctx context.Context, courseID string) (*model.TaskBoardSummary, bool, error) {
	return nil, false, nil
}
func (m *mockCourseRepo) GetCourses(ctx context.Context) ([]model.Course, error) {
	return nil, nil
}

type mockRoleRepo struct {
	mock.Mock
}

func (m *mockRoleRepo) AssignRole(ctx context.Context, role *model.UserRole) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *mockRoleRepo) RevokeRole(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, courseID, roleID)
	return args.Error(0)
}

func (m *mockRoleRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.UserRole, error) {
	args := m.Called(ctx, courseID)
	return args.Get(0).([]model.UserRole), args.Error(1)
}

func (m *mockRoleRepo) GetRoleIDByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID, courseID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockRoleRepo) HasPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, roleID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *mockRoleRepo) AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *mockRoleRepo) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	args := m.Called(ctx, roleID, permission)
	return args.Error(0)
}

func (m *mockRoleRepo) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]model.CourseAdminPermission), args.Error(1)
}
func newTestUser() *model.User {
	return &model.User{ID: uuid.New(), Email: "test@example.com"}
}

func setupTest() (*CourseHandler, *echo.Echo, *mockCourseRepo, *mockRoleRepo) {
	e := echo.New()
	courseRepo := new(mockCourseRepo)
	roleRepo := new(mockRoleRepo)
	svc := service.NewCourseService(courseRepo, roleRepo)
	handler := NewCourseHandler(svc)
	return handler, e, courseRepo, roleRepo
}

func makeContext(e *echo.Echo, user *model.User, params map[string]string) echo.Context {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(UserContextKey, user)
	for k, v := range params {
		c.SetParamNames(k)
		c.SetParamValues(v)
	}
	return c
}

func jsonContext(e *echo.Echo, method string, user *model.User, body []byte) echo.Context {
	req := httptest.NewRequest(method, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(UserContextKey, user)
	return c
}

// Тесты

func TestGetCourses_Unauthorized(t *testing.T) {
	handler, e, _, _ := setupTest()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetCourses(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetCourses_Success(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()
	courses := []model.Course{{ID: uuid.New(), Name: "Go", Slug: "go"}}

	courseRepo.On("GetCoursesByUserID", mock.Anything, user.ID, "").Return(courses, nil)

	c := makeContext(e, user, nil)
	err := handler.GetCourses(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}

func TestGetCourse_Public(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()
	course := &model.Course{ID: uuid.New(), Name: "Pub", Type: model.CourseTypePublic}

	courseRepo.On("GetCourseByID", mock.Anything, "pub").Return(course, nil)

	c := makeContext(e, user, map[string]string{"courseId": "pub"})
	err := handler.GetCourse(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}

func TestGetCourse_Private_NoPermission(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()
	course := &model.Course{ID: uuid.New(), Name: "Priv", Type: model.CourseTypePrivate}

	courseRepo.On("GetCourseByID", mock.Anything, "priv").Return(course, nil)

	c := makeContext(e, user, map[string]string{"courseId": "priv"})
	err := handler.GetCourse(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, c.Response().Status)
}

func TestCreateCourse_Success(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()
	input := model.Course{
		ID: uuid.New(), Name: "New", Slug: "new", Status: "created",
		Type: model.CourseTypePrivate, URL: "/course/new",
	}

	courseRepo.On("GetCourseByID", mock.Anything, "new").Return(nil, errors.New("not found"))
	courseRepo.On("CreateCourse", mock.Anything, mock.Anything).Return(&input, nil)

	body := `{"name":"New","slug":"new","status":"created","startDate":"2026-01-01","endDate":"2026-02-01","repoTemplate":"git@t","description":"d"}`
	c := jsonContext(e, http.MethodPost, user, []byte(body))

	err := handler.CreateCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, c.Response().Status)
}

func TestCreateCourse_Conflict(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()
	existing := &model.Course{ID: uuid.New(), Slug: "dup"}

	courseRepo.On("GetCourseByID", mock.Anything, "dup").Return(existing, nil)

	body := `{"name":"Dup","slug":"dup","status":"created","startDate":"2026-01-01","endDate":"2026-02-01","repoTemplate":"git@t","description":"d"}`
	c := jsonContext(e, http.MethodPost, user, []byte(body))

	err := handler.CreateCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, c.Response().Status)
}

func TestCreateCourse_ValidationError(t *testing.T) {
	handler, e, _, _ := setupTest()
	user := newTestUser()

	body := `{"slug":"a"}`
	c := jsonContext(e, http.MethodPost, user, []byte(body))

	err := handler.CreateCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)
}

func TestUpdateCourse_Success(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()
	course := &model.Course{
		ID: uuid.New(), Name: "Old", Slug: "old", Status: "created",
		Type: model.CourseTypePrivate, StartDate: parseDate("2026-01-01"), EndDate: parseDate("2026-02-01"),
	}

	courseRepo.On("GetCourseByID", mock.Anything, "old").Return(course, nil)
	courseRepo.On("UpdateCourse", mock.Anything, "old", mock.Anything).Return(course, nil)

	body := `{"name":"Updated"}`
	c := makeContext(e, user, map[string]string{"courseId": "old"})
	c = jsonContextWithParams(e, http.MethodPut, user, []byte(body), "old")

	err := handler.UpdateCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}

func TestUpdateCourse_NotFound(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()

	courseRepo.On("GetCourseByID", mock.Anything, "unknown").Return(nil, errors.New("not found"))

	c := jsonContextWithParams(e, http.MethodPut, user, []byte(`{"name":"x"}`), "unknown")
	err := handler.UpdateCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)
}

func TestJoinCourse_Public(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	courseID := uuid.New()
	course := &model.Course{ID: courseID, Type: model.CourseTypePublic}

	courseRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, courseID).Return(uuid.Nil, errors.New("not found"))

	c := jsonContextWithParams(e, http.MethodPost, user, []byte(`{}`), courseID.String())
	err := handler.JoinCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}

func TestJoinCourse_Private_InvalidCode(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	courseID := uuid.New()
	code := "secret"
	course := &model.Course{ID: courseID, Type: model.CourseTypePrivate, InviteCode: &code}

	courseRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, courseID).Return(uuid.Nil, errors.New("not found"))

	c := jsonContextWithParams(e, http.MethodPost, user, []byte(`{"code":"wrong"}`), courseID.String())
	err := handler.JoinCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, c.Response().Status)
}

func TestJoinCourse_Unauthorized(t *testing.T) {
	handler, e, _, _ := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"code":"x"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.JoinCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetCourseBoard_Public(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()
	course := &model.Course{ID: uuid.New(), Name: "Pub", Type: model.CourseTypePublic}

	courseRepo.On("GetCourseByID", mock.Anything, "pub").Return(course, nil)
	courseRepo.On("GetCourseBoard", mock.Anything, "pub").Return(nil, false, nil)

	c := makeContext(e, user, map[string]string{"courseId": "pub"})
	err := handler.GetCourseBoard(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}

func parseDate(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}

func jsonContextWithParams(e *echo.Echo, method string, user *model.User, body []byte, param string) echo.Context {
	req := httptest.NewRequest(method, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(UserContextKey, user)
	c.SetParamNames("courseId")
	c.SetParamValues(param)
	return c
}
