package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type ILatePolicyRepo interface {
	Create(ctx context.Context, policy *model.LatePolicy) error
	GetByHwID(ctx context.Context, hwId uuid.UUID) (*model.LatePolicy, error)
	Update(ctx context.Context, policy *model.LatePolicy) error
	Delete(ctx context.Context, hwId uuid.UUID) error
}

type LatePolicyRepository struct {
	db *gorm.DB
}

var _ ILatePolicyRepo = (*LatePolicyRepository)(nil)

func NewLatePolicyRepository(db *gorm.DB) ILatePolicyRepo {
	return &LatePolicyRepository{db: db}
}

func (r *LatePolicyRepository) Create(ctx context.Context, policy *model.LatePolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

func (r *LatePolicyRepository) GetByHwID(ctx context.Context, hwId uuid.UUID) (*model.LatePolicy, error) {
	var policy model.LatePolicy
	err := r.db.WithContext(ctx).
		Where("hw_id = ?", hwId).
		First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *LatePolicyRepository) Update(ctx context.Context, policy *model.LatePolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

func (r *LatePolicyRepository) Delete(ctx context.Context, hwId uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("hw_id = ?", hwId).
		Delete(&model.LatePolicy{}).Error
}
