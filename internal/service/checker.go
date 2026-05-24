package service

import (
	"context"
	"time"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"

	"github.com/google/uuid"
)

type CheckerService struct {
	taskRepo       repo.ITaskRepo
	scoreRepo      repo.IStudentTaskScoreRepo
	latePolicyRepo repo.ILatePolicyRepo
}

func NewCheckerService(
	taskRepo repo.ITaskRepo,
	scoreRepo repo.IStudentTaskScoreRepo,
	latePolicyRepo repo.ILatePolicyRepo,
) *CheckerService {
	return &CheckerService{
		taskRepo:       taskRepo,
		scoreRepo:      scoreRepo,
		latePolicyRepo: latePolicyRepo,
	}
}

type SubmitGradeInput struct {
	StudentID   uuid.UUID
	TaskID      uuid.UUID
	CourseID    uuid.UUID
	RawScore    int
	IsPassed    bool
	SubmittedAt time.Time
}

func (s *CheckerService) SubmitGrade(ctx context.Context, input SubmitGradeInput) (*model.StudentTaskScore, error) {
	if input.StudentID == uuid.Nil {
		return nil, BadRequest("student_id is required")
	}
	if input.TaskID == uuid.Nil {
		return nil, BadRequest("task_id is required")
	}
	if input.CourseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	if input.RawScore < 0 {
		return nil, BadRequest("score must be non-negative")
	}

	if _, err := s.taskRepo.GetByID(ctx, input.TaskID); err != nil {
		return nil, NotFound("Task not found")
	}

	finalScore := s.applyLatePolicy(ctx, input.TaskID, input.RawScore, input.SubmittedAt)

	score := &model.StudentTaskScore{
		StudentID: input.StudentID,
		TaskID:    input.TaskID,
		CourseID:  input.CourseID,
		Score:     finalScore,
		IsPassed:  input.IsPassed,
		UpdatedAt: time.Now(),
	}

	if err := s.scoreRepo.Upsert(ctx, score); err != nil {
		return nil, Internal("Failed to save grade", err)
	}

	return score, nil
}

func (s *CheckerService) applyLatePolicy(ctx context.Context, taskID uuid.UUID, rawScore int, submittedAt time.Time) int {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return rawScore
	}

	policy, err := s.latePolicyRepo.GetByHwID(ctx, task.HwID)
	if err != nil {
		return rawScore
	}

	if err != nil {
		return rawScore
	}

	if submittedAt.Before(policy.SoftDeadline) || submittedAt.Equal(policy.SoftDeadline) {
		return rawScore
	}

	if submittedAt.Before(policy.HardDeadline) || submittedAt.Equal(policy.HardDeadline) {
		return int(float64(rawScore) * policy.SoftPenalty)
	}

	return int(float64(rawScore) * policy.HardPenalty)
}
