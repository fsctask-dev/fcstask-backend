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

func (r *controllerCourseRepo) GetCourseByID(ctx context.Context, courseID string) (*models.Course, error) {
	course, ok := r.courses[courseID]
	if !ok {
		return nil, nil
	}
	return &course, nil
}

func (r *controllerCourseRepo) CreateCourse(ctx context.Context, course models.Course) (*models.Course, error) {
	r.courses[course.ID] = course
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

func (r *controllerCourseRepo) GetCourseBoard(ctx context.Context, courseID string) (*models.TaskBoardSummary, bool, error) {
	return nil, false, nil
}

func newTestController(userRepo *controllerUserRepo, sessionRepo *controllerSessionRepo, courseRepo *controllerCourseRepo) *APIController {
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, sessionRepo)
	sessionService := service.NewSessionService(sessionRepo)
	courseService := service.NewCourseService(courseRepo)

	return NewAPIController(
		handler.NewAuthHandler(authService),
		handler.NewUserHandler(userService),
		handler.NewSessionHandler(sessionService, userService),
		handler.NewCourseHandler(courseService),
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
	controller := newTestController(
		&controllerUserRepo{},
		&controllerSessionRepo{},
		&controllerCourseRepo{courses: map[string]models.Course{
			"go": {
				ID:           "go",
				Name:         "Go",
				Status:       "created",
				StartDate:    "2026-01-01",
				EndDate:      "2026-02-01",
				RepoTemplate: "git@test/go.git",
				Description:  "Go course",
				URL:          "/course/go",
			},
		}},
	)
	controller.RegisterCourseRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/api/courses/go", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var course models.Course
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &course))
	assert.Equal(t, "go", course.ID)
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
