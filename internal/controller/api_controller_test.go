package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fcstask-backend/internal/api"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/handler"
	"fcstask-backend/internal/service"
)

type controllerUserRepo struct {
	user *models.User
}

func (r *controllerUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	user.ID = uuid.New()
	r.user = user
	return nil
}

func (r *controllerUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *controllerUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *controllerUserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *controllerUserRepo) GetUserByUserID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *controllerUserRepo) GetUserByTgUID(ctx context.Context, tgUID int64) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *controllerUserRepo) UpdateUser(ctx context.Context, user *models.User) error {
	r.user = user
	return nil
}

func (r *controllerUserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	r.user = nil
	return nil
}

func (r *controllerUserRepo) GetUsersWithSessions(ctx context.Context, limit, offset int) ([]models.User, error) {
	return []models.User{}, nil
}

func (r *controllerUserRepo) CountUsersWithSessions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *controllerUserRepo) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (r *controllerUserRepo) ExistsUserByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (r *controllerUserRepo) CountUsers(ctx context.Context) (int64, error) {
	return 0, nil
}

type controllerSessionRepo struct {
	session *models.Session
}

func (r *controllerSessionRepo) CreateSession(ctx context.Context, session *models.Session) error {
	session.ID = uuid.New()
	r.session = session
	return nil
}

func (r *controllerSessionRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	if r.session == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.session, nil
}

func (r *controllerSessionRepo) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	return nil, nil
}

func (r *controllerSessionRepo) GetSessionsWithUser(ctx context.Context, limit, offset int) ([]models.Session, error) {
	return nil, nil
}

func (r *controllerSessionRepo) CountSessions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *controllerSessionRepo) TouchSessionAccessedAt(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *controllerSessionRepo) DeleteSession(ctx context.Context, id uuid.UUID) error {
	r.session = nil
	return nil
}

func (r *controllerSessionRepo) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (r *controllerSessionRepo) CleanOutdatedSessions(ctx context.Context, ttl time.Duration) (int64, error) {
	return 0, nil
}

type controllerCourseRepo struct {
	courses map[string]models.Course
}

func (r *controllerCourseRepo) GetCourses(ctx context.Context) ([]models.Course, error) {
	courses := make([]models.Course, 0, len(r.courses))
	for _, course := range r.courses {
		courses = append(courses, course)
	}
	return courses, nil
}

func (r *controllerCourseRepo) GetCoursesByUserID(ctx context.Context, userID uuid.UUID, status string) ([]models.Course, error) {
	return r.GetCourses(ctx)
}

func (r *controllerCourseRepo) GetCourseByID(ctx context.Context, courseID string) (*models.Course, error) {
	course, ok := r.courses[courseID]
	if !ok {
		return nil, nil
	}
	return &course, nil
}

func (r *controllerCourseRepo) CreateCourse(ctx context.Context, course models.Course) (*models.Course, error) {
	if course.ID == uuid.Nil {
		course.ID = uuid.New()
	}
	r.courses[course.Slug] = course
	return &course, nil
}

func (r *controllerCourseRepo) UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error) {
	r.courses[courseID] = course
	return &course, nil
}

func (r *controllerCourseRepo) DeleteCourse(ctx context.Context, courseID string) error {
	delete(r.courses, courseID)
	return nil
}

func (r *controllerCourseRepo) GetCourseBoard(ctx context.Context, courseID string, userID uuid.UUID) (*models.TaskBoardSummary, bool, error) {
	return nil, false, nil
}

func (r *controllerCourseRepo) GetLeaderboard(ctx context.Context, courseID string) ([]models.LeaderboardEntry, error) {
	return nil, nil
}

func (r *controllerCourseRepo) UpdateInviteCode(ctx context.Context, courseID uuid.UUID, code *string) error {
	return nil
}

func (r *controllerCourseRepo) GetCourseInfo(ctx context.Context, courseID uuid.UUID) (*models.CourseInfo, error) {
	return nil, nil
}

func (r *controllerCourseRepo) GetPublicCourses(ctx context.Context) ([]models.Course, error) {
	var courses []models.Course
	for _, course := range r.courses {
		if course.Type == models.CourseTypePublic {
			courses = append(courses, course)
		}
	}
	return courses, nil
}

type controllerRoleRepo struct{}

func (r *controllerRoleRepo) AssignRoleWithPermissions(ctx context.Context, role *models.UserRole, permissions []string) error {
	return nil
}
func (r *controllerRoleRepo) RevokeRoleWithPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	return nil
}
func (r *controllerRoleRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]models.UserRole, error) {
	return nil, nil
}
func (r *controllerRoleRepo) GetRoleIDByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (r *controllerRoleRepo) RoleBelongsToCourse(ctx context.Context, roleID, courseID uuid.UUID) (bool, error) {
	return true, nil
}
func (r *controllerRoleRepo) HasPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	return true, nil
}
func (r *controllerRoleRepo) AddPermission(ctx context.Context, perm *models.CourseAdminPermission) error {
	return nil
}
func (r *controllerRoleRepo) AddPermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	return nil
}
func (r *controllerRoleRepo) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	return nil
}
func (r *controllerRoleRepo) RemovePermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	return nil
}
func (r *controllerRoleRepo) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]models.CourseAdminPermission, error) {
	return nil, nil
}

type controllerStatsRepo struct{}

func (r *controllerStatsRepo) GetStats(ctx context.Context) (*models.PlatformStats, error) {
	return &models.PlatformStats{
		TotalCourses:   1,
		PublicCourses:  1,
		PrivateCourses: 0,
		TotalUsers:     1,
	}, nil
}

