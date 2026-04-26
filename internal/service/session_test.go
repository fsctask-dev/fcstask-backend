package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	models "fcstask-backend/internal/db/model"
)

type sessionServiceRepo struct {
	sessions []models.Session
	total    int64
	countErr error
	listErr  error
}

func (r *sessionServiceRepo) CreateSession(ctx context.Context, session *models.Session) error {
	return nil
}

func (r *sessionServiceRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *sessionServiceRepo) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	return nil, nil
}

func (r *sessionServiceRepo) GetSessionsWithUser(ctx context.Context, limit, offset int) ([]models.Session, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.sessions, nil
}

func (r *sessionServiceRepo) CountSessions(ctx context.Context) (int64, error) {
	if r.countErr != nil {
		return 0, r.countErr
	}
	return r.total, nil
}

func (r *sessionServiceRepo) TouchSessionAccessedAt(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *sessionServiceRepo) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *sessionServiceRepo) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (r *sessionServiceRepo) CleanOutdatedSessions(ctx context.Context, ttl time.Duration) (int64, error) {
	return 0, nil
}

func TestSessionService_GetSessionsSuccess(t *testing.T) {
	repo := &sessionServiceRepo{
		total:    2,
		sessions: []models.Session{{ID: uuid.New()}, {ID: uuid.New()}},
	}
	svc := NewSessionService(repo)

	sessions, total, err := svc.GetSessions(context.Background(), 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, sessions, 2)
}

func TestSessionService_GetSessionsInvalidLimit(t *testing.T) {
	svc := NewSessionService(&sessionServiceRepo{})

	sessions, total, err := svc.GetSessions(context.Background(), 101, 0)

	assert.Nil(t, sessions)
	assert.Equal(t, int64(0), total)
	var serviceErr *Error
	assert.True(t, errors.As(err, &serviceErr))
	assert.Equal(t, "bad_request", serviceErr.Code)
}

func TestSessionService_GetSessionsCountError(t *testing.T) {
	svc := NewSessionService(&sessionServiceRepo{countErr: errors.New("db down")})

	sessions, total, err := svc.GetSessions(context.Background(), 20, 0)

	assert.Nil(t, sessions)
	assert.Equal(t, int64(0), total)
	var serviceErr *Error
	assert.True(t, errors.As(err, &serviceErr))
	assert.Equal(t, "internal_error", serviceErr.Code)
}
