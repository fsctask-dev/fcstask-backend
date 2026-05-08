package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type IDeadlineRepo interface {
	Create(ctx context.Context, deadline *model.Deadline) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Deadline, error)
	GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.Deadline, error)
	Update(ctx context.Context, deadline *model.Deadline) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type DeadlineRepository struct {
	db *gorm.DB
}

var _ IDeadlineRepo = (*DeadlineRepository)(nil)

func NewDeadlineRepository(db *gorm.DB) IDeadlineRepo {
	return &DeadlineRepository{db: db}
}

func (r *DeadlineRepository) Create(ctx context.Context, deadline *model.Deadline) error {
	return r.db.WithContext(ctx).Create(deadline).Error
}

func (r *DeadlineRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Deadline, error) {
	var deadline model.Deadline
	err := r.db.WithContext(ctx).First(&deadline, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &deadline, nil
}

func (r *DeadlineRepository) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.Deadline, error) {
	var deadlines []model.Deadline
	err := r.db.WithContext(ctx).
		Where("course_id = ?", courseID).
		Order("due_date ASC").
		Find(&deadlines).Error
	if err != nil {
		return nil, err
	}
	return deadlines, nil
}

func (r *DeadlineRepository) Update(ctx context.Context, deadline *model.Deadline) error {
	return r.db.WithContext(ctx).Save(deadline).Error
}

func (r *DeadlineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Deadline{}, "id = ?", id).Error
}