func newTestController(userRepo *controllerUserRepo, sessionRepo *controllerSessionRepo, courseRepo *controllerCourseRepo) *APIController {
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, sessionRepo)
	sessionService := service.NewSessionService(sessionRepo)
	roleRepo := &controllerRoleRepo{}
	courseService := service.NewCourseService(courseRepo, roleRepo, nil)
	statsService := service.NewStatsService(&controllerStatsRepo{}, roleRepo)
	statsHandler := handler.NewStatsHandler(statsService)

	return NewAPIController(
		handler.NewAuthHandler(authService),
		handler.NewUserHandler(userService),
		handler.NewSessionHandler(sessionService, userService),
		handler.NewCourseHandler(courseService),
		nil,
		statsHandler,
	)
}

func TestAPIController_SignUp(t *testing.T) {
	e := echo.New()
	userRepo := &controllerUserRepo{}
	sessionRepo := &controllerSessionRepo{}
	controller := newTestController(userRepo, sessionRepo, &controllerCourseRepo{courses: map[string]models.Course{}})

	body := []byte(`{"email":"new@example.com","username":"newuser","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := controller.SignUp(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "new@example.com", userRepo.user.Email)
	assert.Equal(t, userRepo.user.ID, sessionRepo.session.UserID)
}

func TestAPIController_GetUserByID(t *testing.T) {
	e := echo.New()
	userID := uuid.New()
	userRepo := &controllerUserRepo{user: &models.User{
		ID:       userID,
		Email:    "user@example.com",
		Username: "user",
		UserID:   uuid.New(),
	}}
	controller := newTestController(userRepo, &controllerSessionRepo{}, &controllerCourseRepo{courses: map[string]models.Course{}})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/"+userID.String(), nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := controller.GetUserByID(ctx, openapi_types.UUID(userID))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.User
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, userID, uuid.UUID(resp.Id))
}

func TestAPIController_RegisterCourseRoutes(t *testing.T) {
	e := echo.New()
	userRepo := &controllerUserRepo{
		user: &models.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			Username: "testuser",
			UserID:   uuid.New(),
		},
	}
	controller := newTestController(
		userRepo,
		&controllerSessionRepo{},
		&controllerCourseRepo{courses: map[string]models.Course{
			"go": {
				ID:           uuid.New(),
				Name:         "Go",
				Slug:         "go",
				Status:       "created",
				Type:         models.CourseTypePublic,
				StartDate:    controllerCourseDate("2026-01-01"),
				EndDate:      controllerCourseDate("2026-02-01"),
				RepoTemplate: controllerStringPtr("git@test/go.git"),
				Description:  controllerStringPtr("Go course"),
				URL:          "/course/go",
			},
		}},
	)
	controller.RegisterCourseRoutes(e)

	// Middleware для подстановки пользователя
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", userRepo.user)
			return next(c)
		}
	})

	// Тестируем GET /api/courses вместо GET /api/courses/:courseId
	req := httptest.NewRequest(http.MethodGet, "/api/courses", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var courses []models.Course
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &courses))
	assert.Len(t, courses, 1)
	assert.Equal(t, "go", courses[0].Slug)
}

func TestAPIController_RegisterAdminRoutes_CreateCourse(t *testing.T) {
	e := echo.New()
	userRepo := &controllerUserRepo{
		user: &models.User{
			ID:       uuid.New(),
			Email:    "admin@example.com",
			Username: "admin",
			UserID:   uuid.New(),
		},
	}
	controller := newTestController(
		userRepo,
		&controllerSessionRepo{},
		&controllerCourseRepo{courses: map[string]models.Course{}},
	)
	controller.RegisterAdminRoutes(e, nil, nil, nil)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", userRepo.user)
			return next(c)
		}
	})

	body := []byte(`{"name":"New","slug":"new","status":"created","startDate":"2026-01-01","endDate":"2026-02-01","repoTemplate":"git@test/new.git","description":"desc"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/courses/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestAPIController_RegisterAdminRoutes_RegenerateInviteCode(t *testing.T) {
	e := echo.New()
	userRepo := &controllerUserRepo{
		user: &models.User{
			ID:       uuid.New(),
			Email:    "owner@example.com",
			Username: "owner",
			UserID:   uuid.New(),
		},
	}
	courseID := uuid.New()
	code := "secret"
	controller := newTestController(
		userRepo,
		&controllerSessionRepo{},
		&controllerCourseRepo{courses: map[string]models.Course{
			courseID.String(): {
				ID:         courseID,
				Name:       "Private",
				Slug:       "private",
				Status:     "created",
				Type:       models.CourseTypePrivate,
				InviteCode: &code,
			},
		}},
	)
	controller.RegisterAdminRoutes(e, nil, nil, nil)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", userRepo.user)
			return next(c)
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/admin/courses/"+courseID.String()+"/invite", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAPIController_SignIn(t *testing.T) {
	e := echo.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	userRepo := &controllerUserRepo{user: &models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Username:     "user",
		PasswordHash: string(hash),
		UserID:       uuid.New(),
	}}
	sessionRepo := &controllerSessionRepo{}
	controller := newTestController(userRepo, sessionRepo, &controllerCourseRepo{courses: map[string]models.Course{}})

	body := []byte(`{"email":"user@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/signin", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err = controller.SignIn(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, userRepo.user.ID, sessionRepo.session.UserID)
}

func controllerCourseDate(value string) *time.Time {
	parsed, _ := time.Parse("2006-01-02", value)
	return &parsed
}

func controllerStringPtr(value string) *string {
	return &value
}
