package handler

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
	"gorm.io/gorm"

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
func (m *mockCourseRepo) GetCourseBoard(ctx context.Context, courseID string, userID uuid.UUID) (*model.TaskBoardSummary, bool, error) {
	return nil, false, nil
}
func (m *mockCourseRepo) GetCourses(ctx context.Context) ([]model.Course, error) {
	return nil, nil
}
func (m *mockCourseRepo) GetLeaderboard(ctx context.Context, courseID uuid.UUID) ([]model.LeaderboardEntry, error) {
	args := m.Called(ctx, courseID)
	return args.Get(0).([]model.LeaderboardEntry), args.Error(1)
}

type mockRoleRepo struct {
	mock.Mock
}

func (m *mockRoleRepo) AssignRoleWithPermissions(ctx context.Context, role *model.UserRole, permissions []string) error {
	args := m.Called(ctx, role, permissions)
	return args.Error(0)
}

func (m *mockRoleRepo) RevokeRoleWithPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
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

func (m *mockRoleRepo) RoleBelongsToCourse(ctx context.Context, roleID, courseID uuid.UUID) (bool, error) {
	args := m.Called(ctx, roleID, courseID)
	return args.Bool(0), args.Error(1)
}

func (m *mockRoleRepo) HasPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, roleID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *mockRoleRepo) AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *mockRoleRepo) AddPermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	args := m.Called(ctx, roleID, permissions)
	return args.Error(0)
}

func (m *mockRoleRepo) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	args := m.Called(ctx, roleID, permission)
	return args.Error(0)
}

func (m *mockRoleRepo) RemovePermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	args := m.Called(ctx, roleID, permissions)
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
	svc := service.NewCourseService(courseRepo, roleRepo, nil)
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

func TestGetCourse_Public_NoAuth(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	course := &model.Course{ID: uuid.New(), Name: "Pub", Type: model.CourseTypePublic}

	courseRepo.On("GetCourseByID", mock.Anything, "pub").Return(course, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("pub")

	err := handler.GetCourse(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCourse_Public_WithAuth(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()
	course := &model.Course{ID: uuid.New(), Name: "Pub", Type: model.CourseTypePublic}

	courseRepo.On("GetCourseByID", mock.Anything, "pub").Return(course, nil)

	c := makeContext(e, user, map[string]string{"courseId": "pub"})
	err := handler.GetCourse(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}

func TestGetCourse_Private_NoAuth(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	course := &model.Course{ID: uuid.New(), Name: "Priv", Type: model.CourseTypePrivate}

	courseRepo.On("GetCourseByID", mock.Anything, "priv").Return(course, nil)

	// Без пользователя
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("priv")

	err := handler.GetCourse(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestGetCourse_Private_HasPermission(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	roleID := uuid.New()
	course := &model.Course{ID: uuid.New(), Name: "Priv", Type: model.CourseTypePrivate}

	courseRepo.On("GetCourseByID", mock.Anything, "priv").Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, course.ID).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionCourseRead).Return(true, nil)

	c := makeContext(e, user, map[string]string{"courseId": "priv"})
	err := handler.GetCourse(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}

func TestGetCourse_Private_NoPermission(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	course := &model.Course{ID: uuid.New(), Name: "Priv", Type: model.CourseTypePrivate}

	courseRepo.On("GetCourseByID", mock.Anything, "priv").Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, course.ID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	c := makeContext(e, user, map[string]string{"courseId": "priv"})
	err := handler.GetCourse(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, c.Response().Status)
}

func TestCreateCourse_Success(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	roleID := uuid.New()
	input := model.Course{
		ID: uuid.New(), Name: "New", Slug: "new", Status: "created",
		Type: model.CourseTypePrivate, URL: "/course/new",
	}

	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, uuid.Nil).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionCourseCreate).Return(true, nil)
	courseRepo.On("GetCourseByID", mock.Anything, "new").Return(nil, nil)
	courseRepo.On("CreateCourse", mock.Anything, mock.Anything).Return(&input, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, input.ID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	roleRepo.On("AssignRoleWithPermissions", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	body := `{"name":"New","slug":"new","status":"created","startDate":"2026-01-01","endDate":"2026-02-01","repoTemplate":"git@t","description":"d"}`
	c := jsonContext(e, http.MethodPost, user, []byte(body))

	err := handler.CreateCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, c.Response().Status)
}

func TestCreateCourse_Conflict(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	roleID := uuid.New()
	existing := &model.Course{ID: uuid.New(), Slug: "dup"}

	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, uuid.Nil).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionCourseCreate).Return(true, nil)
	courseRepo.On("GetCourseByID", mock.Anything, "dup").Return(existing, nil)

	body := `{"name":"Dup","slug":"dup","status":"created","startDate":"2026-01-01","endDate":"2026-02-01","repoTemplate":"git@t","description":"d"}`
	c := jsonContext(e, http.MethodPost, user, []byte(body))

	err := handler.CreateCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, c.Response().Status)
}

func TestCreateCourse_ValidationError(t *testing.T) {
	handler, e, _, roleRepo := setupTest()
	user := newTestUser()
	roleID := uuid.New()

	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, uuid.Nil).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionCourseCreate).Return(true, nil)

	body := `{"slug":"a"}`
	c := jsonContext(e, http.MethodPost, user, []byte(body))

	err := handler.CreateCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)
}

