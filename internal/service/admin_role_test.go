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

var testRoleID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func (m *MockRoleRepo) AssignRoleWithPermissions(ctx context.Context, userRole *model.UserRole, permissions []string) error {
	args := m.Called(ctx, userRole, permissions)
	return args.Error(0)
}

func (m *MockRoleRepo) RevokeRoleWithPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
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

func (m *MockRoleRepo) GetRoleIDByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID, courseID)
	if args.Get(0) == nil {
		return uuid.Nil, args.Error(1)
	}
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockRoleRepo) RoleBelongsToCourse(ctx context.Context, roleID, courseID uuid.UUID) (bool, error) {
	args := m.Called(ctx, roleID, courseID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepo) HasPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, roleID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepo) AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockRoleRepo) AddPermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	args := m.Called(ctx, roleID, permissions)
	return args.Error(0)
}

func (m *MockRoleRepo) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	args := m.Called(ctx, roleID, permission)
	return args.Error(0)
}

func (m *MockRoleRepo) RemovePermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	args := m.Called(ctx, roleID, permissions)
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
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).Return(testRoleID, nil)
	roleRepo.On("HasPermission", mock.Anything, testRoleID, mock.Anything).Return(true, nil)
	svc := service.NewAdminRoleService(roleRepo, userRepo)
	return svc, roleRepo, userRepo
}

