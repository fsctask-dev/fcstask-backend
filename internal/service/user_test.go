package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	models "fcstask-backend/internal/db/model"
)

type fakeUserRepo struct {
	user       *models.User
	createErr  error
	emailTaken bool
}

func (r *fakeUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	user.ID = uuid.New()
	r.user = user
	return nil
}

func (r *fakeUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if r.user == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *fakeUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepo) GetUserByUserID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepo) GetUserByTgUID(ctx context.Context, tgUID int64) (*models.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepo) UpdateUser(ctx context.Context, user *models.User) error {
	r.user = user
	return nil
}

func (r *fakeUserRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	r.user = nil
	return nil
}

func (r *fakeUserRepo) GetUsersWithSessions(ctx context.Context, limit, offset int) ([]models.User, error) {
	return nil, nil
}

func (r *fakeUserRepo) CountUsersWithSessions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *fakeUserRepo) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	return r.emailTaken, nil
}

func (r *fakeUserRepo) ExistsUserByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (r *fakeUserRepo) CountUsers(ctx context.Context) (int64, error) {
	return 0, nil
}

func TestUserService_CreateUserSuccess(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewUserService(repo)
	userID := uuid.New()

	user, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:    "test@example.com",
		Username: "testuser",
		UserID:   userID,
	})

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, userID, user.UserID)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestUserService_CreateUserEmailConflict(t *testing.T) {
	repo := &fakeUserRepo{emailTaken: true}
	svc := NewUserService(repo)

	user, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:    "taken@example.com",
		Username: "testuser",
		UserID:   uuid.New(),
	})

	assert.Nil(t, user)
	var serviceErr *Error
	assert.True(t, errors.As(err, &serviceErr))
	assert.Equal(t, "conflict", serviceErr.Code)
}

func TestUserService_GetUserByIDNotFound(t *testing.T) {
	svc := NewUserService(&fakeUserRepo{})

	user, err := svc.GetUserByID(context.Background(), uuid.New())

	assert.Nil(t, user)
	var serviceErr *Error
	assert.True(t, errors.As(err, &serviceErr))
	assert.Equal(t, "not_found", serviceErr.Code)
}
