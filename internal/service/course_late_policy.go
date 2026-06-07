package service

import (
	"context"
	"errors"
	"fcstask-backend/internal/metrics"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type CourseLatePolicy struct {
	courseLateRepo    repo.ICourseLatePolicy
	roleRepo          repo.IRoleRepo
	latepolicyMetrics *metrics.LatePolicyMetrics
}

func (s *CourseLatePolicy) WithMetrics(m *metrics.LatePolicyMetrics) *CourseLatePolicy {
	s.latepolicyMetrics = m
	return s
}

func NewCourseLatePolicy(r repo.ICourseLatePolicy, rr repo.IRoleRepo) *CourseLatePolicy {
	return &CourseLatePolicy{courseLateRepo: r, roleRepo: rr}
}

type CourseLateInput struct {
	PolicyType        model.PolicyType
	SoftPenalty       float64
	HardDeadlineScore float64
	StepPercent       *float64
	Coefficient       *float64
}

func (s *CourseLatePolicy) CreateOrUpdate(ctx context.Context, userID, courseID uuid.UUID, in CourseLateInput) (result *model.CourseLatePolicy, err error) {
	defer func() { s.latepolicyMetrics.IncAction(metrics.LatePolicyActionCreateOrUpdate, adminOutcome(err)) }()
	if err := RequireScopedPermission(ctx, s.roleRepo, userID, courseID, PermissionLatePolicyCreate); err != nil {
		return nil, err
	}
	if err := validateCourseLateInput(in); err != nil {
		return nil, err
	}

	existing, err := s.courseLateRepo.GetByCourseID(ctx, courseID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, Internal("Failed to fetch late policy", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) || existing == nil {
		p := &model.CourseLatePolicy{
			CourseID:          courseID,
			PolicyType:        in.PolicyType,
			SoftPenalty:       in.SoftPenalty,
			HardDeadlineScore: in.HardDeadlineScore,
			StepPercent:       in.StepPercent,
			Coefficient:       in.Coefficient,
		}
		if err := s.courseLateRepo.Create(ctx, p); err != nil {
			return nil, Internal("Failed to create late policy", err)
		}
		return p, nil
	}

	existing.PolicyType = in.PolicyType
	existing.SoftPenalty = in.SoftPenalty
	existing.HardDeadlineScore = in.HardDeadlineScore
	existing.StepPercent = in.StepPercent
	existing.Coefficient = in.Coefficient
	if err := s.courseLateRepo.Update(ctx, existing); err != nil {
		return nil, Internal("Failed to update late policy", err)
	}
	return existing, nil
}

func validateCourseLateInput(in CourseLateInput) error {
	switch in.PolicyType {
	case model.PolicyTypeLinear:
		if in.HardDeadlineScore < 0 || in.HardDeadlineScore > 1 {
			return BadRequest("hard_deadline_score must be between 0.0 and 1.0")
		}
		if in.SoftPenalty < 0 || in.SoftPenalty > 1 {
			return BadRequest("soft_penalty must be between 0.0 and 1.0")
		}
		if 1-in.SoftPenalty < in.HardDeadlineScore {
			return BadRequest("soft_deadline_score must be higher than or equal to hard_deadline_score")
		}
	case model.PolicyTypeStep:
		if in.StepPercent == nil {
			return BadRequest("step_percent is required for step policy")
		}
		if *in.StepPercent < 0 || *in.StepPercent > 1 {
			return BadRequest("step_percent must be between 0.0 and 1.0")
		}
	case model.PolicyTypeCoefficient:
		if in.Coefficient == nil {
			return BadRequest("coefficient is required for coefficient policy")
		}
		if *in.Coefficient < 0 || *in.Coefficient > 1 {
			return BadRequest("coefficient must be between 0.0 and 1.0")
		}
	default:
		return BadRequest("policy_type must be one of: linear, step, coefficient")
	}
	if in.SoftPenalty < 0 || in.SoftPenalty > 1 {
		return BadRequest("soft_penalty must be between 0.0 and 1.0")
	}
	return nil
}
