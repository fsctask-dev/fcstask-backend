// internal/service/stats_service_test.go
package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	models "fcstask-backend/internal/db/model"
)

type MockStatsRepo struct {
	mock.Mock
}

func (m *MockStatsRepo) GetStats(ctx context.Context) (*models.PlatformStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlatformStats), args.Error(1)
}

type MockRoleRepoForStats struct {
	mock.Mock
}

func (m *MockRoleRepoForStats) GetRoleIDByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID, courseID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockRoleRepoForStats) HasPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, roleID, permission)
	return args.Bool(0), args.Error(1)
}

// Остальные методы IRoleRepo — заглушки
func (m *MockRoleRepoForStats) AssignRoleWithPermissions(ctx context.Context, role *models.UserRole, permissions []string) error {
	return nil
}
func (m *MockRoleRepoForStats) RevokeRoleWithPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	return nil
}
func (m *MockRoleRepoForStats) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]models.UserRole, error) {
	return nil, nil
}
func (m *MockRoleRepoForStats) RoleBelongsToCourse(ctx context.Context, roleID, courseID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockRoleRepoForStats) AddPermission(ctx context.Context, perm *models.CourseAdminPermission) error {
	return nil
}
func (m *MockRoleRepoForStats) AddPermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	return nil
}
func (m *MockRoleRepoForStats) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	return nil
}
func (m *MockRoleRepoForStats) RemovePermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	return nil
}
func (m *MockRoleRepoForStats) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]models.CourseAdminPermission, error) {
	return nil, nil
}

func setupStatsService() (*StatsService, *MockStatsRepo, *MockRoleRepoForStats) {
	statsRepo := new(MockStatsRepo)
	roleRepo := new(MockRoleRepoForStats)
	svc := NewStatsService(statsRepo, roleRepo)
	return svc, statsRepo, roleRepo
}

func TestGetStats_Success(t *testing.T) {
	svc, statsRepo, roleRepo := setupStatsService()
	userID := uuid.New()
	roleID := uuid.New()

	expected := &models.PlatformStats{
		TotalCourses:   10,
		PublicCourses:  6,
		PrivateCourses: 4,
		TotalUsers:     42,
	}

	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, PermissionStatsRead).Return(true, nil)
	statsRepo.On("GetStats", mock.Anything).Return(expected, nil)

	result, err := svc.GetStats(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, int64(10), result.TotalCourses)
	assert.Equal(t, int64(42), result.TotalUsers)
}

func TestGetStats_Forbidden(t *testing.T) {
	svc, _, roleRepo := setupStatsService()
	userID := uuid.New()

	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	_, err := svc.GetStats(context.Background(), userID)

	svcErr := err.(*Error)
	assert.Equal(t, "forbidden", svcErr.Code)
}

func TestGetStats_RepoError(t *testing.T) {
	svc, statsRepo, roleRepo := setupStatsService()
	userID := uuid.New()
	roleID := uuid.New()

	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, PermissionStatsRead).Return(true, nil)
	statsRepo.On("GetStats", mock.Anything).Return(nil, assert.AnError)

	_, err := svc.GetStats(context.Background(), userID)

	assert.Error(t, err)
	svcErr := err.(*Error)
	assert.Equal(t, "internal_error", svcErr.Code)
}
