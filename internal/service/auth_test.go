package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	models "fcstask-backend/internal/db/model"
)

type authUserRepo struct {
	user          *models.User
	emailTaken    bool
	usernameTaken bool
	createErr     error
	findErr       error
}

func (r *authUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	user.ID = uuid.New()
	r.user = user
	return nil
}

func (r *authUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *authUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *authUserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *authUserRepo) GetUserByUserID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *authUserRepo) GetUserByTgUID(ctx context.Context, tgUID int64) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *authUserRepo) UpdateUser(ctx context.Context, user *models.User) error {
	r.user = user
	return nil
}

func (r *authUserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	r.user = nil
	return nil
}

func (r *authUserRepo) GetUsersWithSessions(ctx context.Context, limit, offset int) ([]models.User, error) {
	return nil, nil
}

func (r *authUserRepo) CountUsersWithSessions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *authUserRepo) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	return r.emailTaken, nil
}

func (r *authUserRepo) ExistsUserByUsername(ctx context.Context, username string) (bool, error) {
	return r.usernameTaken, nil
}

func (r *authUserRepo) CountUsers(ctx context.Context) (int64, error) {
	return 0, nil
}

type authSessionRepo struct {
	session   *models.Session
	createErr error
	deleteErr error
}

func (r *authSessionRepo) CreateSession(ctx context.Context, session *models.Session) error {
	if r.createErr != nil {
		return r.createErr
	}
	session.ID = uuid.New()
	r.session = session
	return nil
}

func (r *authSessionRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	if r.session == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.session, nil
}

func (r *authSessionRepo) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	return nil, nil
}

func (r *authSessionRepo) GetSessionsWithUser(ctx context.Context, limit, offset int) ([]models.Session, error) {
	return nil, nil
}

func (r *authSessionRepo) CountSessions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *authSessionRepo) TouchSessionAccessedAt(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *authSessionRepo) DeleteSession(ctx context.Context, id uuid.UUID) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.session = nil
	return nil
}

func (r *authSessionRepo) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (r *authSessionRepo) CleanOutdatedSessions(ctx context.Context, ttl time.Duration) (int64, error) {
	return 0, nil
}

func TestAuthService_SignUpSuccess(t *testing.T) {
	userRepo := &authUserRepo{}
	sessionRepo := &authSessionRepo{}
	svc := NewAuthService(userRepo, sessionRepo)

	result, err := svc.SignUp(context.Background(), SignUpInput{
		Email:     "new@example.com",
		Username:  "newuser",
		Password:  "secret",
		IP:        "127.0.0.1",
		UserAgent: "test-agent",
	})

	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", result.User.Email)
	assert.NotEmpty(t, result.User.PasswordHash)
	assert.NotEqual(t, uuid.Nil, result.Session.ID)
	assert.Equal(t, result.User.ID, result.Session.UserID)
}

func TestAuthService_SignUpUsernameConflict(t *testing.T) {
	svc := NewAuthService(&authUserRepo{usernameTaken: true}, &authSessionRepo{})

	result, err := svc.SignUp(context.Background(), SignUpInput{
		Email:    "new@example.com",
		Username: "taken",
		Password: "secret",
	})

	assert.Nil(t, result)
	var serviceErr *Error
	assert.True(t, errors.As(err, &serviceErr))
	assert.Equal(t, "conflict", serviceErr.Code)
}

func TestAuthService_SignInSuccess(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	userID := uuid.New()
	svc := NewAuthService(&authUserRepo{user: &models.User{
		ID:           userID,
		Email:        "user@example.com",
		Username:     "user",
		PasswordHash: string(hash),
	}}, &authSessionRepo{})
	email := "user@example.com"

	result, err := svc.SignIn(context.Background(), SignInInput{
		Email:    &email,
		Password: "secret",
	})

	assert.NoError(t, err)
	assert.Equal(t, userID, result.Session.UserID)
}

func TestAuthService_SignInWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	svc := NewAuthService(&authUserRepo{user: &models.User{
		ID:           uuid.New(),
		PasswordHash: string(hash),
	}}, &authSessionRepo{})
	email := "user@example.com"

	result, err := svc.SignIn(context.Background(), SignInInput{
		Email:    &email,
		Password: "wrong",
	})

	assert.Nil(t, result)
	var serviceErr *Error
	assert.True(t, errors.As(err, &serviceErr))
	assert.Equal(t, "unauthorized", serviceErr.Code)
}

func TestAuthService_SignOutSuccess(t *testing.T) {
	sessionID := uuid.New()
	sessionRepo := &authSessionRepo{session: &models.Session{ID: sessionID}}
	svc := NewAuthService(nil, sessionRepo)

	err := svc.SignOut(context.Background(), &models.Session{ID: sessionID})

	assert.NoError(t, err)
	assert.Nil(t, sessionRepo.session)
}
