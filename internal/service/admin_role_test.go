package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type MockRoleRepo struct {
	mock.Mock
}

func (m *MockRoleRepo) AssignRole(ctx context.Context, userRole *model.UserRole) error {
	args := m.Called(ctx, userRole)
	return args.Error(0)
}

func (m *MockRoleRepo) RevokeRole(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, courseID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.UserRole, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.UserRole), args.Error(1)
}

func (m *MockRoleRepo) AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockRoleRepo) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	args := m.Called(ctx, roleID, permission)
	return args.Error(0)
}

func (m *MockRoleRepo) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.CourseAdminPermission), args.Error(1)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) CreateUser(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByUserID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByTgUID(ctx context.Context, tgUID int64) (*model.User, error) {
	args := m.Called(ctx, tgUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) UpdateUser(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) GetUsersWithSessions(ctx context.Context, limit, offset int) ([]model.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepo) CountUsersWithSessions(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepo) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepo) ExistsUserByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepo) CountUsers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func setupRoleService() (*service.AdminRoleService, *MockRoleRepo, *MockUserRepo) {
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepo)
	svc := service.NewAdminRoleService(roleRepo, userRepo)
	return svc, roleRepo, userRepo
}

func TestAssignRole_Success(t *testing.T) {
	svc, roleRepo, userRepo := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	courseID := uuid.New()
	roleID := uuid.New()

	user := &model.User{ID: userID}
	userRepo.On("GetUserByID", ctx, userID).Return(user, nil)
	roleRepo.On("AssignRole", ctx, mock.MatchedBy(func(ur *model.UserRole) bool {
		return ur.UserID == userID && ur.CourseID == courseID && ur.RoleID == roleID
	})).Return(nil)

	result, err := svc.AssignRole(ctx, service.AssignRoleInput{
		UserID:   userID,
		CourseID: courseID,
		RoleID:   roleID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, courseID, result.CourseID)
	assert.Equal(t, roleID, result.RoleID)
	userRepo.AssertExpectations(t)
	roleRepo.AssertExpectations(t)
}

func TestAssignRole_MissingUserID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.AssignRole(ctx, service.AssignRoleInput{
		UserID:   uuid.Nil,
		CourseID: uuid.New(),
		RoleID:   uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestAssignRole_MissingCourseID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.AssignRole(ctx, service.AssignRoleInput{
		UserID:   uuid.New(),
		CourseID: uuid.Nil,
		RoleID:   uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "course_id is required")
}

func TestAssignRole_MissingRoleID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.AssignRole(ctx, service.AssignRoleInput{
		UserID:   uuid.New(),
		CourseID: uuid.New(),
		RoleID:   uuid.Nil,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "role_id is required")
}

func TestAssignRole_UserNotFound(t *testing.T) {
	svc, _, userRepo := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	userRepo.On("GetUserByID", ctx, userID).Return(nil, assert.AnError)

	result, err := svc.AssignRole(ctx, service.AssignRoleInput{
		UserID:   userID,
		CourseID: uuid.New(),
		RoleID:   uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "User not found")
	userRepo.AssertExpectations(t)
}

func TestAssignRole_RepoError(t *testing.T) {
	svc, roleRepo, userRepo := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	courseID := uuid.New()
	roleID := uuid.New()

	user := &model.User{ID: userID}
	userRepo.On("GetUserByID", ctx, userID).Return(user, nil)
	roleRepo.On("AssignRole", ctx, mock.AnythingOfType("*model.UserRole")).Return(assert.AnError)

	result, err := svc.AssignRole(ctx, service.AssignRoleInput{
		UserID:   userID,
		CourseID: courseID,
		RoleID:   roleID,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to assign role")
	userRepo.AssertExpectations(t)
	roleRepo.AssertExpectations(t)
}

func TestRevokeRole_Success(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	courseID := uuid.New()
	roleID := uuid.New()

	roleRepo.On("RevokeRole", ctx, userID, courseID, roleID).Return(nil)

	err := svc.RevokeRole(ctx, service.RevokeRoleInput{
		UserID:   userID,
		CourseID: courseID,
		RoleID:   roleID,
	})

	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

func TestRevokeRole_MissingUserID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	err := svc.RevokeRole(ctx, service.RevokeRoleInput{
		UserID:   uuid.Nil,
		CourseID: uuid.New(),
		RoleID:   uuid.New(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestRevokeRole_MissingCourseID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	err := svc.RevokeRole(ctx, service.RevokeRoleInput{
		UserID:   uuid.New(),
		CourseID: uuid.Nil,
		RoleID:   uuid.New(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "course_id is required")
}

func TestRevokeRole_MissingRoleID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	err := svc.RevokeRole(ctx, service.RevokeRoleInput{
		UserID:   uuid.New(),
		CourseID: uuid.New(),
		RoleID:   uuid.Nil,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role_id is required")
}

func TestRevokeRole_RepoError(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	courseID := uuid.New()
	roleID := uuid.New()

	roleRepo.On("RevokeRole", ctx, userID, courseID, roleID).Return(assert.AnError)

	err := svc.RevokeRole(ctx, service.RevokeRoleInput{
		UserID:   userID,
		CourseID: courseID,
		RoleID:   roleID,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to revoke role")
	roleRepo.AssertExpectations(t)
}

func TestListUserRoles_Success(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	courseID := uuid.New()
	expectedRoles := []model.UserRole{
		{UserID: uuid.New(), CourseID: courseID, RoleID: uuid.New()},
		{UserID: uuid.New(), CourseID: courseID, RoleID: uuid.New()},
	}
	roleRepo.On("GetByCourseID", ctx, courseID).Return(expectedRoles, nil)

	result, err := svc.ListUserRoles(ctx, courseID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedRoles, result)
	roleRepo.AssertExpectations(t)
}

func TestListUserRoles_EmptyCourseID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.ListUserRoles(ctx, uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "course_id is required")
}

func TestListUserRoles_RepoError(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	courseID := uuid.New()
	roleRepo.On("GetByCourseID", ctx, courseID).Return(nil, assert.AnError)

	result, err := svc.ListUserRoles(ctx, courseID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to fetch roles")
	roleRepo.AssertExpectations(t)
}

func TestAddPermission_Success(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	roleID := uuid.New()
	permission := "edit_course"

	roleRepo.On("AddPermission", ctx, mock.MatchedBy(func(p *model.CourseAdminPermission) bool {
		return p.RoleID == roleID && p.Permission == permission
	})).Return(nil)

	result, err := svc.AddPermission(ctx, service.AddPermissionInput{
		RoleID:     roleID,
		Permission: permission,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, roleID, result.RoleID)
	assert.Equal(t, permission, result.Permission)
	roleRepo.AssertExpectations(t)
}

func TestAddPermission_MissingRoleID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.AddPermission(ctx, service.AddPermissionInput{
		RoleID:     uuid.Nil,
		Permission: "edit_course",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "role_id is required")
}

func TestAddPermission_EmptyPermission(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.AddPermission(ctx, service.AddPermissionInput{
		RoleID:     uuid.New(),
		Permission: "",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "permission is required")
}

func TestAddPermission_RepoError(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	roleID := uuid.New()
	roleRepo.On("AddPermission", ctx, mock.AnythingOfType("*model.CourseAdminPermission")).Return(assert.AnError)

	result, err := svc.AddPermission(ctx, service.AddPermissionInput{
		RoleID:     roleID,
		Permission: "edit_course",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to add permission")
	roleRepo.AssertExpectations(t)
}

func TestRemovePermission_Success(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	roleID := uuid.New()
	permission := "edit_course"

	roleRepo.On("RemovePermission", ctx, roleID, permission).Return(nil)

	err := svc.RemovePermission(ctx, roleID, permission)

	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

func TestRemovePermission_EmptyRoleID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	err := svc.RemovePermission(ctx, uuid.Nil, "edit_course")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role_id is required")
}

func TestRemovePermission_EmptyPermission(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	err := svc.RemovePermission(ctx, uuid.New(), "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission is required")
}

func TestRemovePermission_RepoError(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	roleID := uuid.New()
	permission := "edit_course"
	roleRepo.On("RemovePermission", ctx, roleID, permission).Return(assert.AnError)

	err := svc.RemovePermission(ctx, roleID, permission)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to remove permission")
	roleRepo.AssertExpectations(t)
}

func TestListPermissions_Success(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	roleID := uuid.New()
	expectedPerms := []model.CourseAdminPermission{
		{RoleID: roleID, Permission: "edit_course"},
		{RoleID: roleID, Permission: "delete_course"},
	}
	roleRepo.On("GetPermissions", ctx, roleID).Return(expectedPerms, nil)

	result, err := svc.ListPermissions(ctx, roleID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedPerms, result)
	roleRepo.AssertExpectations(t)
}

func TestListPermissions_EmptyRoleID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.ListPermissions(ctx, uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "role_id is required")
}

func TestListPermissions_RepoError(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	roleID := uuid.New()
	roleRepo.On("GetPermissions", ctx, roleID).Return(nil, assert.AnError)

	result, err := svc.ListPermissions(ctx, roleID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to fetch permissions")
	roleRepo.AssertExpectations(t)
}
