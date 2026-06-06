package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/handler"
	"fcstask-backend/internal/service"
)

type MockAdminRoleService struct {
	mock.Mock
}

func (m *MockAdminRoleService) CreateSuperAdmin(ctx context.Context, userID uuid.UUID, input service.CreateSuperAdminInput) (*model.UserRole, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserRole), args.Error(1)
}

func (m *MockAdminRoleService) GrantCourseCreatePermission(ctx context.Context, userID uuid.UUID, targetUserID uuid.UUID) (*model.UserRole, error) {
    args := m.Called(ctx, userID, targetUserID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.UserRole), args.Error(1)
}

func (m *MockAdminRoleService) RevokeCourseCreatePermission(ctx context.Context, userID uuid.UUID, targetUserID uuid.UUID) error {
    args := m.Called(ctx, userID, targetUserID)
    return args.Error(0)
}

func (m *MockAdminRoleService) AssignCourseAdmin(ctx context.Context, userID uuid.UUID, input service.AssignCourseAdminInput) (*model.UserRole, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserRole), args.Error(1)
}

func (m *MockAdminRoleService) RevokeCourseAdmin(ctx context.Context, userID uuid.UUID, input service.RevokeCourseAdminInput) error {
	args := m.Called(ctx, userID, input)
	return args.Error(0)
}

func (m *MockAdminRoleService) RemoveCourseParticipant(ctx context.Context, userID uuid.UUID, input service.RemoveCourseParticipantInput) error {
	args := m.Called(ctx, userID, input)
	return args.Error(0)
}

func (m *MockAdminRoleService) ListUserRoles(ctx context.Context, userID, courseID uuid.UUID) ([]model.UserRole, error) {
	args := m.Called(ctx, userID, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.UserRole), args.Error(1)
}

func (m *MockAdminRoleService) AddPermission(ctx context.Context, userID uuid.UUID, input service.AddPermissionInput) (*model.CourseAdminPermission, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CourseAdminPermission), args.Error(1)
}

func (m *MockAdminRoleService) RemovePermission(ctx context.Context, userID, courseID, roleID uuid.UUID, permission string) error {
	args := m.Called(ctx, userID, courseID, roleID, permission)
	return args.Error(0)
}

func (m *MockAdminRoleService) ListPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) ([]model.CourseAdminPermission, error) {
	args := m.Called(ctx, userID, courseID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.CourseAdminPermission), args.Error(1)
}

