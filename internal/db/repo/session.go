package repo

import (
	"context"
	"time"

	"github.com/google/uuid"

	"fcstask-backend/internal/db"
	"fcstask-backend/internal/db/model"
)

type ISessionRepository interface {
	CreateSession(ctx context.Context, session *model.Session) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*model.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Session, error)
	GetSessionsWithUser(ctx context.Context, limit, offset int) ([]model.Session, error)
	CountSessions(ctx context.Context) (int64, error)
	TouchSessionAccessedAt(ctx context.Context, id uuid.UUID) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error
	CleanOutdatedSessions(ctx context.Context, ttl time.Duration) (int64, error)
}

type SessionRepository struct {
	rw db.ReadWriter
}

var _ ISessionRepository = (*SessionRepository)(nil)

func NewSessionRepository(rw db.ReadWriter) ISessionRepository {
	return &SessionRepository{rw: rw}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session *model.Session) error {
	return r.rw.WriteDB().WithContext(ctx).Create(session).Error
}

func (r *SessionRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	var session model.Session
	err := r.rw.ReadDB().WithContext(ctx).First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Session, error) {
	var sessions []model.Session
	err := r.rw.ReadDB().WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *SessionRepository) GetSessionsWithUser(ctx context.Context, limit, offset int) ([]model.Session, error) {
	var sessions []model.Session
	err := r.rw.ReadDB().WithContext(ctx).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *SessionRepository) CountSessions(ctx context.Context) (int64, error) {
	var count int64
	err := r.rw.ReadDB().WithContext(ctx).Model(&model.Session{}).Count(&count).Error
	return count, err
}

func (r *SessionRepository) TouchSessionAccessedAt(ctx context.Context, id uuid.UUID) error {
	return r.rw.WriteDB().WithContext(ctx).Model(&model.Session{}).Where("id = ?", id).Update("accessed_at", time.Now()).Error
}

func (r *SessionRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return r.rw.WriteDB().WithContext(ctx).Delete(&model.Session{}, "id = ?", id).Error
}

func (r *SessionRepository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.rw.WriteDB().WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Session{}).Error
}

func (r *SessionRepository) CleanOutdatedSessions(ctx context.Context, ttl time.Duration) (int64, error) {
	cutoff := time.Now().UTC().Add(-ttl)
	result := r.rw.WriteDB().WithContext(ctx).Where("accessed_at < ?", cutoff).Delete(&model.Session{})
	return result.RowsAffected, result.Error
}
