package repo

import (
	"context"

	"gorm.io/gorm"

	models "fcstask-backend/internal/db/model"
)

type CourseRepositoryInterface interface {
	GetCourses(ctx context.Context) ([]models.Course, error)
	GetCourseByID(ctx context.Context, courseID string) (*models.Course, error)
	CreateCourse(ctx context.Context, course models.Course) (*models.Course, error)
	UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error)
	DeleteCourse(ctx context.Context, courseID string) error
	GetCourseBoard(ctx context.Context, courseID string) (*models.TaskBoardSummary, bool, error)
}

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepositoryInterface {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) GetCourses(ctx context.Context) ([]models.Course, error) {
	var courses []models.Course
	if err := r.db.WithContext(ctx).Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *CourseRepository) GetCourseByID(ctx context.Context, courseID string) (*models.Course, error) {
	var course models.Course
	err := r.db.WithContext(ctx).
		Where("id = ? OR slug = ?", courseID, courseID).
		First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) CreateCourse(ctx context.Context, course models.Course) (*models.Course, error) {
	if err := r.db.WithContext(ctx).Create(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error) {
	if course.ID == "" {
		course.ID = courseID
	}
	if course.Slug == "" {
		course.Slug = courseID
	}
	if err := r.db.WithContext(ctx).Save(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) DeleteCourse(ctx context.Context, courseID string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", courseID).
		Delete(&models.Course{}).Error
}

func (r *CourseRepository) GetCourseBoard(ctx context.Context, courseID string) (*models.TaskBoardSummary, bool, error) {
	return nil, false, nil
}
