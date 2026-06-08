package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	models "fcstask-backend/internal/db/model"
)

// Моки
type mockCourseRepo struct {
	mock.Mock
}

func (m *mockCourseRepo) GetCoursesByUserID(ctx context.Context, userID uuid.UUID, status string) ([]models.Course, error) {
	args := m.Called(ctx, userID, status)
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *mockCourseRepo) GetCourseByID(ctx context.Context, courseID string) (*models.Course, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *mockCourseRepo) CreateCourse(ctx context.Context, course models.Course) (*models.Course, error) {
	args := m.Called(ctx, course)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *mockCourseRepo) UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error) {
	args := m.Called(ctx, courseID, course)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *mockCourseRepo) DeleteCourse(ctx context.Context, courseID string) error { return nil }
func (m *mockCourseRepo) GetCourses(ctx context.Context) ([]models.Course, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Course), args.Error(1)
}
func (m *mockCourseRepo) GetCourseBoard(ctx context.Context, courseID string, userID uuid.UUID) (*models.TaskBoardSummary, bool, error) {
	args := m.Called(ctx, courseID, userID)

	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*models.TaskBoardSummary), args.Bool(1), args.Error(2)
}

func (m *mockCourseRepo) GetLeaderboard(ctx context.Context, courseID string) ([]models.LeaderboardEntry, error) {
	args := m.Called(ctx, courseID)
	return args.Get(0).([]models.LeaderboardEntry), args.Error(1)
}

func (m *mockCourseRepo) GetCourseInfo(ctx context.Context, courseID uuid.UUID) (*models.CourseInfo, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CourseInfo), args.Error(1)
}

func (m *mockCourseRepo) UpdateInviteCode(ctx context.Context, courseID uuid.UUID, code *string) error {
	args := m.Called(ctx, courseID, code)
	return args.Error(0)
}

func (m *mockCourseRepo) GetPublicCourses(ctx context.Context) ([]models.Course, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Course), args.Error(1)
}

type mockRoleRepo struct {
	mock.Mock
}

func (m *mockRoleRepo) AssignRoleWithPermissions(ctx context.Context, role *models.UserRole, permissions []string) error {
	args := m.Called(ctx, role, permissions)
	return args.Error(0)
}
func (m *mockRoleRepo) RevokeRoleWithPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	return nil
}
func (m *mockRoleRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]models.UserRole, error) {
	return nil, nil
}
func (m *mockRoleRepo) GetRoleIDByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID, courseID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}
func (m *mockRoleRepo) RoleBelongsToCourse(ctx context.Context, roleID, courseID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockRoleRepo) HasPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, roleID, permission)
	return args.Bool(0), args.Error(1)
}
func (m *mockRoleRepo) AddPermission(ctx context.Context, perm *models.CourseAdminPermission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}
func (m *mockRoleRepo) AddPermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	args := m.Called(ctx, roleID, permissions)
	return args.Error(0)
}
func (m *mockRoleRepo) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	return nil
}
func (m *mockRoleRepo) RemovePermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	args := m.Called(ctx, roleID, permissions)
	return args.Error(0)
}
func (m *mockRoleRepo) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]models.CourseAdminPermission, error) {
	return nil, nil
}

func setupService() (*CourseService, *mockCourseRepo, *mockRoleRepo) {
	cRepo := new(mockCourseRepo)
	rRepo := new(mockRoleRepo)
	svc := NewCourseService(cRepo, rRepo, nil)
	return svc, cRepo, rRepo
}

func validInput() CourseInput {
	return CourseInput{
		Name: "Test", Slug: "test", Status: "in_progress",
		StartDate: "2026-01-01", EndDate: "2026-02-01",
		RepoTemplate: "git@t", Description: "desc",
	}
}

// ==================== CreateCourse ====================

