package repo

import (
	"context"
	"fcstask-backend/internal/db/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ICourseLatePolicy interface {
	GetByCourseID(ctx context.Context, courseID uuid.UUID) (*model.CourseLatePolicy, error)
	Create(ctx context.Context, policy *model.CourseLatePolicy) error
	Update(ctx context.Context, policy *model.CourseLatePolicy) error
}

type CourseLatePolicy struct{ db *gorm.DB }

func NewCourseLatePolicy(db *gorm.DB) *CourseLatePolicy {
	return &CourseLatePolicy{db: db}
}

func (r *CourseLatePolicy) GetByCourseID(ctx context.Context, courseID uuid.UUID) (*model.CourseLatePolicy, error) {
	var p model.CourseLatePolicy
	err := r.db.WithContext(ctx).Where("course_id = ?", courseID).First(&p).Error
	return &p, err
}

func (r *CourseLatePolicy) Create(ctx context.Context, policy *model.CourseLatePolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

func (r *CourseLatePolicy) Update(ctx context.Context, policy *model.CourseLatePolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}