func TestHandlerAssignCourseAdmin_Success(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	userID := uuid.New()

	body := map[string]interface{}{
		"user_id": userID.String(),
	}
	roleID := uuid.New()
	expected := &model.UserRole{UserID: userID, CourseID: courseID, RoleID: roleID}

	c, rec := newEchoContext(http.MethodPost, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("AssignCourseAdmin", mock.Anything, mock.Anything, service.AssignCourseAdminInput{
		UserID:   userID,
		CourseID: courseID,
	}).Return(expected, nil)

	err := h.AssignCourseAdmin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerCreateSuperAdmin_Success(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	targetUserID := uuid.New()
	expected := &model.UserRole{UserID: targetUserID, CourseID: uuid.Nil, RoleID: uuid.New()}

	body := map[string]interface{}{
		"user_id": targetUserID.String(),
	}

	c, rec := newEchoContext(http.MethodPost, "/", body, nil)
	svc.On("CreateSuperAdmin", mock.Anything, mock.Anything, service.CreateSuperAdminInput{
		UserID: targetUserID,
	}).Return(expected, nil)

	err := h.CreateSuperAdmin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerAssignCourseAdmin_InvalidCourseID(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	c, rec := newEchoContext(http.MethodPost, "/", nil, map[string]string{"courseId": "bad"})

	err := h.AssignCourseAdmin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerAssignCourseAdmin_ServiceError_UserNotFound(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	userID := uuid.New()
	body := map[string]interface{}{
		"user_id": userID.String(),
	}

	c, rec := newEchoContext(http.MethodPost, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("AssignCourseAdmin", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.NotFound("User not found"))

	err := h.AssignCourseAdmin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerAssignCourseAdmin_ServiceError_BadRequest(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	body := map[string]interface{}{
		"user_id": uuid.Nil.String(),
	}

	c, rec := newEchoContext(http.MethodPost, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("AssignCourseAdmin", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.BadRequest("user_id is required"))

	err := h.AssignCourseAdmin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerRevokeRole_Success(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()

	body := map[string]interface{}{
		"user_id": userID.String(),
		"role_id": roleID.String(),
	}

	c, rec := newEchoContext(http.MethodDelete, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("RevokeCourseAdmin", mock.Anything, mock.Anything, service.RevokeCourseAdminInput{
		UserID:   userID,
		CourseID: courseID,
	}).Return(nil)

	err := h.RevokeCourseAdmin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerRevokeRole_InvalidCourseID(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"courseId": "bad"})

	err := h.RevokeCourseAdmin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerRevokeRole_ServiceError(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	body := map[string]interface{}{
		"user_id": uuid.New().String(),
	}

	c, rec := newEchoContext(http.MethodDelete, "/", body, map[string]string{"courseId": courseID.String()})
	svc.On("RevokeCourseAdmin", mock.Anything, mock.Anything, mock.Anything).Return(service.Internal("Failed to revoke role", nil))

	err := h.RevokeCourseAdmin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListUserRoles_Success(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	expected := []model.UserRole{
		{UserID: uuid.New(), CourseID: courseID, RoleID: uuid.New()},
	}

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"courseId": courseID.String()})
	svc.On("ListUserRoles", mock.Anything, mock.Anything, courseID).Return(expected, nil)

	err := h.ListUserRoles(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListUserRoles_InvalidCourseID(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"courseId": "bad"})

	err := h.ListUserRoles(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerListUserRoles_ServiceError(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"courseId": courseID.String()})
	svc.On("ListUserRoles", mock.Anything, mock.Anything, courseID).Return(nil, service.Internal("Failed to fetch roles", nil))

	err := h.ListUserRoles(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerAddPermission_Success(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	roleID := uuid.New()
	permission := "edit_course"
	body := map[string]interface{}{"permission": permission}
	expected := &model.CourseAdminPermission{RoleID: roleID, Permission: permission}

	c, rec := newEchoContextMultiParam(http.MethodPost, "/", body,
		[]string{"courseId", "roleId"},
		[]string{courseID.String(), roleID.String()},
	)
	svc.On("AddPermission", mock.Anything, mock.Anything, service.AddPermissionInput{
		CourseID:   courseID,
		RoleID:     roleID,
		Permission: permission,
	}).Return(expected, nil)

	err := h.AddPermission(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerAddPermission_InvalidRoleID(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	c, rec := newEchoContextMultiParam(http.MethodPost, "/", nil,
		[]string{"courseId", "roleId"},
		[]string{uuid.New().String(), "bad"},
	)

	err := h.AddPermission(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerAddPermission_ServiceError(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	roleID := uuid.New()
	body := map[string]interface{}{"permission": ""}

	c, rec := newEchoContextMultiParam(http.MethodPost, "/", body,
		[]string{"courseId", "roleId"},
		[]string{courseID.String(), roleID.String()},
	)
	svc.On("AddPermission", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.BadRequest("permission is required"))

	err := h.AddPermission(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerRemovePermission_Success(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	roleID := uuid.New()
	permission := "edit_course"

	c, rec := newEchoContextMultiParam(http.MethodDelete, "/", nil,
		[]string{"courseId", "roleId", "permission"},
		[]string{courseID.String(), roleID.String(), permission},
	)
	svc.On("RemovePermission", mock.Anything, mock.Anything, courseID, roleID, permission).Return(nil)

	err := h.RemovePermission(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerRemovePermission_InvalidRoleID(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	c, rec := newEchoContextMultiParam(http.MethodDelete, "/", nil,
		[]string{"courseId", "roleId", "permission"},
		[]string{uuid.New().String(), "bad", "edit_course"},
	)

	err := h.RemovePermission(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerRemovePermission_ServiceError(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	roleID := uuid.New()
	permission := "edit_course"

	c, rec := newEchoContextMultiParam(http.MethodDelete, "/", nil,
		[]string{"courseId", "roleId", "permission"},
		[]string{courseID.String(), roleID.String(), permission},
	)
	svc.On("RemovePermission", mock.Anything, mock.Anything, courseID, roleID, permission).Return(service.Internal("Failed to remove permission", nil))

	err := h.RemovePermission(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListPermissions_Success(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	roleID := uuid.New()
	expected := []model.CourseAdminPermission{
		{RoleID: roleID, Permission: "edit_course"},
		{RoleID: roleID, Permission: "delete_course"},
	}

	c, rec := newEchoContextMultiParam(http.MethodGet, "/", nil,
		[]string{"courseId", "roleId"},
		[]string{courseID.String(), roleID.String()},
	)
	svc.On("ListPermissions", mock.Anything, mock.Anything, courseID, roleID).Return(expected, nil)

	err := h.ListPermissions(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListPermissions_InvalidRoleID(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	c, rec := newEchoContextMultiParam(http.MethodGet, "/", nil,
		[]string{"courseId", "roleId"},
		[]string{uuid.New().String(), "bad"},
	)

	err := h.ListPermissions(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerListPermissions_ServiceError(t *testing.T) {
	svc := new(MockAdminRoleService)
	h := handler.NewAdminRoleHandler(svc)

	courseID := uuid.New()
	roleID := uuid.New()
	c, rec := newEchoContextMultiParam(http.MethodGet, "/", nil,
		[]string{"courseId", "roleId"},
		[]string{courseID.String(), roleID.String()},
	)
	svc.On("ListPermissions", mock.Anything, mock.Anything, courseID, roleID).Return(nil, service.Internal("Failed to fetch permissions", nil))

	err := h.ListPermissions(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	svc.AssertExpectations(t)
}