func TestUpdateCourse_Success(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	roleID := uuid.New()
	course := &model.Course{
		ID: uuid.New(), Name: "Old", Slug: "old", Status: "created",
		Type: model.CourseTypePrivate, StartDate: parseDate("2026-01-01"), EndDate: parseDate("2026-02-01"),
	}

	courseRepo.On("GetCourseByID", mock.Anything, "old").Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, course.ID).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionCourseRead).Return(true, nil)   // ← GetCourse
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionCourseUpdate).Return(true, nil) // ← UpdateCourse
	courseRepo.On("UpdateCourse", mock.Anything, "old", mock.Anything).Return(course, nil)

	body := `{"name":"Updated"}`
	c := jsonContextWithParams(e, http.MethodPut, user, []byte(body), "old")

	err := handler.UpdateCourse(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}

func TestUpdateCourse_NotFound(t *testing.T) {
	handler, e, courseRepo, _ := setupTest()
	user := newTestUser()

	courseRepo.On("GetCourseByID", mock.Anything, "unknown").Return(nil, nil)

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
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, courseID).Return(uuid.Nil, gorm.ErrRecordNotFound).Twice()
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)
	roleRepo.On("AssignRoleWithPermissions", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, courseID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

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
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	roleID := uuid.New()
	course := &model.Course{ID: uuid.New(), Name: "Pub", Type: model.CourseTypePublic}

	courseRepo.On("GetCourseByID", mock.Anything, "pub").Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, course.ID).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionCourseRead).Return(true, nil)
	courseRepo.On("GetCourseBoard", mock.Anything, "pub", user.ID).Return(nil, false, nil)


	c := makeContext(e, user, map[string]string{"courseId": "pub"})
	err := handler.GetCourseBoard(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}
func TestGetScores_Success(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	roleID := uuid.New()
	courseID := uuid.New()
	course := &model.Course{ID: courseID, Name: "Go", Type: model.CourseTypePrivate}

	entries := []model.LeaderboardEntry{
		{Username: "alice", TotalScore: 30, Tasks: map[uuid.UUID]int{uuid.New(): 10, uuid.New(): 20}, Rank: 1},
		{Username: "bob", TotalScore: 20, Tasks: map[uuid.UUID]int{uuid.New(): 20}, Rank: 2},
	}

	courseRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, courseID).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionCourseRead).Return(true, nil)      // GetCourse
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionLeaderboardRead).Return(true, nil) // GetLeaderboard
	courseRepo.On("GetLeaderboard", mock.Anything, courseID).Return(entries, nil)

	c := makeContext(e, user, map[string]string{"courseId": courseID.String()})
	err := handler.GetScores(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var resp []model.LeaderboardEntry
	json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, "alice", resp[0].Username)
	assert.Equal(t, 30, resp[0].TotalScore)
	assert.Equal(t, 1, resp[0].Rank)
}

func TestGetScores_Forbidden(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	courseID := uuid.New()
	course := &model.Course{ID: courseID, Type: model.CourseTypePrivate}

	courseRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, courseID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	c := makeContext(e, user, map[string]string{"courseId": courseID.String()})
	err := handler.GetScores(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, c.Response().Status)
}

func TestGetScores_Empty(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	roleID := uuid.New()
	courseID := uuid.New()
	course := &model.Course{ID: courseID, Type: model.CourseTypePrivate}

	courseRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, courseID).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionCourseRead).Return(true, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionLeaderboardRead).Return(true, nil)
	courseRepo.On("GetLeaderboard", mock.Anything, courseID).Return([]model.LeaderboardEntry{}, nil)

	c := makeContext(e, user, map[string]string{"courseId": courseID.String()})
	err := handler.GetScores(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	var resp []model.LeaderboardEntry
	json.Unmarshal(c.Response().Writer.(*httptest.ResponseRecorder).Body.Bytes(), &resp)
	assert.Empty(t, resp)
}

func TestGetCourseBoard_Private_Forbidden(t *testing.T) {
	handler, e, courseRepo, roleRepo := setupTest()
	user := newTestUser()
	course := &model.Course{ID: uuid.New(), Name: "Priv", Type: model.CourseTypePrivate}

	courseRepo.On("GetCourseByID", mock.Anything, "priv").Return(course, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, course.ID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, user.ID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	c := makeContext(e, user, map[string]string{"courseId": "priv"})
	err := handler.GetCourseBoard(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, c.Response().Status)
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