func TestCreateCourse_Public_NoInviteCode(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	superRoleID := uuid.New()
	input := validInput()
	input.Type = models.CourseTypePublic

	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(superRoleID, nil).Once()
	rRepo.On("HasPermission", mock.Anything, superRoleID, PermissionCourseCreate).Return(true, nil).Once()
	cRepo.On("GetCourseByID", mock.Anything, "test").Return(nil, nil)
	cRepo.On("CreateCourse", mock.Anything, mock.Anything).Return(&models.Course{
		ID: uuid.New(), Name: "Test", Slug: "test", Type: models.CourseTypePublic,
	}, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, mock.Anything).Return(uuid.Nil, gorm.ErrRecordNotFound)
	rRepo.On("AssignRoleWithPermissions", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	course, err := svc.CreateCourse(context.Background(), userID, input)

	assert.NoError(t, err)
	assert.Nil(t, course.InviteCode) // публичный — без кода
}

func TestCreateCourse_WithProvidedInviteCode(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	superRoleID := uuid.New()
	input := validInput()
	input.InviteCode = stringPtr("my-secret")

	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(superRoleID, nil).Once()
	rRepo.On("HasPermission", mock.Anything, superRoleID, PermissionCourseCreate).Return(true, nil).Once()
	cRepo.On("GetCourseByID", mock.Anything, "test").Return(nil, nil)
	cRepo.On("CreateCourse", mock.Anything, mock.MatchedBy(func(c models.Course) bool {
		return c.InviteCode != nil && *c.InviteCode == "my-secret"
	})).Return(&models.Course{ID: uuid.New(), Name: "Test", InviteCode: stringPtr("my-secret")}, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, mock.Anything).Return(uuid.Nil, gorm.ErrRecordNotFound)
	rRepo.On("AssignRoleWithPermissions", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	course, err := svc.CreateCourse(context.Background(), userID, input)

	assert.NoError(t, err)
	assert.Equal(t, "my-secret", *course.InviteCode)
}

func TestCreateCourse_Conflict(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	superRoleID := uuid.New()

	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(superRoleID, nil).Once()
	rRepo.On("HasPermission", mock.Anything, superRoleID, PermissionCourseCreate).Return(true, nil).Once()
	cRepo.On("GetCourseByID", mock.Anything, "test").Return(&models.Course{ID: uuid.New()}, nil)

	_, err := svc.CreateCourse(context.Background(), userID, validInput())

	assert.Error(t, err)
	svcErr := err.(*Error)
	assert.Equal(t, "conflict", svcErr.Code)
}

func TestCreateCourse_ValidationErrors(t *testing.T) {
	svc, _, rRepo := setupService()
	userID := uuid.New()
	superRoleID := uuid.New()
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(superRoleID, nil)
	rRepo.On("HasPermission", mock.Anything, superRoleID, PermissionCourseCreate).Return(true, nil)

	tests := []struct {
		name  string
		input CourseInput
	}{
		{"no name", func() CourseInput { i := validInput(); i.Name = ""; return i }()},
		{"no slug", func() CourseInput { i := validInput(); i.Slug = ""; return i }()},
		{"no status", func() CourseInput { i := validInput(); i.Status = ""; return i }()},
		{"invalid status", func() CourseInput { i := validInput(); i.Status = "bad"; return i }()},
		{"no startDate", func() CourseInput { i := validInput(); i.StartDate = ""; return i }()},
		{"invalid startDate", func() CourseInput { i := validInput(); i.StartDate = "01-01-2026"; return i }()},
		{"no endDate", func() CourseInput { i := validInput(); i.EndDate = ""; return i }()},
		{"end before start", func() CourseInput { i := validInput(); i.EndDate = "2025-01-01"; return i }()},
		{"no repoTemplate", func() CourseInput { i := validInput(); i.RepoTemplate = ""; return i }()},
		{"no description", func() CourseInput { i := validInput(); i.Description = ""; return i }()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateCourse(context.Background(), userID, tt.input)
			assert.Error(t, err)
			svcErr := err.(*Error)
			assert.Equal(t, "bad_request", svcErr.Code)
		})
	}
}

// ==================== UpdateCourse ====================

func TestUpdateCourse_Success(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	courseID := uuid.New().String()
	existing := &models.Course{
		ID: uuid.MustParse(courseID), Name: "Old", Slug: "old",
		Status: "in_progress", Type: models.CourseTypePrivate,
		StartDate: parseDate("2026-01-01"), EndDate: parseDate("2026-02-01"),
	}

	cRepo.On("GetCourseByID", mock.Anything, courseID).Return(existing, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, existing.ID).Return(uuid.New(), nil)
	rRepo.On("HasPermission", mock.Anything, mock.Anything, PermissionCourseUpdate).Return(true, nil)
	rRepo.On("HasPermission", mock.Anything, mock.Anything, PermissionCourseRead).Return(true, nil)
	cRepo.On("UpdateCourse", mock.Anything, courseID, mock.MatchedBy(func(c models.Course) bool {
		return c.Name == "Updated"
	})).Return(&models.Course{Name: "Updated"}, nil)

	course, err := svc.UpdateCourse(context.Background(), userID, courseID, CourseInput{Name: "Updated"})

	assert.NoError(t, err)
	assert.Equal(t, "Updated", course.Name)
}

func TestUpdateCourse_TypeChange_PrivateToPublic(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	inviteCode := "old-code"
	courseID := uuid.New().String()
	existing := &models.Course{
		ID: uuid.MustParse(courseID), Name: "Old", Type: models.CourseTypePrivate,
		InviteCode: &inviteCode,
		StartDate:  parseDate("2026-01-01"), EndDate: parseDate("2026-02-01"),
	}

	cRepo.On("GetCourseByID", mock.Anything, courseID).Return(existing, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, existing.ID).Return(uuid.New(), nil)
	rRepo.On("HasPermission", mock.Anything, mock.Anything, PermissionCourseUpdate).Return(true, nil)
	rRepo.On("HasPermission", mock.Anything, mock.Anything, PermissionCourseRead).Return(true, nil)
	cRepo.On("UpdateCourse", mock.Anything, courseID, mock.MatchedBy(func(c models.Course) bool {
		return c.InviteCode == nil && c.Type == models.CourseTypePublic
	})).Return(&models.Course{Type: models.CourseTypePublic}, nil)

	course, err := svc.UpdateCourse(context.Background(), userID, courseID, CourseInput{Type: models.CourseTypePublic})

	assert.NoError(t, err)
	assert.Nil(t, course.InviteCode)
}

func TestUpdateCourse_NotFound(t *testing.T) {
	svc, cRepo, _ := setupService()

	cRepo.On("GetCourseByID", mock.Anything, "unknown").Return(nil, nil)

	_, err := svc.UpdateCourse(context.Background(), uuid.New(), "unknown", CourseInput{Name: "X"})

	svcErr := err.(*Error)
	assert.Equal(t, "not_found", svcErr.Code)
}

func TestUpdateCourse_Forbidden(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	courseID := uuid.New().String()
	existing := &models.Course{ID: uuid.MustParse(courseID), Type: models.CourseTypePrivate}

	cRepo.On("GetCourseByID", mock.Anything, courseID).Return(existing, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, existing.ID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	_, err := svc.UpdateCourse(context.Background(), userID, courseID, CourseInput{Name: "X"})

	svcErr := err.(*Error)
	assert.Equal(t, "forbidden", svcErr.Code)
}

func TestJoinCourse_Public_Success(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	courseID := uuid.New()
	course := &models.Course{ID: courseID, Type: models.CourseTypePublic}

	cRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, courseID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)
	rRepo.On("AssignRoleWithPermissions", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := svc.JoinCourse(context.Background(), userID, courseID.String(), "")

	assert.NoError(t, err)
}

func TestJoinCourse_Private_WrongCode(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	courseID := uuid.New()
	secret := "secret"
	course := &models.Course{ID: courseID, Type: models.CourseTypePrivate, InviteCode: &secret}

	cRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, courseID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	err := svc.JoinCourse(context.Background(), userID, courseID.String(), "wrong")

	svcErr := err.(*Error)
	assert.Equal(t, "forbidden", svcErr.Code)
}

func TestJoinCourse_AlreadyParticipant(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	courseID := uuid.New()
	course := &models.Course{ID: courseID, Type: models.CourseTypePublic}

	cRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	rRepo.On("HasPermission", mock.Anything, mock.Anything, PermissionCourseRead).Return(true, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, courseID).Return(uuid.New(), nil)

	err := svc.JoinCourse(context.Background(), userID, courseID.String(), "")

	svcErr := err.(*Error)
	assert.Equal(t, "conflict", svcErr.Code)
}

func TestJoinCourse_Finished_Forbidden(t *testing.T) {
	svc, cRepo, _ := setupService()
	userID := uuid.New()
	courseID := uuid.New()
	course := &models.Course{ID: courseID, Type: models.CourseTypePublic, Status: "finished"}

	cRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)

	err := svc.JoinCourse(context.Background(), userID, courseID.String(), "")

	svcErr := err.(*Error)
	assert.Equal(t, "forbidden", svcErr.Code)
}

func TestJoinCourse_Hidden_NotFound(t *testing.T) {
	svc, cRepo, _ := setupService()
	userID := uuid.New()
	courseID := uuid.New()
	course := &models.Course{ID: courseID, Type: models.CourseTypePublic, Status: "hidden"}

	cRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)

	err := svc.JoinCourse(context.Background(), userID, courseID.String(), "")

	svcErr := err.(*Error)
	assert.Equal(t, "not_found", svcErr.Code)
}

