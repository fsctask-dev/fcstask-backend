package service

import (
	"context"

	"github.com/google/uuid"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/metrics"
)

type AdminTaskService struct {
	taskRepo     repo.ITaskRepo
	homeworkRepo repo.IHomeworkRepo
	roleRepo     repo.IRoleRepo

	adminMetrics *metrics.AdminMetrics
}

func NewAdminTaskService(taskRepo repo.ITaskRepo, homeworkRepo repo.IHomeworkRepo, roleRepo repo.IRoleRepo) *AdminTaskService {
	return &AdminTaskService{
		taskRepo:     taskRepo,
		homeworkRepo: homeworkRepo,
		roleRepo:     roleRepo,
	}
}

func (s *AdminTaskService) WithMetrics(m *metrics.AdminMetrics) *AdminTaskService {
	s.adminMetrics = m
	return s
}

type CreateTaskInput struct {
	CourseID uuid.UUID
	HwID     uuid.UUID
	Title    *string
	RepoURL  string
	TaskURL  string
	Score    int
}

type UpdateTaskInput struct {
	Title   *string
	RepoURL string
	TaskURL string
	Score   int
}

type PublishTaskInput struct {
	CourseID uuid.UUID
	HwID     uuid.UUID
	TaskID   uuid.UUID
	IsPublic bool
}

type SetTaskScoreInput struct {
	CourseID uuid.UUID
	HwID     uuid.UUID
	TaskID   uuid.UUID
	Score    int
}

func (s *AdminTaskService) CreateTask(ctx context.Context, userID uuid.UUID, input CreateTaskInput) (task *model.Task, err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionCreateTask, adminOutcome(err)) }()

	if input.CourseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	if input.HwID == uuid.Nil {
		return nil, BadRequest("homework_id is required")
	}
	if input.Title == nil || *input.Title == "" {
		return nil, BadRequest("title is required")
	}

	hw, err := requireHomeworkInCourse(ctx, s.homeworkRepo, input.HwID, input.CourseID)
	if err != nil {
		return nil, err
	}
	if err = RequireScopedPermission(ctx, s.roleRepo, userID, hw.CourseID, PermissionTaskCreate); err != nil {
		return nil, err
	}

	if input.Score <= 0 {
		return nil, BadRequest("score must be positive")
	}

	task = &model.Task{
		HwID:  input.HwID,
		Score: &input.Score,
		Title: *input.Title,
	}

	if input.RepoURL != "" {
		task.RepoURL = stringPtr(input.RepoURL)
	}
	if input.TaskURL != "" {
		task.TaskURL = stringPtr(input.TaskURL)
	}
	if err = s.taskRepo.Create(ctx, task); err != nil {
		return nil, Internal("Failed to create task", err)
	}

	return task, nil
}

func (s *AdminTaskService) GetTask(ctx context.Context, userID, courseID, hwID, taskID uuid.UUID) (*model.Task, error) {
	if taskID == uuid.Nil {
		return nil, BadRequest("task_id is required")
	}

	task, err := requireTaskInHomework(ctx, s.taskRepo, taskID, hwID)
	if err != nil {
		return nil, err
	}
	hw, err := requireHomeworkInCourse(ctx, s.homeworkRepo, hwID, courseID)
	if err != nil {
		return nil, err
	}
	if err := RequireScopedPermission(ctx, s.roleRepo, userID, hw.CourseID, PermissionTaskRead); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *AdminTaskService) ListTasks(ctx context.Context, userID, courseID, hwID uuid.UUID) ([]model.Task, error) {
	if hwID == uuid.Nil {
		return nil, BadRequest("homework_id is required")
	}

	hw, err := requireHomeworkInCourse(ctx, s.homeworkRepo, hwID, courseID)
	if err != nil {
		return nil, err
	}
	if err := RequireScopedPermission(ctx, s.roleRepo, userID, hw.CourseID, PermissionTaskRead); err != nil {
		return nil, err
	}

	tasks, err := s.taskRepo.GetByHwID(ctx, hwID)
	if err != nil {
		return nil, Internal("Failed to fetch tasks", err)
	}

	return tasks, nil
}

