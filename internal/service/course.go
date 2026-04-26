package service

import (
	"context"
	"time"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type CourseService struct {
	courseRepo repo.CourseRepositoryInterface
}

func NewCourseService(courseRepo repo.CourseRepositoryInterface) *CourseService {
	return &CourseService{courseRepo: courseRepo}
}

type CourseInput struct {
	Name         string
	Slug         string
	Status       string
	StartDate    string
	EndDate      string
	RepoTemplate string
	Description  string
}

func (s *CourseService) GetCourses(ctx context.Context, status string) ([]models.Course, error) {
	courses, err := s.courseRepo.GetCourses(ctx)
	if err != nil {
		return nil, Internal("Failed to get courses", err)
	}

	filtered := make([]models.Course, 0, len(courses))
	for _, course := range courses {
		if status == "" || course.Status == status {
			filtered = append(filtered, course)
		}
	}
	return filtered, nil
}

func (s *CourseService) GetCourse(ctx context.Context, courseID string) (*models.Course, error) {
	course, err := s.courseRepo.GetCourseByID(ctx, courseID)
	if err != nil {
		return nil, Internal("Failed to get course", err)
	}
	if course == nil {
		return nil, NotFound("course not found")
	}
	return course, nil
}

func (s *CourseService) CreateCourse(ctx context.Context, input CourseInput) (*models.Course, error) {
	if err := validateCreateCourse(input); err != nil {
		return nil, err
	}

	existing, err := s.courseRepo.GetCourseByID(ctx, input.Slug)
	if err != nil {
		return nil, Internal("Failed to check course uniqueness", err)
	}
	if existing != nil {
		return nil, Conflict("course with this slug already exists")
	}

	course := models.Course{
		ID:           input.Slug,
		Name:         input.Name,
		Status:       input.Status,
		StartDate:    input.StartDate,
		EndDate:      input.EndDate,
		RepoTemplate: input.RepoTemplate,
		Description:  input.Description,
		URL:          "/course/" + input.Slug,
	}

	return s.courseRepo.CreateCourse(ctx, course)
}

func (s *CourseService) UpdateCourse(ctx context.Context, courseID string, input CourseInput) (*models.Course, error) {
	course, err := s.GetCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if input.Status != "" && !IsValidCourseStatus(input.Status) {
		return nil, BadRequest("invalid status value")
	}
	if input.StartDate != "" && !IsValidDate(input.StartDate) {
		return nil, BadRequest("startDate must be in format YYYY-MM-DD")
	}
	if input.EndDate != "" && !IsValidDate(input.EndDate) {
		return nil, BadRequest("endDate must be in format YYYY-MM-DD")
	}

	updated := *course
	if input.Name != "" {
		updated.Name = input.Name
	}
	if input.Status != "" {
		updated.Status = input.Status
	}
	if input.StartDate != "" {
		updated.StartDate = input.StartDate
	}
	if input.EndDate != "" {
		updated.EndDate = input.EndDate
	}
	if input.RepoTemplate != "" {
		updated.RepoTemplate = input.RepoTemplate
	}
	if input.Description != "" {
		updated.Description = input.Description
	}

	if !IsValidDateRange(updated.StartDate, updated.EndDate) {
		return nil, BadRequest("endDate must be after startDate")
	}

	return s.courseRepo.UpdateCourse(ctx, courseID, updated)
}

func (s *CourseService) GetCourseBoard(ctx context.Context, courseID string) (*models.TaskBoardSummary, error) {
	if courseID == "" {
		return nil, BadRequest("course ID is required")
	}

	course, err := s.GetCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	board, ok, err := s.courseRepo.GetCourseBoard(ctx, courseID)
	if err != nil {
		return nil, Internal("Failed to get course board", err)
	}
	if ok {
		return board, nil
	}

	return &models.TaskBoardSummary{
		CourseName:   course.Name,
		CourseStatus: course.Status,
		Groups:       []models.BoardGroup{},
	}, nil
}

func validateCreateCourse(input CourseInput) error {
	if input.Name == "" {
		return BadRequest("name is required")
	}
	if input.Slug == "" {
		return BadRequest("slug is required")
	}
	if input.Status == "" {
		return BadRequest("status is required")
	}
	if !IsValidCourseStatus(input.Status) {
		return BadRequest("invalid status value")
	}
	if input.StartDate == "" {
		return BadRequest("startDate is required")
	}
	if !IsValidDate(input.StartDate) {
		return BadRequest("startDate must be in format YYYY-MM-DD")
	}
	if input.EndDate == "" {
		return BadRequest("endDate is required")
	}
	if !IsValidDate(input.EndDate) {
		return BadRequest("endDate must be in format YYYY-MM-DD")
	}
	if !IsValidDateRange(input.StartDate, input.EndDate) {
		return BadRequest("endDate must be after startDate")
	}
	if input.RepoTemplate == "" {
		return BadRequest("repoTemplate is required")
	}
	if input.Description == "" {
		return BadRequest("description is required")
	}
	return nil
}

func IsValidCourseStatus(status string) bool {
	valid := map[string]bool{
		"created":          true,
		"hidden":           true,
		"in_progress":      true,
		"all_tasks_issued": true,
		"doreshka":         true,
		"finished":         true,
	}
	return valid[status]
}

func IsValidDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

func IsValidDateRange(start, end string) bool {
	startDate, _ := time.Parse("2006-01-02", start)
	endDate, _ := time.Parse("2006-01-02", end)
	return endDate.After(startDate)
}
