package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type NamespaceUserRow struct {
	UserID   uuid.UUID
	Username string
	Role     string
}

type INamespaceRepo interface {
	ListNamespaces(ctx context.Context) ([]model.Namespace, error)
	GetNamespaceByID(ctx context.Context, id uuid.UUID) (*model.Namespace, error)
	GetNamespaceUsers(ctx context.Context, namespaceID uuid.UUID) ([]NamespaceUserRow, error)
	GetNamespaceCourses(ctx context.Context, namespaceID uuid.UUID) ([]model.Course, error)
	CountUsers(ctx context.Context, namespaceID uuid.UUID) (int64, error)
	CountCourses(ctx context.Context, namespaceID uuid.UUID) (int64, error)
}

type NamespaceRepository struct {
	db *gorm.DB
}

func NewNamespaceRepository(db *gorm.DB) INamespaceRepo {
	return &NamespaceRepository{db: db}
}

func (r *NamespaceRepository) ListNamespaces(ctx context.Context) ([]model.Namespace, error) {
	var namespaces []model.Namespace
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&namespaces).Error; err != nil {
		return nil, err
	}
	return namespaces, nil
}

func (r *NamespaceRepository) GetNamespaceByID(ctx context.Context, id uuid.UUID) (*model.Namespace, error) {
	var ns model.Namespace
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&ns).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ns, nil
}

func (r *NamespaceRepository) GetNamespaceUsers(ctx context.Context, namespaceID uuid.UUID) ([]NamespaceUserRow, error) {
	var rows []NamespaceUserRow
	err := r.db.WithContext(ctx).
		Table("namespace_users nu").
		Select("nu.user_id, u.username, nu.role").
		Joins("JOIN users u ON u.id = nu.user_id").
		Where("nu.namespace_id = ?", namespaceID).
		Scan(&rows).Error
	return rows, err
}

func (r *NamespaceRepository) GetNamespaceCourses(ctx context.Context, namespaceID uuid.UUID) ([]model.Course, error) {
	var courses []model.Course
	err := r.db.WithContext(ctx).
		Joins("JOIN namespace_courses nc ON nc.course_id = courses.id").
		Where("nc.namespace_id = ? AND courses.deleted_at IS NULL", namespaceID).
		Find(&courses).Error
	return courses, err
}

func (r *NamespaceRepository) CountUsers(ctx context.Context, namespaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.NamespaceUser{}).
		Where("namespace_id = ?", namespaceID).
		Count(&count).Error
	return count, err
}

func (r *NamespaceRepository) CountCourses(ctx context.Context, namespaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.NamespaceCourse{}).
		Where("namespace_id = ?", namespaceID).
		Count(&count).Error
	return count, err
}
