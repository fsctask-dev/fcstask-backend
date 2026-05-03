package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type IEmailRegistrationRepo interface {
	Create(ctx context.Context, reg *model.EmailRegistration) error
	Update(ctx context.Context, reg *model.EmailRegistration) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.EmailRegistration, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByEmail(ctx context.Context, email string) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type EmailRegistrationRepository struct {
	db *gorm.DB
}

var _ IEmailRegistrationRepo = (*EmailRegistrationRepository)(nil)

func NewEmailRegistrationRepository(db *gorm.DB) IEmailRegistrationRepo {
	return &EmailRegistrationRepository{db: db}
}

func (r *EmailRegistrationRepository) Create(ctx context.Context, reg *model.EmailRegistration) error {
	return r.db.WithContext(ctx).Create(reg).Error
}

func (r *EmailRegistrationRepository) Update(ctx context.Context, reg *model.EmailRegistration) error {
	return r.db.WithContext(ctx).Save(reg).Error
}

func (r *EmailRegistrationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.EmailRegistration, error) {
	var reg model.EmailRegistration
	if err := r.db.WithContext(ctx).First(&reg, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *EmailRegistrationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.EmailRegistration{}, "id = ?", id).Error
}

func (r *EmailRegistrationRepository) DeleteByEmail(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).Where("email = ?", email).Delete(&model.EmailRegistration{}).Error
}

func (r *EmailRegistrationRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&model.EmailRegistration{})
	return res.RowsAffected, res.Error
}
