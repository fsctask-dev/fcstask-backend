package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

func setupLatePolicyService() (*service.CourseLatePolicy, *MockCourseLateRepo) {
	lateRepo := new(MockCourseLateRepo)
	roleRepo := newPermissiveRoleRepo()
	svc := service.NewCourseLatePolicy(lateRepo, roleRepo)
	return svc, lateRepo
}

func TestCourseLatePolicy_InvalidType(t *testing.T) {
	svc, _ := setupLatePolicyService()
	_, err := svc.CreateOrUpdate(context.Background(), uuid.New(), uuid.New(), service.CourseLateInput{
		PolicyType: "bad_type",
	})
	assert.ErrorContains(t, err, "policy_type must be one of")
}

func TestCourseLatePolicy_LinearBadScore(t *testing.T) {
	svc, _ := setupLatePolicyService()
	_, err := svc.CreateOrUpdate(context.Background(), uuid.New(), uuid.New(), service.CourseLateInput{
		PolicyType: model.PolicyTypeLinear, HardDeadlineScore: 1.5,
	})
	assert.ErrorContains(t, err, "hard_deadline_score must be between")
}

func TestCourseLatePolicy_StepMissingPercent(t *testing.T) {
	svc, _ := setupLatePolicyService()
	_, err := svc.CreateOrUpdate(context.Background(), uuid.New(), uuid.New(), service.CourseLateInput{
		PolicyType: model.PolicyTypeStep,
	})
	assert.ErrorContains(t, err, "step_percent is required")
}

func TestCourseLatePolicy_StepBadPercent(t *testing.T) {
	svc, _ := setupLatePolicyService()
	_, err := svc.CreateOrUpdate(context.Background(), uuid.New(), uuid.New(), service.CourseLateInput{
		PolicyType: model.PolicyTypeStep, StepPercent: f64Ptr(1.5),
	})
	assert.ErrorContains(t, err, "step_percent must be between")
}

func TestCourseLatePolicy_CoefficientMissing(t *testing.T) {
	svc, _ := setupLatePolicyService()
	_, err := svc.CreateOrUpdate(context.Background(), uuid.New(), uuid.New(), service.CourseLateInput{
		PolicyType: model.PolicyTypeCoefficient,
	})
	assert.ErrorContains(t, err, "coefficient is required")
}

func TestCourseLatePolicy_CoefficientBadValue(t *testing.T) {
	svc, _ := setupLatePolicyService()
	_, err := svc.CreateOrUpdate(context.Background(), uuid.New(), uuid.New(), service.CourseLateInput{
		PolicyType: model.PolicyTypeCoefficient, Coefficient: f64Ptr(-0.1),
	})
	assert.ErrorContains(t, err, "coefficient must be between")
}

func TestCourseLatePolicy_Create_Linear(t *testing.T) {
	svc, lateRepo := setupLatePolicyService()
	ctx := context.Background()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeLinear && p.HardDeadlineScore == 0.5
	})).Return(nil)

	result, err := svc.CreateOrUpdate(ctx, uuid.New(), courseID, service.CourseLateInput{
		PolicyType: model.PolicyTypeLinear, HardDeadlineScore: 0.5,
	})
	assert.NoError(t, err)
	assert.Equal(t, model.PolicyTypeLinear, result.PolicyType)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Create_Step(t *testing.T) {
	svc, lateRepo := setupLatePolicyService()
	ctx := context.Background()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeStep
	})).Return(nil)

	result, err := svc.CreateOrUpdate(ctx, uuid.New(), courseID, service.CourseLateInput{
		PolicyType: model.PolicyTypeStep, StepPercent: f64Ptr(0.1),
	})
	assert.NoError(t, err)
	assert.Equal(t, model.PolicyTypeStep, result.PolicyType)
}

func TestCourseLatePolicy_Create_Coefficient(t *testing.T) {
	svc, lateRepo := setupLatePolicyService()
	ctx := context.Background()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeCoefficient
	})).Return(nil)

	result, err := svc.CreateOrUpdate(ctx, uuid.New(), courseID, service.CourseLateInput{
		PolicyType: model.PolicyTypeCoefficient, Coefficient: f64Ptr(0.8),
	})
	assert.NoError(t, err)
	assert.Equal(t, model.PolicyTypeCoefficient, result.PolicyType)
}

func TestCourseLatePolicy_Update_Existing(t *testing.T) {
	svc, lateRepo := setupLatePolicyService()
	ctx := context.Background()
	courseID := uuid.New()

	existing := &model.CourseLatePolicy{
		ID: uuid.New(), CourseID: courseID,
		PolicyType: model.PolicyTypeLinear, HardDeadlineScore: 0.3,
	}
	lateRepo.On("GetByCourseID", ctx, courseID).Return(existing, nil)
	lateRepo.On("Update", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.HardDeadlineScore == 0.7
	})).Return(nil)

	result, err := svc.CreateOrUpdate(ctx, uuid.New(), courseID, service.CourseLateInput{
		PolicyType: model.PolicyTypeLinear, HardDeadlineScore: 0.7,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0.7, result.HardDeadlineScore)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_RepoError_OnFetch(t *testing.T) {
	svc, lateRepo := setupLatePolicyService()
	ctx := context.Background()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, assert.AnError)

	_, err := svc.CreateOrUpdate(ctx, uuid.New(), courseID, service.CourseLateInput{
		PolicyType: model.PolicyTypeLinear, HardDeadlineScore: 0.5,
	})
	assert.ErrorContains(t, err, "Failed to fetch late policy")
}
