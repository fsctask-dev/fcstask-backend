package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (r *CourseRepository) GetCourseBoard(ctx context.Context, courseID string, userID uuid.UUID) (*models.TaskBoardSummary, bool, error) {
	courseUUID, err := uuid.Parse(courseID)
	if err != nil {
		return nil, false, err
	}

	var course models.Course
	if err := r.rw.ReadDB().WithContext(ctx).First(&course, "id = ?", courseUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var homeworks []models.Homework
	if err := r.rw.ReadDB().WithContext(ctx).
		Where("course_id = ?", courseUUID).
		Order("created_at ASC").
		Find(&homeworks).Error; err != nil {
		return nil, false, err
	}

	if len(homeworks) == 0 {
		return &models.TaskBoardSummary{
			CourseName:   course.Name,
			CourseStatus: course.Status,
			Groups:       []models.BoardGroup{},
		}, true, nil
	}

	hwIDs := make([]uuid.UUID, len(homeworks))
	for i, hw := range homeworks {
		hwIDs[i] = hw.HwID
	}

	var tasks []models.Task
	r.rw.ReadDB().WithContext(ctx).
		Where("hw_id IN ?", hwIDs).
		Find(&tasks)
	tasksByHW := make(map[uuid.UUID][]models.Task)
	for _, t := range tasks {
		tasksByHW[t.HwID] = append(tasksByHW[t.HwID], t)
	}

	var deadlines []models.Deadline
	r.rw.ReadDB().WithContext(ctx).
		Where("homework_id IN ?", hwIDs).
		Find(&deadlines)
	deadlinesByHW := make(map[uuid.UUID][]models.Deadline)
	for _, dl := range deadlines {
		if dl.HomeworkID != nil {
			deadlinesByHW[*dl.HomeworkID] = append(deadlinesByHW[*dl.HomeworkID], dl)
		}
	}

	type taskScore struct {
		TaskID uuid.UUID
		Score  int
	}
	var scores []taskScore
	r.rw.ReadDB().WithContext(ctx).
		Model(&models.StudentTaskScore{}).
		Select("task_id, score").
		Where("student_id = ? AND course_id = ?", userID, courseUUID).
		Find(&scores)
	scoreMap := make(map[uuid.UUID]int)
	for _, s := range scores {
		scoreMap[s.TaskID] = s.Score
	}

	summary := &models.TaskBoardSummary{
		CourseName:   course.Name,
		CourseStatus: course.Status,
		Groups:       make([]models.BoardGroup, 0, len(homeworks)),
	}

	totalMax := 0
	totalSolved := 0

	for _, hw := range homeworks {
		group := models.BoardGroup{
			ID:        hw.HwID.String(),
			Name:      fmt.Sprintf("HW %s", hw.HwID.String()[:8]),
			IsSpecial: hw.IsPublic,
			StartedAt: formatTime(hw.StartDate),
			EndsAt:    formatTime(hw.EndDate),
			Deadlines: make([]models.BoardDeadline, 0),
			Tasks:     make([]models.BoardTask, 0),
		}

		for _, dl := range deadlinesByHW[hw.HwID] {
			group.Deadlines = append(group.Deadlines, models.BoardDeadline{
				ID:      dl.ID.String(),
				Label:   dl.Title,
				DueAt:   dl.DueDate.Format(time.RFC3339),
				Status:  deadlineStatus(dl.DueDate),
				Percent: 0,
			})
		}

		for _, task := range tasksByHW[hw.HwID] {
			scoreCfg := 0
			if task.Score != nil {
				scoreCfg = *task.Score
			}
			earned := scoreMap[task.TaskID]

			stats := 0.0
			if scoreCfg > 0 {
				stats = float64(earned) / float64(scoreCfg)
			}

			group.Tasks = append(group.Tasks, models.BoardTask{
				ID:          task.TaskID.String(),
				Name:        task.Title,
				Score:       scoreCfg,
				ScoreEarned: earned,
				Stats:       stats,
				URL:         stringPtrValue(task.TaskURL),
			})

			totalMax += scoreCfg
			totalSolved += earned
		}

		summary.Groups = append(summary.Groups, group)
	}

	summary.MaxScore = totalMax
	summary.SolvedScore = totalSolved
	if totalMax > 0 {
		summary.SolvedPercent = (totalSolved * 100) / totalMax
	}

	return summary, true, nil
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

func deadlineStatus(dueDate time.Time) string {
	now := time.Now()
	if dueDate.Before(now) {
		return "expired"
	}
	if dueDate.Before(now.Add(24 * time.Hour)) {
		return "urgent"
	}
	return "active"
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
