package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db"
	models "fcstask-backend/internal/db/model"
)

type NamespaceRepositoryInterface interface {
	GetAll(ctx context.Context) ([]models.Namespace, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Namespace, error)
	GetBySlug(ctx context.Context, slug string) (*models.Namespace, error)
	GetUsers(ctx context.Context, namespaceID uuid.UUID) ([]models.NamespaceUser, error)
}

type NamespaceRepository struct {
	rw db.ReadWriter
}

func NewNamespaceRepository(rw db.ReadWriter) NamespaceRepositoryInterface {
	return &NamespaceRepository{rw: rw}
}

func (r *NamespaceRepository) GetAll(ctx context.Context) ([]models.Namespace, error) {
	var namespaces []models.Namespace
	if err := r.rw.ReadDB().WithContext(ctx).Find(&namespaces).Error; err != nil {
		return nil, err
	}
	return namespaces, nil
}

func (r *NamespaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Namespace, error) {
	var namespace models.Namespace
	err := r.rw.ReadDB().WithContext(ctx).First(&namespace, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &namespace, nil
}

func (r *NamespaceRepository) GetBySlug(ctx context.Context, slug string) (*models.Namespace, error) {
	var namespace models.Namespace
	err := r.rw.ReadDB().WithContext(ctx).Where("slug = ?", slug).First(&namespace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &namespace, nil
}

func (r *NamespaceRepository) GetUsers(ctx context.Context, namespaceID uuid.UUID) ([]models.NamespaceUser, error) {
	var users []models.NamespaceUser
	if err := r.rw.ReadDB().WithContext(ctx).
		Where("namespace_id = ?", namespaceID).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
