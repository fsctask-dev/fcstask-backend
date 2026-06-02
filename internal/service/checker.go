package service

import (
	"context"
	"fcstask-backend/internal/metrics"
	"time"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"

	"github.com/google/uuid"
)

type CheckerService struct {
	taskRepo       repo.ITaskRepo
	homeworkRepo   repo.IHomeworkRepo
	scoreRepo      repo.IStudentTaskScoreRepo
	deadlineRepo   repo.IDeadlineRepo
	courseLateRepo repo.ICourseLatePolicy
	checkerMetrics *metrics.CheckerMetrics
	roleRepo       repo.IRoleRepo
}

func (s *CheckerService) WithMetrics(m *metrics.CheckerMetrics) *CheckerService {
	s.checkerMetrics = m
	return s
}

func NewCheckerService(
	taskRepo repo.ITaskRepo,
	homeworkRepo repo.IHomeworkRepo,
	scoreRepo repo.IStudentTaskScoreRepo,
	deadlineRepo repo.IDeadlineRepo,
	courseLateRepo repo.ICourseLatePolicy,
	roleRepo repo.IRoleRepo,
) *CheckerService {
	return &CheckerService{
		taskRepo:       taskRepo,
		homeworkRepo:   homeworkRepo,
		scoreRepo:      scoreRepo,
		deadlineRepo:   deadlineRepo,
		courseLateRepo: courseLateRepo,
		roleRepo:       roleRepo,
	}
}

type SubmitGradeInput struct {
	StudentID   uuid.UUID
	TaskID      uuid.UUID
	CourseID    uuid.UUID
	Status      string
	SubmittedAt time.Time
}

func (s *CheckerService) SubmitGrade(ctx context.Context, input SubmitGradeInput) (score *model.StudentTaskScore, err error) {
	defer func() { s.checkerMetrics.IncAction(metrics.CheckerActionSubmitGrade, adminOutcome(err)) }()

	if err := RequireScopedPermission(ctx, s.roleRepo, input.StudentID, input.CourseID, PermissionTaskSubmit); err != nil {
		return nil, err
	}
	if input.StudentID == uuid.Nil {
		return nil, BadRequest("student_id is required")
	}
	if input.TaskID == uuid.Nil {
		return nil, BadRequest("task_id is required")
	}
	if input.CourseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	if input.Status != "passed" && input.Status != "fail" {
		return nil, BadRequest("status must be 'passed' or 'fail'")
	}
	if input.SubmittedAt.IsZero() {
		return nil, BadRequest("submitted_at is required")
	}
	if input.SubmittedAt.After(time.Now()) {
		return nil, BadRequest("submitted_at cannot be in the future")
	}

	task, err := s.taskRepo.GetByID(ctx, input.TaskID)
	if err != nil {
		return nil, NotFound("Task not found")
	}
	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}
	if hw.CourseID != input.CourseID {
		return nil, BadRequest("task does not belong to the specified course")
	}

	finalScore := s.applyLatePolicy(ctx, input.TaskID, input.SubmittedAt, input.Status)
	var passed = (input.Status == "passed")
	score = &model.StudentTaskScore{
		StudentID: input.StudentID,
		TaskID:    input.TaskID,
		CourseID:  input.CourseID,
		Score:     finalScore,
		IsPassed:  passed,
		UpdatedAt: time.Now(),
	}

	if err := s.scoreRepo.Upsert(ctx, score); err != nil {
		return nil, Internal("Failed to save grade", err)
	}

	return score, nil
}

func (s *CheckerService) applyLatePolicy(ctx context.Context, taskID uuid.UUID, submittedAt time.Time, status string) int {
	if status != "passed" {
		return 0
	}
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task.Score == nil {
		return 0
	}
	baseScore := *task.Score

	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return baseScore
	}

	deadlines, err := s.deadlineRepo.GetByHomeworkID(ctx, task.HwID)
	if err != nil {
		return baseScore
	}
	policy, err := s.courseLateRepo.GetByCourseID(ctx, hw.CourseID)
	if err != nil {
		return baseScore
	}

	soft := deadlines.SoftDeadline
	hard := deadlines.HardDeadline

	if !submittedAt.After(soft) {
		return baseScore
	}
	if submittedAt.After(hard) {
		return 0
	}

	switch policy.PolicyType {

	case model.PolicyTypeLinear:
		total := hard.Sub(soft).Seconds()
		elapsed := submittedAt.Sub(soft).Seconds()
		ratio := elapsed / total
		factor := (1 - policy.SoftPenalty) - ratio*((1-policy.SoftPenalty)-policy.HardDeadlineScore)
		return int(float64(baseScore) * factor)

	case model.PolicyTypeStep:
		daysLate := int(submittedAt.Sub(soft).Hours()/24) + 1
		factor := 1.0 - float64(daysLate)**policy.StepPercent
		if factor < 0 {
			factor = 0
		}
		return int(float64(baseScore) * factor)

	case model.PolicyTypeCoefficient:
		return int(float64(baseScore) * *policy.Coefficient)

	default:
		return baseScore
	}
}