func TestGetCourse_Public_NoPermissionCheck(t *testing.T) {
	svc, cRepo, _ := setupService()
	userID := uuid.Nil // неавторизованный
	course := &models.Course{ID: uuid.New(), Name: "Pub", Type: models.CourseTypePublic}

	cRepo.On("GetCourseByID", mock.Anything, "pub").Return(course, nil)

	result, err := svc.GetCourse(context.Background(), userID, "pub")
	assert.NoError(t, err)
	assert.Equal(t, "Pub", result.Name)
}

func TestGetCourse_Private_HasPermission(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	roleID := uuid.New()
	course := &models.Course{ID: uuid.New(), Name: "Priv", Type: models.CourseTypePrivate}

	cRepo.On("GetCourseByID", mock.Anything, "priv").Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, course.ID).Return(roleID, nil)
	rRepo.On("HasPermission", mock.Anything, roleID, PermissionCourseRead).Return(true, nil)

	result, err := svc.GetCourse(context.Background(), userID, "priv")
	assert.NoError(t, err)
	assert.Equal(t, "Priv", result.Name)
}

func TestGetCourse_Private_NoPermission(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	course := &models.Course{ID: uuid.New(), Name: "Priv", Type: models.CourseTypePrivate}

	cRepo.On("GetCourseByID", mock.Anything, "priv").Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, course.ID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	_, err := svc.GetCourse(context.Background(), userID, "priv")
	svcErr := err.(*Error)
	assert.Equal(t, "forbidden", svcErr.Code)
}

