package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type IOAuthIdentityRepo interface {
	Create(ctx context.Context, identity *model.OAuthIdentity) error
	Update(ctx context.Context, identity *model.OAuthIdentity) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.OAuthIdentity, error)
	GetByProviderUID(ctx context.Context, provider, providerUID string) (*model.OAuthIdentity, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthIdentity, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type OAuthIdentityRepository struct {
	db *gorm.DB
}

var _ IOAuthIdentityRepo = (*OAuthIdentityRepository)(nil)

func NewOAuthIdentityRepository(db *gorm.DB) IOAuthIdentityRepo {
	return &OAuthIdentityRepository{db: db}
}

func (r *OAuthIdentityRepository) Create(ctx context.Context, identity *model.OAuthIdentity) error {
	return r.db.WithContext(ctx).Create(identity).Error
}

func (r *OAuthIdentityRepository) Update(ctx context.Context, identity *model.OAuthIdentity) error {
	return r.db.WithContext(ctx).Save(identity).Error
}

func (r *OAuthIdentityRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.OAuthIdentity, error) {
	var identity model.OAuthIdentity
	err := r.db.WithContext(ctx).First(&identity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *OAuthIdentityRepository) GetByProviderUID(ctx context.Context, provider, providerUID string) (*model.OAuthIdentity, error) {
	var identity model.OAuthIdentity
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_uid = ?", provider, providerUID).
		First(&identity).Error
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *OAuthIdentityRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthIdentity, error) {
	var identities []model.OAuthIdentity
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&identities).Error
	if err != nil {
		return nil, err
	}
	return identities, nil
}

func (r *OAuthIdentityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.OAuthIdentity{}, "id = ?", id).Error
}
