package repo

import (
	"context"
	"fcstask-backend/internal/db/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IStudentTaskScoreRepo interface {
	Upsert(ctx context.Context, score *model.StudentTaskScore) error
	GetByStudentAndCourse(ctx context.Context, studentID, courseID uuid.UUID) ([]model.StudentTaskScore, error)
}

type StudentTaskScoreRepository struct {
	db *gorm.DB
}

func NewStudentTaskScoreRepository(db *gorm.DB) IStudentTaskScoreRepo {
	return &StudentTaskScoreRepository{db: db}
}

func (r *StudentTaskScoreRepository) Upsert(ctx context.Context, score *model.StudentTaskScore) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "student_id"}, {Name: "task_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"score", "is_passed", "updated_at"}),
		}).
		Create(score).Error
}

func (r *StudentTaskScoreRepository) GetByStudentAndCourse(ctx context.Context, studentID, courseID uuid.UUID) ([]model.StudentTaskScore, error) {
	var scores []model.StudentTaskScore
	err := r.db.WithContext(ctx).
		Where("student_id = ? AND course_id = ?", studentID, courseID).
		Find(&scores).Error

	return scores, err
}
