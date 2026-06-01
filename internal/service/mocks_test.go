package service_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
)

type MockScoreRepo struct{ mock.Mock }

func (m *MockScoreRepo) Upsert(ctx context.Context, s *model.StudentTaskScore) error {
	return m.Called(ctx, s).Error(0)
}
func (m *MockScoreRepo) GetByStudentAndCourse(ctx context.Context, studentID, courseID uuid.UUID) ([]model.StudentTaskScore, error) {
	args := m.Called(ctx, studentID, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.StudentTaskScore), args.Error(1)
}

type MockHwDeadlineRepo struct{ mock.Mock }

func (m *MockHwDeadlineRepo) GetByHwID(ctx context.Context, hwID uuid.UUID) (*model.HomeworkDeadline, error) {
	args := m.Called(ctx, hwID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.HomeworkDeadline), args.Error(1)
}
func (m *MockHwDeadlineRepo) Create(ctx context.Context, d *model.HomeworkDeadline) error {
	return m.Called(ctx, d).Error(0)
}
func (m *MockHwDeadlineRepo) Update(ctx context.Context, d *model.HomeworkDeadline) error {
	return m.Called(ctx, d).Error(0)
}

type MockCourseLateRepo struct{ mock.Mock }

func (m *MockCourseLateRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) (*model.CourseLatePolicy, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CourseLatePolicy), args.Error(1)
}
func (m *MockCourseLateRepo) Create(ctx context.Context, p *model.CourseLatePolicy) error {
	return m.Called(ctx, p).Error(0)
}
func (m *MockCourseLateRepo) Update(ctx context.Context, p *model.CourseLatePolicy) error {
	return m.Called(ctx, p).Error(0)
}

func newPermissiveRoleRepo() *MockRoleRepo {
	rr := new(MockRoleRepo)
	roleID := uuid.New()
	rr.On("GetRoleIDByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).Return(roleID, nil)
	rr.On("HasPermission", mock.Anything, roleID, mock.Anything).Return(true, nil)
	return rr
}
