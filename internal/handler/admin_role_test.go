package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/handler"
)

type MockRoleRepo struct {
	mock.Mock
}

func (m *MockRoleRepo) AssignRole(ctx context.Context, userRole *model.UserRole) error {
	return m.Called(ctx, userRole).Error(0)
}

func (m *MockRoleRepo) RevokeRole(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	return m.Called(ctx, userID, courseID, roleID).Error(0)
}

func (m *MockRoleRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.UserRole, error) {
	args := m.Called(ctx, courseID)
	return args.Get(0).([]model.UserRole), args.Error(1)
}

func (m *MockRoleRepo) AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error {
	return m.Called(ctx, perm).Error(0)
}

func (m *MockRoleRepo) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	return m.Called(ctx, roleID, permission).Error(0)
}

func (m *MockRoleRepo) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]model.CourseAdminPermission), args.Error(1)
}

type MockUserRepoForRole struct {
	mock.Mock
}

func (m *MockUserRepoForRole) CreateUser(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepoForRole) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepoForRole) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepoForRole) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepoForRole) GetUserByUserID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepoForRole) GetUserByTgUID(ctx context.Context, tgUID int64) (*model.User, error) {
	args := m.Called(ctx, tgUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepoForRole) UpdateUser(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepoForRole) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockUserRepoForRole) GetUsersWithSessions(ctx context.Context, limit, offset int) ([]model.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepoForRole) CountUsersWithSessions(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepoForRole) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepoForRole) ExistsUserByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepoForRole) CountUsers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func TestAdminAssignRole_Success(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	courseID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()

	body := map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
	}
	req, rec := newRequest(http.MethodPost, "/admin/courses/"+courseID.String()+"/roles", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	user := &model.User{ID: userID}
	userRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	roleRepo.On("AssignRole", mock.Anything, mock.AnythingOfType("*model.UserRole")).Return(nil)

	err := h.AdminAssignRoleHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestAdminAssignRole_InvalidCourseID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	req, rec := newRequest(http.MethodPost, "/admin/courses/invalid-uuid/roles", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("invalid-uuid")

	err := h.AdminAssignRoleHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminAssignRole_MissingUserID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	courseID := uuid.New()
	roleID := uuid.New()

	body := map[string]interface{}{
		"role_id": roleID,
	}
	req, rec := newRequest(http.MethodPost, "/admin/courses/"+courseID.String()+"/roles", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := h.AdminAssignRoleHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminAssignRole_MissingRoleID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	courseID := uuid.New()
	userID := uuid.New()

	body := map[string]interface{}{
		"user_id": userID,
	}
	req, rec := newRequest(http.MethodPost, "/admin/courses/"+courseID.String()+"/roles", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := h.AdminAssignRoleHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminAssignRole_UserNotFound(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	courseID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()

	body := map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
	}
	req, rec := newRequest(http.MethodPost, "/admin/courses/"+courseID.String()+"/roles", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	userRepo.On("GetUserByID", mock.Anything, userID).Return(nil, assert.AnError)

	err := h.AdminAssignRoleHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminRevokeRole_Success(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	courseID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()

	body := map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
	}
	req, rec := newRequest(http.MethodDelete, "/admin/courses/"+courseID.String()+"/roles", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	roleRepo.On("RevokeRole", mock.Anything, userID, courseID, roleID).Return(nil)

	err := h.AdminRevokeRoleHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAdminRevokeRole_InvalidCourseID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	req, rec := newRequest(http.MethodDelete, "/admin/courses/invalid-uuid/roles", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("invalid-uuid")

	err := h.AdminRevokeRoleHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminRevokeRole_MissingUserID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	courseID := uuid.New()
	roleID := uuid.New()

	body := map[string]interface{}{
		"role_id": roleID,
	}
	req, rec := newRequest(http.MethodDelete, "/admin/courses/"+courseID.String()+"/roles", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := h.AdminRevokeRoleHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminRevokeRole_MissingRoleID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	courseID := uuid.New()
	userID := uuid.New()

	body := map[string]interface{}{
		"user_id": userID,
	}
	req, rec := newRequest(http.MethodDelete, "/admin/courses/"+courseID.String()+"/roles", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	err := h.AdminRevokeRoleHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminListUserRoles_Success(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	courseID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/courses/"+courseID.String()+"/roles", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(courseID.String())

	expectedRoles := []model.UserRole{
		{UserID: uuid.New(), CourseID: courseID, RoleID: uuid.New()},
		{UserID: uuid.New(), CourseID: courseID, RoleID: uuid.New()},
	}
	roleRepo.On("GetByCourseID", mock.Anything, courseID).Return(expectedRoles, nil)

	err := h.AdminListUserRolesHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result []model.UserRole
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Len(t, result, 2)
}

func TestAdminListUserRoles_InvalidCourseID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	req, rec := newRequest(http.MethodGet, "/admin/courses/invalid-uuid/roles", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("invalid-uuid")

	err := h.AdminListUserRolesHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminAddPermission_Success(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	roleID := uuid.New()
	body := map[string]interface{}{
		"permission": "edit_course",
	}
	req, rec := newRequest(http.MethodPost, "/admin/roles/"+roleID.String()+"/permissions", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("roleId")
	c.SetParamValues(roleID.String())

	roleRepo.On("AddPermission", mock.Anything, mock.AnythingOfType("*model.CourseAdminPermission")).Return(nil)

	err := h.AdminAddPermissionHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestAdminAddPermission_InvalidRoleID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	req, rec := newRequest(http.MethodPost, "/admin/roles/invalid-uuid/permissions", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("roleId")
	c.SetParamValues("invalid-uuid")

	err := h.AdminAddPermissionHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminAddPermission_MissingPermission(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	roleID := uuid.New()
	body := map[string]interface{}{}
	req, rec := newRequest(http.MethodPost, "/admin/roles/"+roleID.String()+"/permissions", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("roleId")
	c.SetParamValues(roleID.String())

	err := h.AdminAddPermissionHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminAddPermission_EmptyPermission(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	roleID := uuid.New()
	body := map[string]interface{}{
		"permission": "",
	}
	req, rec := newRequest(http.MethodPost, "/admin/roles/"+roleID.String()+"/permissions", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("roleId")
	c.SetParamValues(roleID.String())

	err := h.AdminAddPermissionHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminRemovePermission_Success(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	roleID := uuid.New()
	permission := "edit_course"
	req, rec := newRequest(http.MethodDelete, "/admin/roles/"+roleID.String()+"/permissions/"+permission, nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("roleId", "permission")
	c.SetParamValues(roleID.String(), permission)

	roleRepo.On("RemovePermission", mock.Anything, roleID, permission).Return(nil)

	err := h.AdminRemovePermissionHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAdminRemovePermission_InvalidRoleID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	req, rec := newRequest(http.MethodDelete, "/admin/roles/invalid-uuid/permissions/edit_course", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("roleId", "permission")
	c.SetParamValues("invalid-uuid", "edit_course")

	err := h.AdminRemovePermissionHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminRemovePermission_EmptyPermission(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	roleID := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/admin/roles/"+roleID.String()+"/permissions/", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("roleId", "permission")
	c.SetParamValues(roleID.String(), "")

	err := h.AdminRemovePermissionHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminListPermissions_Success(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	roleID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/roles/"+roleID.String()+"/permissions", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("roleId")
	c.SetParamValues(roleID.String())

	expectedPerms := []model.CourseAdminPermission{
		{RoleID: roleID, Permission: "edit_course"},
		{RoleID: roleID, Permission: "delete_course"},
	}
	roleRepo.On("GetPermissions", mock.Anything, roleID).Return(expectedPerms, nil)

	err := h.AdminListPermissionsHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result []model.CourseAdminPermission
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Len(t, result, 2)
}

func TestAdminListPermissions_InvalidRoleID(t *testing.T) {
	e := setupEcho()
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepoForRole)
	h := handler.NewAdminRoleHandler(roleRepo, userRepo)

	req, rec := newRequest(http.MethodGet, "/admin/roles/invalid-uuid/permissions", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("roleId")
	c.SetParamValues("invalid-uuid")

	err := h.AdminListPermissionsHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
