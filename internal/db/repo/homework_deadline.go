package repo

import (
	"context"
	"fcstask-backend/internal/db/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IHomeworkDeadlineRepo interface {
	GetByHwID(ctx context.Context, HwID uuid.UUID) (*model.HomeworkDeadline, error)
	Create(ctx context.Context, d *model.HomeworkDeadline) error
	Update(ctx context.Context, d *model.HomeworkDeadline) error
}

type HomeworkDeadlineRepo struct{ db *gorm.DB }

func NewHomeworkDeadlineRepo(db *gorm.DB) *HomeworkDeadlineRepo {
	return &HomeworkDeadlineRepo{db: db}
}

func (r *HomeworkDeadlineRepo) GetByHwID(ctx context.Context, HwID uuid.UUID) (*model.HomeworkDeadline, error) {
	var d model.HomeworkDeadline
	err := r.db.WithContext(ctx).Where("hw_id = ?", HwID).First(&d).Error
	return &d, err
}

func (r *HomeworkDeadlineRepo) Create(ctx context.Context, d *model.HomeworkDeadline) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *HomeworkDeadlineRepo) Update(ctx context.Context, d *model.HomeworkDeadline) error {
	return r.db.WithContext(ctx).Save(d).Error
}
