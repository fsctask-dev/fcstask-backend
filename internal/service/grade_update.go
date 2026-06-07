package service

import (
	"context"
	"fcstask-backend/internal/metrics"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"

	"github.com/google/uuid"
)

type UpdateGradeInput struct {
	StudentID uuid.UUID `json:"student_id"`
	TaskID    uuid.UUID `json:"task_id"`
	CourseID  uuid.UUID `json:"course_id"`
	Score     *int      `json:"score"`
	UserID    uuid.UUID `json:"user_id"`
}

func (s *GradeUpdateService) WithMetrics(m *metrics.AdminMetrics) *GradeUpdateService {
	s.adminMetrics = m
	return s
}

type GradeUpdateService struct {
	taskRepo     repo.ITaskRepo
	homeworkRepo repo.IHomeworkRepo
	scoreRepo    repo.IStudentTaskScoreRepo
	roleRepo     repo.IRoleRepo
	adminMetrics *metrics.AdminMetrics
}

func NewGradeUpdateService(
	taskRepo repo.ITaskRepo,
	homeworkRepo repo.IHomeworkRepo,
	scoreRepo repo.IStudentTaskScoreRepo,
	roleRepo repo.IRoleRepo,
) *GradeUpdateService {
	return &GradeUpdateService{
		taskRepo:     taskRepo,
		homeworkRepo: homeworkRepo,
		scoreRepo:    scoreRepo,
		roleRepo:     roleRepo,
	}
}

func (s *GradeUpdateService) UpdateGrade(ctx context.Context, userID uuid.UUID, input UpdateGradeInput) (score *model.StudentTaskScore, err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionUpdateGrade, adminOutcome(err)) }()
	if err := RequireScopedPermission(ctx, s.roleRepo, userID, input.CourseID, PermissionGradeUpdate); err != nil {
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
	if input.Score == nil {
		return nil, BadRequest("score is required")
	}
	if *input.Score < 0 {
		return nil, BadRequest("score must be 0 or higher")
	}

	task, err := s.taskRepo.GetByID(ctx, input.TaskID)
	if err != nil {
		return nil, NotFound("task not found")
	}
	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return nil, NotFound("homework not found")
	}
	if hw.CourseID != input.CourseID {
		return nil, BadRequest("task does not belong to the specified course")
	}

	isMember, err := HasScopedPermission(ctx, s.roleRepo, input.StudentID, input.CourseID, PermissionCourseRead)
	if err != nil {
		return nil, Internal("failed to check student membership", err)
	}
	if !isMember {
		return nil, BadRequest("student is not a member of this course")
	}

	score = &model.StudentTaskScore{
		StudentID: input.StudentID,
		TaskID:    input.TaskID,
		CourseID:  input.CourseID,
		Score:     *input.Score,
	}

	if err := s.scoreRepo.Upsert(ctx, score); err != nil {
		return nil, Internal("Failed to update grade", err)
	}
	return score, nil
}
