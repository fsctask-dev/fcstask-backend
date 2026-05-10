package service

import (
	"context"

	"github.com/google/uuid"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type AdminTaskService struct {
	taskRepo     repo.ITaskRepo
	homeworkRepo repo.IHomeworkRepo
	roleRepo     repo.IRoleRepo
}

func NewAdminTaskService(taskRepo repo.ITaskRepo, homeworkRepo repo.IHomeworkRepo, roleRepo repo.IRoleRepo) *AdminTaskService {
	return &AdminTaskService{
		taskRepo:     taskRepo,
		homeworkRepo: homeworkRepo,
		roleRepo:     roleRepo,
	}
}

type CreateTaskInput struct {
	HwID    uuid.UUID
	RepoURL string
	TaskURL string
	Score   int
}

type UpdateTaskInput struct {
	RepoURL string
	TaskURL string
	Score   int
}

type SetTaskScoreInput struct {
	TaskID uuid.UUID
	Score  int
}

func (s *AdminTaskService) CreateTask(ctx context.Context, userID uuid.UUID, input CreateTaskInput) (*model.Task, error) {
	if input.HwID == uuid.Nil {
		return nil, BadRequest("homework_id is required")
	}
	hw, err := s.homeworkRepo.GetByID(ctx, input.HwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}
	isAdmin, err := IsCourseAdmin(ctx, s.roleRepo, userID, hw.CourseID)
	if err != nil {
		return nil, Internal("Failed to check permissions", err)
	}
	if !isAdmin {
		return nil, Forbidden("You don't have permission to manage this course")
	}
	if input.Score <= 0 {
		return nil, BadRequest("score must be positive")
	}
	task := &model.Task{
		HwID:  input.HwID,
		Score: &input.Score,
	}
	if input.RepoURL != "" {
		task.RepoURL = stringPtr(input.RepoURL)
	}
	if input.TaskURL != "" {
		task.TaskURL = stringPtr(input.TaskURL)
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, Internal("Failed to create task", err)
	}

	return task, nil
}

func (s *AdminTaskService) GetTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) (*model.Task, error) {
	if taskID == uuid.Nil {
		return nil, BadRequest("task_id is required")
	}
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, NotFound("Task not found")
	}
	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}
	isAdmin, err := IsCourseAdmin(ctx, s.roleRepo, userID, hw.CourseID)
	if err != nil {
		return nil, Internal("Failed to check permissions", err)
	}
	if !isAdmin {
		return nil, Forbidden("You don't have permission to manage this course")
	}

	return task, nil
}

func (s *AdminTaskService) ListTasks(ctx context.Context, userID uuid.UUID, hwID uuid.UUID) ([]model.Task, error) {
	if hwID == uuid.Nil {
		return nil, BadRequest("homework_id is required")
	}
	hw, err := s.homeworkRepo.GetByID(ctx, hwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}
	isAdmin, err := IsCourseAdmin(ctx, s.roleRepo, userID, hw.CourseID)
	if err != nil {
		return nil, Internal("Failed to check permissions", err)
	}
	if !isAdmin {
		return nil, Forbidden("You don't have permission to manage this course")
	}
	tasks, err := s.taskRepo.GetByHwID(ctx, hwID)
	if err != nil {
		return nil, Internal("Failed to fetch tasks", err)
	}
	return tasks, nil
}

func (s *AdminTaskService) UpdateTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, input UpdateTaskInput) (*model.Task, error) {
	task, err := s.GetTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
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
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, Internal("Failed to update task", err)
	}

	return task, nil
}

func (s *AdminTaskService) DeleteTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error {
	if taskID == uuid.Nil {
		return BadRequest("task_id is required")
	}
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return NotFound("Task not found")
	}
	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return NotFound("Homework not found")
	}
	isAdmin, err := IsCourseAdmin(ctx, s.roleRepo, userID, hw.CourseID)
	if err != nil {
		return Internal("Failed to check permissions", err)
	}
	if !isAdmin {
		return Forbidden("You don't have permission to manage this course")
	}
	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return Internal("Failed to delete task", err)
	}

	return nil
}

func (s *AdminTaskService) SetScore(ctx context.Context, userID uuid.UUID, input SetTaskScoreInput) (*model.Task, error) {
	if input.TaskID == uuid.Nil {
		return nil, BadRequest("task_id is required")
	}
	if input.Score <= 0 {
		return nil, BadRequest("score must be positive")
	}
	task, err := s.taskRepo.GetByID(ctx, input.TaskID)
	if err != nil {
		return nil, NotFound("Task not found")
	}
	hw, err := s.homeworkRepo.GetByID(ctx, task.HwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}
	isAdmin, err := IsCourseAdmin(ctx, s.roleRepo, userID, hw.CourseID)
	if err != nil {
		return nil, Internal("Failed to check permissions", err)
	}
	if !isAdmin {
		return nil, Forbidden("You don't have permission to manage this course")
	}
	if err := s.taskRepo.SetScore(ctx, input.TaskID, input.Score); err != nil {
		return nil, Internal("Failed to set score", err)
	}
	task, err = s.taskRepo.GetByID(ctx, input.TaskID)
	if err != nil {
		return nil, err
	}

	return task, nil
}