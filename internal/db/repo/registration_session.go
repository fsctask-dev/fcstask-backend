package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type IRegistrationSessionRepo interface {
	Create(ctx context.Context, session *model.RegistrationSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.RegistrationSession, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type RegistrationSessionRepository struct {
	db *gorm.DB
}

var _ IRegistrationSessionRepo = (*RegistrationSessionRepository)(nil)

func NewRegistrationSessionRepository(db *gorm.DB) IRegistrationSessionRepo {
	return &RegistrationSessionRepository{db: db}
}

func (r *RegistrationSessionRepository) Create(ctx context.Context, session *model.RegistrationSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *RegistrationSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.RegistrationSession, error) {
	var session model.RegistrationSession
	err := r.db.WithContext(ctx).First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *RegistrationSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.RegistrationSession{}, "id = ?", id).Error
}

func (r *RegistrationSessionRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&model.RegistrationSession{})
	return res.RowsAffected, res.Error
}
