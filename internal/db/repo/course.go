package repo

import (
	"context"
	"errors"
	"sort"

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
	GetCourseBoard(ctx context.Context, courseID string, userID uuid.UUID) (*models.TaskBoardSummary, bool, error)
	GetLeaderboard(ctx context.Context, courseID uuid.UUID) ([]models.LeaderboardEntry, error)
	UpdateInviteCode(ctx context.Context, courseID uuid.UUID, code *string) error
	GetPublicCourses(ctx context.Context) ([]models.Course, error)
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

func (r *CourseRepository) GetPublicCourses(ctx context.Context) ([]models.Course, error) {
    var courses []models.Course
    if err := r.rw.ReadDB().WithContext(ctx).
        Where("type = ?", models.CourseTypePublic).
        Find(&courses).Error; err != nil {
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

func (r *CourseRepository) GetCourseBoard(ctx context.Context, courseID string, userID uuid.UUID) (*models.TaskBoardSummary, bool, error) {
	return nil, false, nil
}

func (r *CourseRepository) GetLeaderboard(ctx context.Context, courseID uuid.UUID) ([]models.LeaderboardEntry, error) {
	type taskScoreRow struct {
		UserID   uuid.UUID
		Username string
		TaskID   uuid.UUID
		Score    int
	}
	var rows []taskScoreRow
	err := r.rw.ReadDB().WithContext(ctx).
		Model(&models.UserRole{}).
		Select("u.id AS user_id, u.username, sts.task_id, COALESCE(sts.score, 0) AS score").
		Joins("JOIN users u ON u.id = user_roles.user_id").
		Joins("JOIN course_admin_permissions cap ON cap.role_id = user_roles.role_id AND cap.permission = ?", "task.submit").
		Joins("LEFT JOIN student_task_scores sts ON sts.student_id = user_roles.user_id AND sts.course_id = user_roles.course_id").
		Where("user_roles.course_id = ?", courseID).
		Order("u.username ASC, sts.task_id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	type userScores struct {
		Username   string
		TotalScore int
		Tasks      map[uuid.UUID]int
	}
	userMap := make(map[uuid.UUID]*userScores)
	var userOrder []uuid.UUID

	for _, row := range rows {
		us, ok := userMap[row.UserID]
		if !ok {
			us = &userScores{
				Username: row.Username,
				Tasks:    make(map[uuid.UUID]int),
			}
			userMap[row.UserID] = us
			userOrder = append(userOrder, row.UserID)
		}
		if row.TaskID != uuid.Nil {
			us.Tasks[row.TaskID] = row.Score
			us.TotalScore += row.Score
		}
	}

	entries := make([]models.LeaderboardEntry, 0, len(userOrder))
	for _, userID := range userOrder {
		us := userMap[userID]
		entries = append(entries, models.LeaderboardEntry{
			Username:   us.Username,
			TotalScore: us.TotalScore,
			Tasks:      us.Tasks,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].TotalScore > entries[j].TotalScore
	})
	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries, nil
}

func (r *CourseRepository) UpdateInviteCode(ctx context.Context, courseID uuid.UUID, code *string) error {
    return r.rw.WriteDB().WithContext(ctx).
        Model(&models.Course{}).
        Where("id = ?", courseID).
        Update("invite_code", code).Error
}