package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

func setupLatePolicyService(hasPermission bool) (*service.CourseLatePolicy, *MockCourseLateRepo, *MockRoleRepo) {
	lateRepo := new(MockCourseLateRepo)
	roleRepo := new(MockRoleRepo)
	roleID := uuid.New()
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionLatePolicyCreate).Return(hasPermission, nil)
	svc := service.NewCourseLatePolicy(lateRepo, roleRepo)
	return svc, lateRepo, roleRepo
}

func TestCourseLatePolicy_InvalidPolicyType(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType: "invalid_type",
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "policy_type must be one of")
}

func TestCourseLatePolicy_Linear_Valid(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeLinear &&
			p.SoftPenalty == 0.2 &&
			p.HardDeadlineScore == 0.5
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.2,
		HardDeadlineScore: 0.5,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Linear_HardDeadlineScoreTooHigh(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.2,
		HardDeadlineScore: 1.5,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "hard_deadline_score must be between 0.0 and 1.0")
}

func TestCourseLatePolicy_Linear_HardDeadlineScoreNegative(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.2,
		HardDeadlineScore: -0.5,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "hard_deadline_score must be between 0.0 and 1.0")
}

func TestCourseLatePolicy_Linear_SoftScoreLessThanHardScore(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.3,
		HardDeadlineScore: 0.8, // Soft score = 1 - 0.3 = 0.7, which is less than 0.8
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "soft_deadline_score must be higher than or equal to hard_deadline_score")
}

func TestCourseLatePolicy_Linear_SoftScoreEqualToHardScore(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.Anything).Return(nil)

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.3,
		HardDeadlineScore: 0.7, // Soft score = 1 - 0.3 = 0.7, equal to hard score
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCourseLatePolicy_Step_Valid(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeStep &&
			p.StepPercent != nil &&
			*p.StepPercent == 0.1 &&
			p.SoftPenalty == 0.2
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeStep,
		SoftPenalty: 0.2,
		StepPercent: f64Ptr(0.1),
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Step_MissingStepPercent(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeStep,
		SoftPenalty: 0.2,
		StepPercent: nil,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "step_percent is required")
}

func TestCourseLatePolicy_Step_StepPercentTooHigh(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeStep,
		SoftPenalty: 0.2,
		StepPercent: f64Ptr(1.5),
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "step_percent must be between 0.0 and 1.0")
}

func TestCourseLatePolicy_Step_StepPercentNegative(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeStep,
		SoftPenalty: 0.2,
		StepPercent: f64Ptr(-0.1),
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "step_percent must be between 0.0 and 1.0")
}

func TestCourseLatePolicy_Step_StepPercentZero(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeStep &&
			p.StepPercent != nil &&
			*p.StepPercent == 0.0
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeStep,
		SoftPenalty: 0.2,
		StepPercent: f64Ptr(0.0),
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCourseLatePolicy_Coefficient_Valid(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeCoefficient &&
			p.Coefficient != nil &&
			*p.Coefficient == 0.8 &&
			p.SoftPenalty == 0.2
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeCoefficient,
		SoftPenalty: 0.2,
		Coefficient: f64Ptr(0.8),
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Coefficient_MissingCoefficient(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeCoefficient,
		SoftPenalty: 0.2,
		Coefficient: nil,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "coefficient is required")
}

func TestCourseLatePolicy_Coefficient_CoefficientTooHigh(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeCoefficient,
		SoftPenalty: 0.2,
		Coefficient: f64Ptr(1.5),
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "coefficient must be between 0.0 and 1.0")
}

func TestCourseLatePolicy_Coefficient_CoefficientNegative(t *testing.T) {
	svc, _, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeCoefficient,
		SoftPenalty: 0.2,
		Coefficient: f64Ptr(-0.1),
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "coefficient must be between 0.0 and 1.0")
}

