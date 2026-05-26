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
	homeworkRepo   repo.IHomeworkRepo
	scoreRepo      repo.IStudentTaskScoreRepo
	latePolicyRepo repo.ILatePolicyRepo
}

func NewCheckerService(
	taskRepo repo.ITaskRepo,
	homeworkRepo repo.IHomeworkRepo,
	scoreRepo repo.IStudentTaskScoreRepo,
	latePolicyRepo repo.ILatePolicyRepo,
) *CheckerService {
	return &CheckerService{
		taskRepo:       taskRepo,
		homeworkRepo:   homeworkRepo,
		scoreRepo:      scoreRepo,
		latePolicyRepo: latePolicyRepo,
	}
}

type SubmitGradeInput struct {
	StudentID   uuid.UUID
	TaskID      uuid.UUID
	CourseID    uuid.UUID
	Status      string
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
	if input.Status != "passed" && input.Status != "fail" {
		return nil, BadRequest("status must be 'passed' or 'fail'")
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
	if _, err := s.taskRepo.GetByID(ctx, input.TaskID); err != nil {
		return nil, NotFound("Task not found")
	}

	finalScore := s.applyLatePolicy(ctx, input.TaskID, input.SubmittedAt, input.Status)
	var passed = (input.Status == "passed")
	score := &model.StudentTaskScore{
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

func (s *CheckerService) applyLatePolicy(ctx context.Context, taskID uuid.UUID, submittedAt time.Time, Status string) int {
	if !(Status == "passed") {
		return 0
	}
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return 0
	}
	if task.Score == nil {
		return 0
	}
	var score = task.Score
	policy, err := s.latePolicyRepo.GetByHwID(ctx, task.HwID)
	if err != nil {
		return *score
	}

	if submittedAt.Before(policy.SoftDeadline) || submittedAt.Equal(policy.SoftDeadline) {
		return *score
	}

	if submittedAt.Before(policy.HardDeadline) || submittedAt.Equal(policy.HardDeadline) {
		return int(float64(*score) * policy.SoftPenalty)
	}

	return int(float64(*score) * policy.HardPenalty)
}