func TestGetCourse_Hidden_NoAccess(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	course := &models.Course{ID: uuid.New(), Name: "Hidden", Status: "hidden", Type: models.CourseTypePublic}

	cRepo.On("GetCourseByID", mock.Anything, "hidden").Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, course.ID).Return(uuid.New(), nil)
	rRepo.On("HasPermission", mock.Anything, mock.Anything, PermissionCourseHiddenRead).Return(false, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	_, err := svc.GetCourse(context.Background(), userID, "hidden")

	svcErr := err.(*Error)
	assert.Equal(t, "not_found", svcErr.Code)
}

func TestGetCourse_Hidden_AdminVisible(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	roleID := uuid.New()
	course := &models.Course{ID: uuid.New(), Name: "Hidden", Status: "hidden", Type: models.CourseTypePublic}

	cRepo.On("GetCourseByID", mock.Anything, "hidden").Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, course.ID).Return(roleID, nil)
	rRepo.On("HasPermission", mock.Anything, roleID, PermissionCourseHiddenRead).Return(true, nil)

	result, err := svc.GetCourse(context.Background(), userID, "hidden")
	assert.NoError(t, err)
	assert.Equal(t, "Hidden", result.Name)
}

func TestGetCourses_Success(t *testing.T) {
	svc, cRepo, _ := setupService()
	userID := uuid.New()
	courses := []models.Course{
		{ID: uuid.New(), Name: "Visible", Status: "in_progress"},
		{ID: uuid.New(), Name: "Hidden", Status: "hidden"},
	}

	cRepo.On("GetCoursesByUserID", mock.Anything, userID, "").Return(courses, nil)

	result, err := svc.GetCourses(context.Background(), userID, "")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, courses, result)
}