func TestCourseLatePolicy_Coefficient_CoefficientZero(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeCoefficient &&
			p.Coefficient != nil &&
			*p.Coefficient == 0.0
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeCoefficient,
		SoftPenalty: 0.2,
		Coefficient: f64Ptr(0.0),
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCourseLatePolicy_Coefficient_CoefficientOne(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeCoefficient &&
			p.Coefficient != nil &&
			*p.Coefficient == 1.0
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeCoefficient,
		SoftPenalty: 0.2,
		Coefficient: f64Ptr(1.0),
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCourseLatePolicy_Create_Linear_WithAllFields(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.CourseID == courseID &&
			p.PolicyType == model.PolicyTypeLinear &&
			p.SoftPenalty == 0.15 &&
			p.HardDeadlineScore == 0.4 &&
			p.StepPercent == nil &&
			p.Coefficient == nil
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.15,
		HardDeadlineScore: 0.4,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, courseID, result.CourseID)
	assert.Equal(t, model.PolicyTypeLinear, result.PolicyType)
	assert.Equal(t, 0.15, result.SoftPenalty)
	assert.Equal(t, 0.4, result.HardDeadlineScore)
	assert.Nil(t, result.StepPercent)
	assert.Nil(t, result.Coefficient)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Create_Step_WithAllFields(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()
	stepPercent := 0.25

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.CourseID == courseID &&
			p.PolicyType == model.PolicyTypeStep &&
			p.SoftPenalty == 0.1 &&
			p.StepPercent != nil &&
			*p.StepPercent == stepPercent &&
			p.HardDeadlineScore == 0.0
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeStep,
		SoftPenalty: 0.1,
		StepPercent: &stepPercent,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, stepPercent, *result.StepPercent)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Create_Coefficient_WithAllFields(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()
	coefficient := 0.75

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.CourseID == courseID &&
			p.PolicyType == model.PolicyTypeCoefficient &&
			p.SoftPenalty == 0.05 &&
			p.Coefficient != nil &&
			*p.Coefficient == coefficient &&
			p.HardDeadlineScore == 0.0
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeCoefficient,
		SoftPenalty: 0.05,
		Coefficient: &coefficient,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, coefficient, *result.Coefficient)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Create_GetByCourseIDReturnsOtherError(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, errors.New("database connection error"))

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.1,
		HardDeadlineScore: 0.5,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to fetch late policy")
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Create_RepoCreateError(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, gorm.ErrRecordNotFound)
	lateRepo.On("Create", ctx, mock.AnythingOfType("*model.CourseLatePolicy")).Return(errors.New("db error"))

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.1,
		HardDeadlineScore: 0.5,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to create late policy")
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Update_ExistingPolicy(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	existing := &model.CourseLatePolicy{
		ID:                uuid.New(),
		CourseID:          courseID,
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.1,
		HardDeadlineScore: 0.3,
		StepPercent:       nil,
		Coefficient:       nil,
	}

	lateRepo.On("GetByCourseID", ctx, courseID).Return(existing, nil)
	lateRepo.On("Update", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeLinear &&
			p.SoftPenalty == 0.2 &&
			p.HardDeadlineScore == 0.6 &&
			p.StepPercent == nil &&
			p.Coefficient == nil
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.2,
		HardDeadlineScore: 0.6,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0.2, result.SoftPenalty)
	assert.Equal(t, 0.6, result.HardDeadlineScore)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Update_ChangeFromLinearToStep(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()
	stepPercent := 0.15

	existing := &model.CourseLatePolicy{
		ID:                uuid.New(),
		CourseID:          courseID,
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.1,
		HardDeadlineScore: 0.3,
		StepPercent:       nil,
		Coefficient:       nil,
	}

	lateRepo.On("GetByCourseID", ctx, courseID).Return(existing, nil)
	lateRepo.On("Update", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeStep &&
			p.SoftPenalty == 0.05 &&
			p.StepPercent != nil &&
			*p.StepPercent == stepPercent &&
			p.Coefficient == nil
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeStep,
		SoftPenalty: 0.05,
		StepPercent: &stepPercent,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.PolicyTypeStep, result.PolicyType)
	assert.Equal(t, stepPercent, *result.StepPercent)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Update_ChangeFromStepToCoefficient(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()
	coefficient := 0.9

	stepPercent := 0.1
	existing := &model.CourseLatePolicy{
		ID:                uuid.New(),
		CourseID:          courseID,
		PolicyType:        model.PolicyTypeStep,
		SoftPenalty:       0.1,
		HardDeadlineScore: 0.0,
		StepPercent:       &stepPercent,
		Coefficient:       nil,
	}

	lateRepo.On("GetByCourseID", ctx, courseID).Return(existing, nil)
	lateRepo.On("Update", ctx, mock.MatchedBy(func(p *model.CourseLatePolicy) bool {
		return p.PolicyType == model.PolicyTypeCoefficient &&
			p.SoftPenalty == 0.2 &&
			p.Coefficient != nil &&
			*p.Coefficient == coefficient &&
			p.StepPercent == nil
	})).Return(nil)

	input := service.CourseLateInput{
		PolicyType:  model.PolicyTypeCoefficient,
		SoftPenalty: 0.2,
		Coefficient: &coefficient,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.PolicyTypeCoefficient, result.PolicyType)
	assert.Equal(t, coefficient, *result.Coefficient)
	assert.Nil(t, result.StepPercent)
	lateRepo.AssertExpectations(t)
}

func TestCourseLatePolicy_Update_RepoUpdateError(t *testing.T) {
	svc, lateRepo, _ := setupLatePolicyService(true)
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	existing := &model.CourseLatePolicy{
		ID:                uuid.New(),
		CourseID:          courseID,
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.1,
		HardDeadlineScore: 0.3,
	}

	lateRepo.On("GetByCourseID", ctx, courseID).Return(existing, nil)
	lateRepo.On("Update", ctx, mock.AnythingOfType("*model.CourseLatePolicy")).Return(errors.New("db error"))

	input := service.CourseLateInput{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.2,
		HardDeadlineScore: 0.5,
	}

	result, err := svc.CreateOrUpdate(ctx, userID, courseID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to update late policy")
	lateRepo.AssertExpectations(t)
}
