package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type AdminHomeworkService struct {
	homeworkRepo repo.IHomeworkRepo
	deadlineRepo repo.IDeadlineRepo
}

func NewAdminHomeworkService(homeworkRepo repo.IHomeworkRepo, deadlineRepo repo.IDeadlineRepo) *AdminHomeworkService {
	return &AdminHomeworkService{
		homeworkRepo: homeworkRepo,
		deadlineRepo: deadlineRepo,
	}
}

type CreateHomeworkInput struct {
	CourseID  uuid.UUID
	StartDate string
	EndDate   string
}

type UpdateHomeworkInput struct {
	StartDate string
	EndDate   string
}

type SetDeadlineInput struct {
	CourseID    uuid.UUID
	HomeworkID  uuid.UUID
	Title       string
	Description string
	DueDate     string // RFC3339
	AssignedBy  *uuid.UUID
}

type UpdateDeadlineInput struct {
	Title       string
	Description string
	DueDate     string // RFC3339
}

func (s *AdminHomeworkService) CreateHomework(ctx context.Context, input CreateHomeworkInput) (*model.Homework, error) {
	if input.CourseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	if input.StartDate == "" {
		return nil, BadRequest("start date is required")
	}
	if !IsValidDate(input.StartDate) {
		return nil, BadRequest("start date must be in format YYYY-MM-DD")
	}
	if input.EndDate == "" {
		return nil, BadRequest("end date is required")
	}
	if !IsValidDate(input.EndDate) {
		return nil, BadRequest("end date must be in format YYYY-MM-DD")
	}
	if !IsValidDateRange(input.StartDate, input.EndDate) {
		return nil, BadRequest("end date must be after start_date")
	}
	startDate, err := parseDatePtr(input.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := parseDatePtr(input.EndDate)
	if err != nil {
		return nil, err
	}
	hw := &model.Homework{
		CourseID:  input.CourseID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	if err := s.homeworkRepo.Create(ctx, hw); err != nil {
		return nil, Internal("Failed to create homework", err)
	}

	return hw, nil
}

func (s *AdminHomeworkService) GetHomework(ctx context.Context, hwID uuid.UUID) (*model.Homework, error) {
	if hwID == uuid.Nil {
		return nil, BadRequest("homework ID is required")
	}

	hw, err := s.homeworkRepo.GetByID(ctx, hwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}

	return hw, nil
}

func (s *AdminHomeworkService) ListHomework(ctx context.Context, courseID uuid.UUID) ([]model.Homework, error) {
	if courseID == uuid.Nil {
		return nil, BadRequest("course ID is required")
	}

	hws, err := s.homeworkRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, Internal("Failed to fetch homework list", err)
	}

	return hws, nil
}

func (s *AdminHomeworkService) UpdateHomework(ctx context.Context, hwID uuid.UUID, input UpdateHomeworkInput) (*model.Homework, error) {
	hw, err := s.GetHomework(ctx, hwID)
	if err != nil {
		return nil, err
	}

	if input.StartDate != "" {
		if !IsValidDate(input.StartDate) {
			return nil, BadRequest("start date must be in format YYYY-MM-DD")
		}
		hw.StartDate, err = parseDatePtr(input.StartDate)
		if err != nil {
			return nil, err
		}
	}
	if input.EndDate != "" {
		if !IsValidDate(input.EndDate) {
			return nil, BadRequest("end date must be in format YYYY-MM-DD")
		}
		hw.EndDate, err = parseDatePtr(input.EndDate)
		if err != nil {
			return nil, err
		}
	}

	if hw.StartDate != nil && hw.EndDate != nil && !hw.EndDate.After(*hw.StartDate) {
		return nil, BadRequest("end date must be after start_date")
	}

	if err := s.homeworkRepo.Update(ctx, hw); err != nil {
		return nil, Internal("Failed to update homework", err)
	}

	return hw, nil
}

func (s *AdminHomeworkService) DeleteHomework(ctx context.Context, hwID uuid.UUID) error {
	if hwID == uuid.Nil {
		return BadRequest("homework ID is required")
	}

	if _, err := s.GetHomework(ctx, hwID); err != nil {
		return err
	}

	if err := s.homeworkRepo.Delete(ctx, hwID); err != nil {
		return Internal("Failed to delete homework", err)
	}

	return nil
}

func (s *AdminHomeworkService) PublishHomework(ctx context.Context, hwID uuid.UUID, isPublic bool) (*model.Homework, error) {
	hw, err := s.GetHomework(ctx, hwID)
	if err != nil {
		return nil, err
	}

	hw.IsPublic = &isPublic

	if err := s.homeworkRepo.Update(ctx, hw); err != nil {
		return nil, Internal("Failed to publish homework", err)
	}

	return hw, nil
}

func (s *AdminHomeworkService) SetDeadline(ctx context.Context, input SetDeadlineInput) (*model.Deadline, error) {
	if input.CourseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	if input.HomeworkID == uuid.Nil {
		return nil, BadRequest("homework ID is required")
	}
	if input.Title == "" {
		return nil, BadRequest("title is required")
	}

	dueDate, err := time.Parse(time.RFC3339, input.DueDate)
	if err != nil {
		return nil, BadRequest("due date must be in RFC3339 format")
	}

	if _, err := s.GetHomework(ctx, input.HomeworkID); err != nil {
		return nil, err
	}

	deadline := &model.Deadline{
		Title:       input.Title,
		Description: stringPtr(input.Description),
		CourseID:    input.CourseID,
		DueDate:     dueDate,
		AssignedBy:  input.AssignedBy,
	}

	if err := s.deadlineRepo.Create(ctx, deadline); err != nil {
		return nil, Internal("Failed to set deadline", err)
	}

	return deadline, nil
}

func (s *AdminHomeworkService) UpdateDeadline(ctx context.Context, deadlineID uuid.UUID, input UpdateDeadlineInput) (*model.Deadline, error) {
	if deadlineID == uuid.Nil {
		return nil, BadRequest("deadline ID is required")
	}

	deadline, err := s.deadlineRepo.GetByID(ctx, deadlineID)
	if err != nil {
		return nil, NotFound("Deadline not found")
	}

	if input.Title != "" {
		deadline.Title = input.Title
	}
	if input.Description != "" {
		deadline.Description = stringPtr(input.Description)
	}
	if input.DueDate != "" {
		dueDate, err := time.Parse(time.RFC3339, input.DueDate)
		if err != nil {
			return nil, BadRequest("due date must be in RFC3339 format")
		}
		deadline.DueDate = dueDate
	}

	if err := s.deadlineRepo.Update(ctx, deadline); err != nil {
		return nil, Internal("Failed to update deadline", err)
	}

	return deadline, nil
}

func (s *AdminHomeworkService) DeleteDeadline(ctx context.Context, deadlineID uuid.UUID) error {
	if deadlineID == uuid.Nil {
		return BadRequest("deadline ID is required")
	}

	if _, err := s.deadlineRepo.GetByID(ctx, deadlineID); err != nil {
		return NotFound("Deadline not found")
	}

	if err := s.deadlineRepo.Delete(ctx, deadlineID); err != nil {
		return Internal("Failed to delete deadline", err)
	}

	return nil
}

func parseDatePtr(date string) (*time.Time, error) {
	if date == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, BadRequest("date must be in format YYYY-MM-DD")
	}
	return &parsed, nil
}