func (s *AdminTaskService) UpdateTask(ctx context.Context, userID, courseID, hwID, taskID uuid.UUID, input UpdateTaskInput) (task *model.Task, err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionUpdateTask, adminOutcome(err)) }()

	task, err = s.GetTask(ctx, userID, courseID, hwID, taskID)
	if err != nil {
		return nil, err
	}
	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}
	if err = RequireScopedPermission(ctx, s.roleRepo, userID, hw.CourseID, PermissionTaskUpdate); err != nil {
		return nil, err
	}

	if input.Title != nil {
		if *input.Title == "" {
			return nil, BadRequest("title cannot be empty")
		}
		task.Title = *input.Title
	}

	if input.RepoURL != "" {
		task.RepoURL = stringPtr(input.RepoURL)
	}
	if input.TaskURL != "" {
		task.TaskURL = stringPtr(input.TaskURL)
	}
	if input.Score > 0 {
		task.Score = &input.Score
	}

	if err = s.taskRepo.Update(ctx, task); err != nil {
		return nil, Internal("Failed to update task", err)
	}

	return task, nil
}

func (s *AdminTaskService) PublishTask(ctx context.Context, userID uuid.UUID, input PublishTaskInput) (*model.Task, error) {
	if input.TaskID == uuid.Nil {
		return nil, BadRequest("task_id is required")
	}
	if input.HwID == uuid.Nil {
		return nil, BadRequest("homework_id is required")
	}
	if input.CourseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}

	task, err := s.GetTask(ctx, userID, input.CourseID, input.HwID, input.TaskID)
	if err != nil {
		return nil, err
	}

	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}

	if err := RequireScopedPermission(ctx, s.roleRepo, userID, hw.CourseID, PermissionTaskPublish); err != nil {
		return nil, err
	}

	task.IsPublic = &input.IsPublic

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, Internal("Failed to update task visibility", err)
	}

	return task, nil
}

func (s *AdminTaskService) DeleteTask(ctx context.Context, userID, courseID, hwID, taskID uuid.UUID) (err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionDeleteTask, adminOutcome(err)) }()

	if taskID == uuid.Nil {
		return BadRequest("task_id is required")
	}

	task, err := s.GetTask(ctx, userID, courseID, hwID, taskID)
	if err != nil {
		return err
	}
	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return NotFound("Homework not found")
	}
	if err = RequireScopedPermission(ctx, s.roleRepo, userID, hw.CourseID, PermissionTaskDelete); err != nil {
		return err
	}

	if err = s.taskRepo.Delete(ctx, taskID); err != nil {
		return Internal("Failed to delete task", err)
	}

	return nil
}

func (s *AdminTaskService) SetScore(ctx context.Context, userID uuid.UUID, input SetTaskScoreInput) (task *model.Task, err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionScoreTask, adminOutcome(err)) }()

	if input.HwID == uuid.Nil {
		return nil, BadRequest("homework_id is required")
	}
	if input.TaskID == uuid.Nil {
		return nil, BadRequest("task_id is required")
	}
	if input.CourseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	if input.Score <= 0 {
		return nil, BadRequest("score must be positive")
	}

	task, err = s.GetTask(ctx, userID, input.CourseID, input.HwID, input.TaskID)
	if err != nil {
		return nil, err
	}
	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}
	if err = RequireScopedPermission(ctx, s.roleRepo, userID, hw.CourseID, PermissionTaskScoreUpdate); err != nil {
		return nil, err
	}

	if err = s.taskRepo.SetScore(ctx, input.TaskID, input.Score); err != nil {
		return nil, Internal("Failed to set score", err)
	}

	task, err = s.GetTask(ctx, userID, input.CourseID, input.HwID, input.TaskID)
	if err != nil {
		return nil, err
	}

	return task, nil
}
