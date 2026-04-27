package repo

import (
	"context"

	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type CourseRepositoryInterface interface {
	Create(ctx context.Context, course *model.Course) error
	GetByID(ctx context.Context, id string) (*model.Course, error)
	GetBySlug(ctx context.Context, slug string) (*model.Course, error)
	GetAll(ctx context.Context, statusFilter string) ([]model.Course, error)
	Update(ctx context.Context, course *model.Course) error
	Delete(ctx context.Context, id string) error
}

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepositoryInterface {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) Create(ctx context.Context, course *model.Course) error {
	return r.db.WithContext(ctx).Create(course).Error
}

func (r *CourseRepository) GetByID(ctx context.Context, id string) (*model.Course, error) {
	var course model.Course
	err := r.db.WithContext(ctx).First(&course, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) GetBySlug(ctx context.Context, slug string) (*model.Course, error) {
	var course model.Course
	err := r.db.WithContext(ctx).First(&course, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) GetAll(ctx context.Context, statusFilter string) ([]model.Course, error) {
	var courses []model.Course
	query := r.db.WithContext(ctx)

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	err := query.Find(&courses).Error
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *CourseRepository) Update(ctx context.Context, course *model.Course) error {
	return r.db.WithContext(ctx).Save(course).Error
}

func (r *CourseRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Course{}, "id = ?", id).Error
}
