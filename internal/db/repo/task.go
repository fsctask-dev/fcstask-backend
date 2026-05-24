package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type ITaskRepo interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	GetByHwID(ctx context.Context, hwID uuid.UUID) ([]model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetScore(ctx context.Context, id uuid.UUID, score int) error
}

type TaskRepository struct {
	db *gorm.DB
}

var _ ITaskRepo = (*TaskRepository)(nil)

func NewTaskRepository(db *gorm.DB) ITaskRepo {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).First(&task, "task_id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) GetByHwID(ctx context.Context, hwID uuid.UUID) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.WithContext(ctx).
		Where("hw_id = ?", hwID).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Task{}, "task_id = ?", id).Error
}

func (r *TaskRepository) SetScore(ctx context.Context, id uuid.UUID, score int) error {
	return r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("task_id = ?", id).
		Update("score", score).Error
}
