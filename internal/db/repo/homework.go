package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type IHomeworkRepo interface {
	Create(ctx context.Context, hw *model.Homework) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Homework, error)
	GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.Homework, error)
	Update(ctx context.Context, hw *model.Homework) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type HomeworkRepository struct {
	db *gorm.DB
}

var _ IHomeworkRepo = (*HomeworkRepository)(nil)

func NewHomeworkRepository(db *gorm.DB) IHomeworkRepo {
	return &HomeworkRepository{db: db}
}

func (r *HomeworkRepository) Create(ctx context.Context, hw *model.Homework) error {
	return r.db.WithContext(ctx).Create(hw).Error
}

func (r *HomeworkRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Homework, error) {
	var hw model.Homework
	err := r.db.WithContext(ctx).First(&hw, "hw_id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &hw, nil
}

func (r *HomeworkRepository) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.Homework, error) {
	var hws []model.Homework
	err := r.db.WithContext(ctx).
		Where("course_id = ?", courseID).
		Order("position ASC, created_at ASC").
		Find(&hws).Error
	if err != nil {
		return nil, err
	}
	return hws, nil
}

func (r *HomeworkRepository) Update(ctx context.Context, hw *model.Homework) error {
	return r.db.WithContext(ctx).Save(hw).Error
}

func (r *HomeworkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Homework{}, "hw_id = ?", id).Error
}