func TestGetPublicCourses_Success(t *testing.T) {
	svc, cRepo, _ := setupService()
	cRepo.On("GetPublicCourses", mock.Anything).Return([]models.Course{
		{ID: uuid.New(), Name: "Visible", Status: "in_progress", Type: models.CourseTypePublic},
	}, nil)

	result, err := svc.GetPublicCourses(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Visible", result[0].Name)
}

func TestGetCourseBoard_Success(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	roleID := uuid.New()
	courseID := uuid.New().String()
	course := &models.Course{ID: uuid.MustParse(courseID), Name: "Go", Status: "in_progress"}

	cRepo.On("GetCourseByID", mock.Anything, courseID).Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, course.ID).Return(roleID, nil)
	rRepo.On("HasPermission", mock.Anything, roleID, PermissionCourseRead).Return(true, nil)
	cRepo.On("GetCourseBoard", mock.Anything, courseID, userID).Return(nil, false, nil)

	board, err := svc.GetCourseBoard(context.Background(), userID, courseID)

	assert.NoError(t, err)
	assert.Equal(t, "Go", board.CourseName)
}

func TestGetCourseBoard_NotFound(t *testing.T) {
	svc, cRepo, _ := setupService()
	userID := uuid.New()

	cRepo.On("GetCourseByID", mock.Anything, "unknown").Return(nil, nil)

	_, err := svc.GetCourseBoard(context.Background(), userID, "unknown")

	svcErr := err.(*Error)
	assert.Equal(t, "not_found", svcErr.Code)
}

func TestGetCourseBoard_Forbidden(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	courseID := uuid.New().String()
	course := &models.Course{ID: uuid.MustParse(courseID), Type: models.CourseTypePrivate}

	cRepo.On("GetCourseByID", mock.Anything, courseID).Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, course.ID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	_, err := svc.GetCourseBoard(context.Background(), userID, courseID)
	svcErr := err.(*Error)
	assert.Equal(t, "forbidden", svcErr.Code)
}

