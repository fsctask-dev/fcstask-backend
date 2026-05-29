package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type CourseService struct {
	CourseRepo       repo.CourseRepositoryInterface
	RoleRepo         repo.IRoleRepo
	StudentScoreRepo repo.IStudentTaskScoreRepo
}

func NewCourseService(courseRepo repo.CourseRepositoryInterface, roleRepo repo.IRoleRepo, studentScoreRepo repo.IStudentTaskScoreRepo) *CourseService {
	return &CourseService{CourseRepo: courseRepo, RoleRepo: roleRepo, StudentScoreRepo: studentScoreRepo}
}

type CourseInput struct {
	Name         string
	Slug         string
	Status       string
	Type         models.CourseType
	InviteCode   *string
	StartDate    string
	EndDate      string
	RepoTemplate string
	Description  string
}

func (s *CourseService) GetCourses(ctx context.Context, userID uuid.UUID, status string) ([]models.Course, error) {
	courses, err := s.CourseRepo.GetCoursesByUserID(ctx, userID, status)
	if err != nil {
		return nil, Internal("Failed to get courses", err)
	}
	return courses, nil
}

func (s *CourseService) GetCourse(ctx context.Context, courseID string) (*models.Course, error) {
	course, err := s.CourseRepo.GetCourseByID(ctx, courseID)
	if err != nil {
		return nil, Internal("Failed to get course", err)
	}
	if course == nil {
		return nil, NotFound("course not found")
	}
	return course, nil
}

func (s *CourseService) CanReadCourse(ctx context.Context, userID uuid.UUID, course *models.Course) (bool, error) {
	if course == nil {
		return false, BadRequest("course is required")
	}

	if course.Type == models.CourseTypePublic {
		return true, nil
	}

	return HasScopedPermission(ctx, s.RoleRepo, userID, course.ID, PermissionHomeworkRead)
}

func (s *CourseService) CreateCourse(ctx context.Context, userID uuid.UUID, input CourseInput) (*models.Course, error) {
	if err := RequireScopedPermission(ctx, s.RoleRepo, userID, uuid.Nil, PermissionCourseCreate); err != nil {
		return nil, err
	}

	if err := validateCreateCourse(input); err != nil {
		return nil, err
	}

	existing, err := s.CourseRepo.GetCourseByID(ctx, input.Slug)
	if err != nil {
		return nil, Internal("Failed to check course uniqueness", err)
	}
	if existing != nil {
		return nil, Conflict("course with this slug already exists")
	}

	courseType := courseTypeOrDefault(input.Type)

	course := models.Course{
		Name:         input.Name,
		Slug:         input.Slug,
		Status:       input.Status,
		Type:         courseType,
		InviteCode:   inviteCodeForCourse(courseType, input.InviteCode),
		StartDate:    parseCourseDate(input.StartDate),
		EndDate:      parseCourseDate(input.EndDate),
		RepoTemplate: stringPtr(input.RepoTemplate),
		Description:  stringPtr(input.Description),
		URL:          "/course/" + input.Slug,
	}

	created, err := s.CourseRepo.CreateCourse(ctx, course)
	if err != nil {
		return nil, Internal("Failed to create course", err)
	}

	if _, err := EnsureUserRoleWithPermissions(ctx, s.RoleRepo, userID, created.ID, CourseAdminPermissions()); err != nil {
		return nil, Internal("Failed to assign admin permissions", err)
	}

	return created, nil
}

func (s *CourseService) UpdateCourse(ctx context.Context, userID uuid.UUID, courseID string, input CourseInput) (*models.Course, error) {
	course, err := s.GetCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if err := RequireScopedPermission(ctx, s.RoleRepo, userID, course.ID, PermissionHomeworkUpdate); err != nil {
		return nil, err
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
	switch {
	case updated.Type == models.CourseTypePublic:
		updated.InviteCode = nil
	case updated.Type == models.CourseTypePrivate && (input.Type != "" || input.InviteCode != nil):
		updated.InviteCode = inviteCodeForCourse(models.CourseTypePrivate, input.InviteCode)
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

	return s.CourseRepo.UpdateCourse(ctx, courseID, updated)
}

func (s *CourseService) GetCourseBoard(ctx context.Context, courseID string) (*models.TaskBoardSummary, error) {
	if courseID == "" {
		return nil, BadRequest("course ID is required")
	}

	course, err := s.GetCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	board, ok, err := s.CourseRepo.GetCourseBoard(ctx, courseID)
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

func (s *CourseService) JoinCourse(ctx context.Context, userID uuid.UUID, courseID string, code string) error {
	course, err := s.CourseRepo.GetCourseByID(ctx, courseID)
	if err != nil {
		return Internal("Failed to get course by ID", err)
	}
	if course == nil {
		return NotFound("course not found")
	}

	already, err := HasScopedPermission(ctx, s.RoleRepo, userID, course.ID, PermissionHomeworkRead)
	if err != nil {
		return Internal("Failed to check participation", err)
	}
	if already {
		return Conflict("already a participant")
	}

	if course.Type == models.CourseTypePublic {
		_, err := EnsureUserRoleWithPermissions(ctx, s.RoleRepo, userID, course.ID, CourseStudentPermissions())
		if err != nil {
			return Internal("Failed to join course", err)
		}
		return nil
	}
	if course.InviteCode == nil {
		return BadRequest("course has no invite code")
	}
	if *course.InviteCode != code {
		return Forbidden("invalid invite code")
	}

	_, err = EnsureUserRoleWithPermissions(ctx, s.RoleRepo, userID, course.ID, CourseStudentPermissions())
	if err != nil {
		return Internal("Failed to join course", err)
	}

	return nil
}

func (s *CourseService) GetLeaderboard(ctx context.Context, userID uuid.UUID, courseID string) ([]models.LeaderboardEntry, error) {
	if courseID == "" {
		return nil, BadRequest("course_id is required")
	}
	course, err := s.GetCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	// Проверяем, что у пользователя есть права на чтение leaderboard
	if err := RequireScopedPermission(ctx, s.RoleRepo, userID, course.ID, PermissionLeaderboardRead); err != nil {
		return nil, err
	}
	entries, err := s.CourseRepo.GetLeaderboard(ctx, course.ID)
	if err != nil {
		return nil, Internal("Failed to get leaderboard", err)
	}
	return entries, nil
}

func generateInviteCode() string {
	code := make([]byte, 6)
	_, _ = rand.Read(code)
	return hex.EncodeToString(code)
}

func inviteCodeForCourse(courseType models.CourseType, inviteCode *string) *string {
	if courseType != models.CourseTypePrivate {
		return nil
	}
	if inviteCode != nil {
		return inviteCode
	}
	code := generateInviteCode()
	return &code
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