func TestAssignCourseAdmin_Success(t *testing.T) {
	svc, roleRepo, userRepo := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	courseID := uuid.New()

	user := &model.User{ID: userID}
	userRepo.On("GetUserByID", ctx, userID).Return(user, nil)
	roleRepo.On("AddPermissions", ctx, mock.AnythingOfType("uuid.UUID"), service.CourseAdminPermissions()).Return(nil)

	result, err := svc.AssignCourseAdmin(ctx, uuid.New(), service.AssignCourseAdminInput{
		UserID:   userID,
		CourseID: courseID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, courseID, result.CourseID)
	assert.NotEqual(t, uuid.Nil, result.RoleID)
	userRepo.AssertExpectations(t)
	roleRepo.AssertExpectations(t)
}

func TestAssignCourseAdmin_MissingUserID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.AssignCourseAdmin(ctx, uuid.New(), service.AssignCourseAdminInput{
		UserID:   uuid.Nil,
		CourseID: uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestAssignCourseAdmin_MissingCourseID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.AssignCourseAdmin(ctx, uuid.New(), service.AssignCourseAdminInput{
		UserID:   uuid.New(),
		CourseID: uuid.Nil,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "course_id is required")
}

func TestAssignCourseAdmin_UserNotFound(t *testing.T) {
	svc, _, userRepo := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	userRepo.On("GetUserByID", ctx, userID).Return(nil, assert.AnError)

	result, err := svc.AssignCourseAdmin(ctx, uuid.New(), service.AssignCourseAdminInput{
		UserID:   userID,
		CourseID: uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "User not found")
	userRepo.AssertExpectations(t)
}

func TestAssignCourseAdmin_NotCourseParticipant(t *testing.T) {
	roleRepo := new(MockRoleRepo)
	userRepo := new(MockUserRepo)
	svc := service.NewAdminRoleService(roleRepo, userRepo)
	ctx := context.Background()

	actorID := uuid.New()
	actorRoleID := uuid.New()
	userID := uuid.New()
	courseID := uuid.New()

	roleRepo.On("GetRoleIDByUserAndCourse", ctx, actorID, courseID).Return(actorRoleID, nil)
	roleRepo.On("HasPermission", ctx, actorRoleID, service.PermissionCourseRoleAssign).Return(true, nil)
	user := &model.User{ID: userID}
	userRepo.On("GetUserByID", ctx, userID).Return(user, nil)
	roleRepo.On("GetRoleIDByUserAndCourse", ctx, userID, courseID).Return(uuid.Nil, assert.AnError)

	result, err := svc.AssignCourseAdmin(ctx, actorID, service.AssignCourseAdminInput{
		UserID:   userID,
		CourseID: courseID,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "User is not a course participant")
	userRepo.AssertExpectations(t)
	roleRepo.AssertExpectations(t)
}

func TestRevokeRole_Success(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	courseID := uuid.New()
	roleID := testRoleID

	roleRepo.On("GetRoleIDByUserAndCourse", ctx, userID, courseID).Return(roleID, nil)
	roleRepo.On("RemovePermissions", ctx, roleID, service.CourseAdminPermissions()).Return(nil)

	err := svc.RevokeCourseAdmin(ctx, uuid.New(), service.RevokeCourseAdminInput{
		UserID:   userID,
		CourseID: courseID,
	})

	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

func TestRevokeRole_MissingUserID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	err := svc.RevokeCourseAdmin(ctx, uuid.New(), service.RevokeCourseAdminInput{
		UserID:   uuid.Nil,
		CourseID: uuid.New(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestRevokeRole_MissingCourseID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	err := svc.RevokeCourseAdmin(ctx, uuid.New(), service.RevokeCourseAdminInput{
		UserID:   uuid.New(),
		CourseID: uuid.Nil,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "course_id is required")
}

func TestRevokeRole_RepoError(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	courseID := uuid.New()
	roleID := testRoleID

	roleRepo.On("GetRoleIDByUserAndCourse", ctx, userID, courseID).Return(roleID, nil)
	roleRepo.On("RemovePermissions", ctx, roleID, service.CourseAdminPermissions()).Return(assert.AnError)

	err := svc.RevokeCourseAdmin(ctx, uuid.New(), service.RevokeCourseAdminInput{
		UserID:   userID,
		CourseID: courseID,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to remove permissions")
	roleRepo.AssertExpectations(t)
}

func TestRemoveCourseParticipant_Success(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	userID := uuid.New()
	courseID := uuid.New()
	roleID := testRoleID

	roleRepo.On("GetRoleIDByUserAndCourse", ctx, userID, courseID).Return(roleID, nil)
	roleRepo.On("RevokeRoleWithPermissions", ctx, userID, courseID, roleID).Return(nil)

	err := svc.RemoveCourseParticipant(ctx, uuid.New(), service.RemoveCourseParticipantInput{
		UserID:   userID,
		CourseID: courseID,
	})

	assert.NoError(t, err)
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

	result, err := svc.ListUserRoles(ctx, uuid.New(), courseID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedRoles, result)
	roleRepo.AssertExpectations(t)
}

func TestListUserRoles_EmptyCourseID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.ListUserRoles(ctx, uuid.New(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "course_id is required")
}

func TestListUserRoles_RepoError(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	courseID := uuid.New()
	roleRepo.On("GetByCourseID", ctx, courseID).Return(nil, assert.AnError)

	result, err := svc.ListUserRoles(ctx, uuid.New(), courseID)

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
	courseID := uuid.New()

	roleRepo.On("AddPermission", ctx, mock.MatchedBy(func(p *model.CourseAdminPermission) bool {
		return p.RoleID == roleID && p.Permission == permission
	})).Return(nil)
	roleRepo.On("RoleBelongsToCourse", ctx, roleID, courseID).Return(true, nil)

	result, err := svc.AddPermission(ctx, uuid.New(), service.AddPermissionInput{
		CourseID:   courseID,
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

	result, err := svc.AddPermission(ctx, uuid.New(), service.AddPermissionInput{
		CourseID:   uuid.New(),
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

	result, err := svc.AddPermission(ctx, uuid.New(), service.AddPermissionInput{
		CourseID:   uuid.New(),
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
	courseID := uuid.New()
	roleRepo.On("AddPermission", ctx, mock.AnythingOfType("*model.CourseAdminPermission")).Return(assert.AnError)
	roleRepo.On("RoleBelongsToCourse", ctx, roleID, courseID).Return(true, nil)

	result, err := svc.AddPermission(ctx, uuid.New(), service.AddPermissionInput{
		CourseID:   courseID,
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
	courseID := uuid.New()

	roleRepo.On("RemovePermission", ctx, roleID, permission).Return(nil)
	roleRepo.On("RoleBelongsToCourse", ctx, roleID, courseID).Return(true, nil)

	err := svc.RemovePermission(ctx, uuid.New(), courseID, roleID, permission)

	assert.NoError(t, err)
	roleRepo.AssertExpectations(t)
}

func TestRemovePermission_EmptyRoleID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	err := svc.RemovePermission(ctx, uuid.New(), uuid.New(), uuid.Nil, "edit_course")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role_id is required")
}

func TestRemovePermission_EmptyPermission(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	err := svc.RemovePermission(ctx, uuid.New(), uuid.New(), uuid.New(), "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission is required")
}

func TestRemovePermission_RepoError(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	roleID := uuid.New()
	permission := "edit_course"
	courseID := uuid.New()
	roleRepo.On("RemovePermission", ctx, roleID, permission).Return(assert.AnError)
	roleRepo.On("RoleBelongsToCourse", ctx, roleID, courseID).Return(true, nil)

	err := svc.RemovePermission(ctx, uuid.New(), courseID, roleID, permission)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to remove permission")
	roleRepo.AssertExpectations(t)
}

func TestListPermissions_Success(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	roleID := uuid.New()
	courseID := uuid.New()
	expectedPerms := []model.CourseAdminPermission{
		{RoleID: roleID, Permission: "edit_course"},
		{RoleID: roleID, Permission: "delete_course"},
	}
	roleRepo.On("GetPermissions", ctx, roleID).Return(expectedPerms, nil)
	roleRepo.On("RoleBelongsToCourse", ctx, roleID, courseID).Return(true, nil)

	result, err := svc.ListPermissions(ctx, uuid.New(), courseID, roleID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedPerms, result)
	roleRepo.AssertExpectations(t)
}

func TestListPermissions_EmptyRoleID(t *testing.T) {
	svc, _, _ := setupRoleService()
	ctx := context.Background()

	result, err := svc.ListPermissions(ctx, uuid.New(), uuid.New(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "role_id is required")
}

func TestListPermissions_RepoError(t *testing.T) {
	svc, roleRepo, _ := setupRoleService()
	ctx := context.Background()

	roleID := uuid.New()
	courseID := uuid.New()
	roleRepo.On("GetPermissions", ctx, roleID).Return(nil, assert.AnError)
	roleRepo.On("RoleBelongsToCourse", ctx, roleID, courseID).Return(true, nil)

	result, err := svc.ListPermissions(ctx, uuid.New(), courseID, roleID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to fetch permissions")
	roleRepo.AssertExpectations(t)
}