func TestGetLeaderboard_Success(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	roleID := uuid.New()
	courseID := uuid.New()
	course := &models.Course{ID: courseID, Name: "Go", Type: models.CourseTypePrivate}

	hwID := uuid.New()
	task1 := uuid.New()
	task2 := uuid.New()
	entries := []models.LeaderboardEntry{
		{
			Username:   "alice",
			TotalScore: 30,
			Homeworks: []models.HomeworkScore{
				{
					HomeworkID:    hwID,
					HomeworkTitle: "Week 1",
					TotalScore:    30,
					Tasks: []models.TaskScore{
						{TaskID: task1, Title: "Task 1", Score: 10},
						{TaskID: task2, Title: "Task 2", Score: 20},
					},
				},
			},
			Rank: 1,
		},
		{
			Username:   "bob",
			TotalScore: 20,
			Homeworks: []models.HomeworkScore{
				{
					HomeworkID:    hwID,
					HomeworkTitle: "Week 1",
					TotalScore:    20,
					Tasks: []models.TaskScore{
						{TaskID: task1, Title: "Task 1", Score: 20},
					},
				},
			},
			Rank: 2,
		},
	}

	cRepo.On("GetCourseByID", mock.Anything, course.ID.String()).Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, course.ID).Return(roleID, nil)
	rRepo.On("HasPermission", mock.Anything, roleID, PermissionCourseRead).Return(true, nil)
	rRepo.On("HasPermission", mock.Anything, roleID, PermissionLeaderboardRead).Return(true, nil)
	cRepo.On("GetLeaderboard", mock.Anything, courseID.String()).Return(entries, nil)

	result, err := svc.GetLeaderboard(context.Background(), userID, courseID.String())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "alice", result[0].Username)
	assert.Equal(t, 30, result[0].TotalScore)
	assert.Equal(t, 1, result[0].Rank)
	assert.Len(t, result[0].Homeworks, 1)
	assert.Equal(t, "Week 1", result[0].Homeworks[0].HomeworkTitle)
	assert.Equal(t, 30, result[0].Homeworks[0].TotalScore)
	assert.Len(t, result[0].Homeworks[0].Tasks, 2)
	assert.Equal(t, "Task 1", result[0].Homeworks[0].Tasks[0].Title)
	assert.Equal(t, 10, result[0].Homeworks[0].Tasks[0].Score)
}

func TestGetLeaderboard_Empty(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	roleID := uuid.New()
	courseID := uuid.New()
	course := &models.Course{ID: courseID, Type: models.CourseTypePrivate}

	cRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, courseID).Return(roleID, nil)
	rRepo.On("HasPermission", mock.Anything, roleID, PermissionCourseRead).Return(true, nil)
	rRepo.On("HasPermission", mock.Anything, roleID, PermissionLeaderboardRead).Return(true, nil)
	cRepo.On("GetLeaderboard", mock.Anything, courseID.String()).Return([]models.LeaderboardEntry{}, nil)

	result, err := svc.GetLeaderboard(context.Background(), userID, courseID.String())

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetLeaderboard_Forbidden(t *testing.T) {
	svc, cRepo, rRepo := setupService()
	userID := uuid.New()
	courseID := uuid.New()
	course := &models.Course{ID: courseID, Type: models.CourseTypePrivate}

	cRepo.On("GetCourseByID", mock.Anything, courseID.String()).Return(course, nil)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, courseID).Return(uuid.Nil, gorm.ErrRecordNotFound)
	rRepo.On("GetRoleIDByUserAndCourse", mock.Anything, userID, uuid.Nil).Return(uuid.Nil, gorm.ErrRecordNotFound)

	_, err := svc.GetLeaderboard(context.Background(), userID, courseID.String())

	svcErr := err.(*Error)
	assert.Equal(t, "forbidden", svcErr.Code)
}

func TestIsValidCourseStatus(t *testing.T) {
	assert.True(t, IsValidCourseStatus("finished"))
	assert.False(t, IsValidCourseStatus("invalid"))
}

func TestIsValidCourseType(t *testing.T) {
	assert.True(t, IsValidCourseType(models.CourseTypePublic))
	assert.True(t, IsValidCourseType(models.CourseTypePrivate))
	assert.False(t, IsValidCourseType("invalid"))
}

func TestIsValidDate(t *testing.T) {
	assert.True(t, IsValidDate("2026-01-01"))
	assert.False(t, IsValidDate("01-01-2026"))
}

func TestIsValidDateRange(t *testing.T) {
	assert.True(t, IsValidDateRange("2026-01-01", "2026-02-01"))
	assert.False(t, IsValidDateRange("2026-02-01", "2026-01-01"))
	assert.False(t, IsValidDateRange("2026-01-01", "2026-01-01"))
}

func parseDate(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}
