package service

import (
	"context"
	"time"
	"github.com/google/uuid"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type CourseService struct {
	courseRepo repo.CourseRepositoryInterface
	roleRepo   repo.IRoleRepo
}

func NewCourseService(courseRepo repo.CourseRepositoryInterface, roleRepo repo.IRoleRepo) *CourseService {
	return &CourseService{courseRepo: courseRepo, roleRepo: roleRepo}
}

type CourseInput struct {
	Name         string
	Slug         string
	Status       string
	Type         models.CourseType
	StartDate    string
	EndDate      string
	RepoTemplate string
	Description  string
}

func (s *CourseService) GetCourses(ctx context.Context, userID uuid.UUID, status string) ([]models.Course, error) {
	courses, err := s.courseRepo.GetCoursesByUserID(ctx, userID, status)
	if err != nil {
		return nil, Internal("Failed to get courses", err)
	}

	return courses, nil
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
func (s *CourseService) CreateCourse(ctx context.Context, userID uuid.UUID, input CourseInput) (*models.Course, error) {
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
		Name:         input.Name,
		Slug:         input.Slug,
		Status:       input.Status,
		Type:         courseTypeOrDefault(input.Type),
		StartDate:    parseCourseDate(input.StartDate),
		EndDate:      parseCourseDate(input.EndDate),
		RepoTemplate: stringPtr(input.RepoTemplate),
		Description:  stringPtr(input.Description),
		URL:          "/course/" + input.Slug,
	}

	created, err := s.courseRepo.CreateCourse(ctx, course)
	if err != nil {
		return nil, Internal("Failed to create course", err)
	}

	roleID := uuid.New()
	userRole := &models.UserRole{
		UserID:   userID,
		CourseID: created.ID,
		RoleID:   roleID,
	}
	if err := s.roleRepo.AssignRole(ctx, userRole); err != nil {
		return nil, Internal("Failed to assign creator role", err)
	}

	adminPerm := &models.CourseAdminPermission{
		RoleID:     roleID,
		Permission: "admin",
	}
	if err := s.roleRepo.AddPermission(ctx, adminPerm); err != nil {
		return nil, Internal("Failed to assign admin permission", err)
	}

	return created, nil
}

func (s *CourseService) UpdateCourse(ctx context.Context, userID uuid.UUID, courseID string, input CourseInput) (*models.Course, error) {
	course, err := s.GetCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	courseUUID, _ := uuid.Parse(courseID)
	isAdmin, err := IsCourseAdmin(ctx, s.roleRepo, userID, courseUUID)
	if err != nil {
		return nil, Internal("Failed to check permissions", err)
	}
	if !isAdmin {
		return nil, Forbidden("You don't have permission to update this course")
	}
	if input.Status != "" && !IsValidCourseStatus(input.Status) {
		return nil, BadRequest("invalid status value")
	}
	if input.Type != "" && !IsValidCourseType(input.Type) {
		return nil, BadRequest("type must be 'public' or 'private'")
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
	if input.Type != "" {
		updated.Type = input.Type
	}
	if input.StartDate != "" {
		updated.StartDate = parseCourseDate(input.StartDate)
	}
	if input.EndDate != "" {
		updated.EndDate = parseCourseDate(input.EndDate)
	}
	if input.RepoTemplate != "" {
		updated.RepoTemplate = stringPtr(input.RepoTemplate)
	}
	if input.Description != "" {
		updated.Description = stringPtr(input.Description)
	}

	if !IsValidDateRange(formatCourseDate(updated.StartDate), formatCourseDate(updated.EndDate)) {
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
	if input.Type != "" && !IsValidCourseType(input.Type) {
		return BadRequest("type must be 'public' or 'private'")
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

func IsValidCourseType(courseType models.CourseType) bool {
	return courseType == models.CourseTypePublic || courseType == models.CourseTypePrivate
}

func courseTypeOrDefault(courseType models.CourseType) models.CourseType {
	if courseType == "" {
		return models.CourseTypePrivate
	}
	return courseType
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

func parseCourseDate(date string) *time.Time {
	parsed, _ := time.Parse("2006-01-02", date)
	return &parsed
}

func formatCourseDate(date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format("2006-01-02")
}

func stringPtr(value string) *string {
	return &value
}
