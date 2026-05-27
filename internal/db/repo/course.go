package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db"
	models "fcstask-backend/internal/db/model"
)

type CourseRepositoryInterface interface {
	GetCourses(ctx context.Context) ([]models.Course, error)
	GetCoursesByUserID(ctx context.Context, userID uuid.UUID, status string) ([]models.Course, error)
	GetCourseByID(ctx context.Context, courseID string) (*models.Course, error)
	CreateCourse(ctx context.Context, course models.Course) (*models.Course, error)
	UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error)
	DeleteCourse(ctx context.Context, courseID string) error
	GetCourseBoard(ctx context.Context, courseID uuid.UUID, userID uuid.UUID) (*models.TaskBoardSummary, bool, error)
	GetLeaderboard(ctx context.Context, courseID uuid.UUID) ([]models.LeaderboardEntry, error)
}

type CourseRepository struct {
	rw db.ReadWriter
}

func NewCourseRepository(rw db.ReadWriter) CourseRepositoryInterface {
	return &CourseRepository{rw: rw}
}

func (r *CourseRepository) GetCourses(ctx context.Context) ([]models.Course, error) {
	var courses []models.Course
	if err := r.rw.ReadDB().WithContext(ctx).Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *CourseRepository) GetCoursesByUserID(ctx context.Context, userID uuid.UUID, status string) ([]models.Course, error) {
	var courseIDs []uuid.UUID
	if err := r.rw.ReadDB().WithContext(ctx).
		Model(&models.UserRole{}).
		Where("user_id = ?", userID).
		Pluck("course_id", &courseIDs).Error; err != nil {
		return nil, err
	}
	var courses []models.Course
	if len(courseIDs) == 0 {
		return courses, nil
	}
	query := r.rw.ReadDB().WithContext(ctx).Where("id IN ?", courseIDs)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *CourseRepository) GetCourseByID(ctx context.Context, courseID string) (*models.Course, error) {
	var course models.Course
	query := r.rw.ReadDB().WithContext(ctx)
	if id, err := uuid.Parse(courseID); err == nil {
		query = query.Where("id = ? OR slug = ?", id, courseID)
	} else {
		query = query.Where("slug = ?", courseID)
	}

	err := query.First(&course).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) CreateCourse(ctx context.Context, course models.Course) (*models.Course, error) {
	if err := r.rw.WriteDB().WithContext(ctx).Create(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error) {
	if course.ID == uuid.Nil {
		return nil, errors.New("course id is required")
	}
	if course.Slug == "" {
		course.Slug = courseID
	}
	if err := r.rw.WriteDB().WithContext(ctx).Save(&course).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) DeleteCourse(ctx context.Context, courseID string) error {
	if id, err := uuid.Parse(courseID); err == nil {
		return r.rw.WriteDB().WithContext(ctx).
			Where("id = ?", id).
			Delete(&models.Course{}).Error
	}

	return r.rw.WriteDB().WithContext(ctx).
		Where("slug = ?", courseID).
		Delete(&models.Course{}).Error
}

func (r *CourseRepository) GetCourseBoard(ctx context.Context, courseID uuid.UUID, userID uuid.UUID) (*models.TaskBoardSummary, bool, error) {
	return nil, false, nil
}

func (r *CourseRepository) GetLeaderboard(ctx context.Context, courseID uuid.UUID) ([]models.LeaderboardEntry, error) {
	type result struct {
		Username   string
		TotalScore int
	}
	var results []result
	err := r.rw.ReadDB().WithContext(ctx).
		Model(&models.UserRole{}).
		Select("u.username, COALESCE(SUM(sts.score), 0) AS total_score").
		Joins("JOIN users u ON u.id = user_roles.user_id").
		Joins("JOIN course_admin_permissions cap ON cap.role_id = user_roles.role_id AND cap.permission = ?", "task.submit").
		Joins("LEFT JOIN student_task_scores sts ON sts.student_id = user_roles.user_id AND sts.course_id = user_roles.course_id").
		Where("user_roles.course_id = ?", courseID).
		Group("user_roles.user_id, u.username").
		Order("total_score DESC").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	entries := make([]models.LeaderboardEntry, len(results))
	for i, r := range results {
		entries[i] = models.LeaderboardEntry{
			Username:   r.Username,
			TotalScore: r.TotalScore,
			Rank:       i + 1,
		}
	}
	return entries, nil
}
