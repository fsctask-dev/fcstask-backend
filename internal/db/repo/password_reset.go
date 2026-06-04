package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type IPasswordResetRepository interface {
	Create(ctx context.Context, pr *model.PasswordReset) error
	Update(ctx context.Context, pr *model.PasswordReset) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.PasswordReset, error)
	GetByUserEmail(ctx context.Context, email string) (*model.PasswordReset, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

type PasswordResetRepository struct {
	db *gorm.DB
}

var _ IPasswordResetRepository = (*PasswordResetRepository)(nil)

func NewPasswordResetRepository(db *gorm.DB) IPasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(ctx context.Context, pr *model.PasswordReset) error {
	return r.db.WithContext(ctx).Create(pr).Error
}

func (r *PasswordResetRepository) Update(ctx context.Context, pr *model.PasswordReset) error {
	return r.db.WithContext(ctx).Save(pr).Error
}

func (r *PasswordResetRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.PasswordReset, error) {
	var pr model.PasswordReset
	if err := r.db.WithContext(ctx).First(&pr, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PasswordResetRepository) GetByUserEmail(ctx context.Context, email string) (*model.PasswordReset, error) {
	var pr model.PasswordReset
	if err := r.db.WithContext(ctx).
		Joins("JOIN users ON users.id = password_resets.user_id").
		Where("users.email = ?", email).
		First(&pr).Error; err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PasswordResetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.PasswordReset{}, "id = ?", id).Error
}

func (r *PasswordResetRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.PasswordReset{}).Error
}

func (r *PasswordResetRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&model.PasswordReset{})
	return res.RowsAffected, res.Error
}
