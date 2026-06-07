package repo

import (
	"context"
	"errors"
	"sort"
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
	GetCourseInfo(ctx context.Context, courseID uuid.UUID) (*models.CourseInfo, error)
	GetLeaderboard(ctx context.Context, courseID string) ([]models.LeaderboardEntry, error)
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
	courseUUID, err := uuid.Parse(courseID)
	var course models.Course
	if err != nil {
		if err := r.rw.ReadDB().WithContext(ctx).First(&course, "slug = ?", courseID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, err
		}
	} else {
		if err := r.rw.ReadDB().WithContext(ctx).First(&course, "id = ?", courseUUID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, err
		}
	}

	var homeworks []models.Homework
	if err := r.rw.ReadDB().WithContext(ctx).
		Where("course_id = ? AND is_public = true", course.ID).
		Order("position ASC").
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
	if err := r.rw.ReadDB().WithContext(ctx).
		Where("hw_id IN ? AND is_public = true", hwIDs).
		Find(&tasks).Error; err != nil {
		return nil, false, err
	}
	tasksByHW := make(map[uuid.UUID][]models.Task)
	for _, t := range tasks {
		tasksByHW[t.HwID] = append(tasksByHW[t.HwID], t)
	}

	var deadlines []models.Deadline
	if err := r.rw.ReadDB().WithContext(ctx).
		Where("homework_id IN ?", hwIDs).
		Find(&deadlines).Error; err != nil {
		return nil, false, err
	}
	deadlinesByHW := make(map[uuid.UUID][]models.Deadline)
	for _, dl := range deadlines {
		if dl.HomeworkID != uuid.Nil {
			deadlinesByHW[dl.HomeworkID] = append(deadlinesByHW[dl.HomeworkID], dl)
		}
	}

	type taskScore struct {
		TaskID uuid.UUID
		Score  int
	}
	var scores []taskScore
	if err := r.rw.ReadDB().WithContext(ctx).
		Model(&models.StudentTaskScore{}).
		Select("task_id, score").
		Where("student_id = ? AND course_id = ?", userID, course.ID).
		Find(&scores).Error; err != nil {
		return nil, false, err
	}
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
			Name:      hw.Title,
			IsSpecial: hw.IsPublic,
			StartedAt: formatTime(hw.StartDate),
			EndsAt:    formatTime(hw.EndDate),
			Deadlines: make([]models.BoardDeadline, 0),
			Tasks:     make([]models.BoardTask, 0),
		}

		for _, dl := range deadlinesByHW[hw.HwID] {
			group.Deadlines = append(group.Deadlines, models.BoardDeadline{
				ID:           dl.ID.String(),
				Label:        dl.Title,
				SoftDeadline: dl.SoftDeadline,
				HardDeadline: dl.HardDeadline,
				SoftStatus:   deadlineStatus(dl.SoftDeadline),
				HardStatus:   deadlineStatus(dl.HardDeadline),
				Percent:      0,
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

func (r *CourseRepository) GetCourseInfo(ctx context.Context, courseID uuid.UUID) (*models.CourseInfo, error) {
	var course models.Course
	if err := r.rw.ReadDB().WithContext(ctx).First(&course, "id = ?", courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var homeworks []models.Homework
	if err := r.rw.ReadDB().WithContext(ctx).Where("course_id = ? AND is_public = ?", courseID, true).Order("position ASC, created_at ASC").Find(&homeworks).Error; err != nil {
		return nil, err
	}

	hwIDs := make([]uuid.UUID, len(homeworks))
	for i, hw := range homeworks {
		hwIDs[i] = hw.HwID
	}

	var allTasks []models.Task
	if len(hwIDs) > 0 {
		if err := r.rw.ReadDB().WithContext(ctx).Where("hw_id IN ? AND is_public = ?", hwIDs, true).Find(&allTasks).Error; err != nil {
			return nil, err
		}
	}

	var allDeadlines []models.Deadline
	if err := r.rw.ReadDB().WithContext(ctx).
		Joins("JOIN homework h ON h.hw_id = deadlines.homework_id").
		Where("deadlines.course_id = ? AND h.is_public = ?", courseID, true).
		Order("hard_deadline ASC").
		Find(&allDeadlines).Error; err != nil {
		return nil, err
	}

	tasksByHwID := make(map[uuid.UUID][]models.Task)
	for _, t := range allTasks {
		tasksByHwID[t.HwID] = append(tasksByHwID[t.HwID], t)
	}

	deadlinesByHwID := make(map[uuid.UUID][]models.Deadline)
	for _, d := range allDeadlines {
		deadlinesByHwID[d.HomeworkID] = append(deadlinesByHwID[d.HomeworkID], d)
	}

	details := make([]models.HomeworkWithTasks, len(homeworks))
	for i, hw := range homeworks {
		details[i] = models.HomeworkWithTasks{
			Homework:  hw,
			Tasks:     tasksByHwID[hw.HwID],
			Deadlines: deadlinesByHwID[hw.HwID],
		}
	}

	return &models.CourseInfo{
		Course:    course,
		Homeworks: details,
	}, nil
}

func (r *CourseRepository) GetLeaderboard(ctx context.Context, courseID string) ([]models.LeaderboardEntry, error) {
	courseUUID, err := uuid.Parse(courseID)
	var course models.Course
	if err != nil {
		if err := r.rw.ReadDB().WithContext(ctx).First(&course, "slug = ?", courseID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
	} else {
		if err := r.rw.ReadDB().WithContext(ctx).First(&course, "id = ?", courseUUID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
	}

	type taskScoreRow struct {
		UserID        uuid.UUID
		Username      string
		TaskID        uuid.UUID
		TaskTitle     string
		Score         int
		HomeworkID    uuid.UUID
		HomeworkTitle string
	}
	var rows []taskScoreRow
	err = r.rw.ReadDB().WithContext(ctx).
		Model(&models.UserRole{}).
		Select("u.id AS user_id, u.username, sts.task_id, t.title AS task_title, COALESCE(sts.score, 0) AS score, hw.hw_id AS homework_id, hw.title AS homework_title").
		Joins("JOIN users u ON u.id = user_roles.user_id").
		Joins("JOIN course_admin_permissions cap ON cap.role_id = user_roles.role_id AND cap.permission = ?", "task.submit").
		Joins("LEFT JOIN student_task_scores sts ON sts.student_id = user_roles.user_id AND sts.course_id = user_roles.course_id").
		Joins("LEFT JOIN tasks t ON t.task_id = sts.task_id AND t.is_public = true").
		Joins("LEFT JOIN homework hw ON hw.hw_id = t.hw_id AND hw.is_public = true").
		Where("user_roles.course_id = ?", course.ID).
		Order("u.username ASC, hw.position ASC, sts.task_id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	type userScores struct {
		Username   string
		TotalScore int
		Homeworks  []models.HomeworkScore
	}
	userMap := make(map[uuid.UUID]*userScores)
	var userOrder []uuid.UUID

	for _, row := range rows {
		us, ok := userMap[row.UserID]
		if !ok {
			us = &userScores{
				Username:  row.Username,
				Homeworks: make([]models.HomeworkScore, 0),
			}
			userMap[row.UserID] = us
			userOrder = append(userOrder, row.UserID)
		}
		if row.TaskID != uuid.Nil {
			var hw *models.HomeworkScore
			for i := range us.Homeworks {
				if us.Homeworks[i].HomeworkID == row.HomeworkID {
					hw = &us.Homeworks[i]
					break
				}
			}
			if hw == nil {
				us.Homeworks = append(us.Homeworks, models.HomeworkScore{
					HomeworkID:    row.HomeworkID,
					HomeworkTitle: row.HomeworkTitle,
					Tasks:         make([]models.TaskScore, 0),
				})
				hw = &us.Homeworks[len(us.Homeworks)-1]
			}

			hw.Tasks = append(hw.Tasks, models.TaskScore{
				TaskID: row.TaskID,
				Title:  row.TaskTitle,
				Score:  row.Score,
			})
			hw.TotalScore += row.Score
			us.TotalScore += row.Score
		}
	}

	entries := make([]models.LeaderboardEntry, 0, len(userOrder))
	for _, userID := range userOrder {
		us := userMap[userID]
		entries = append(entries, models.LeaderboardEntry{
			Username:   us.Username,
			TotalScore: us.TotalScore,
			Homeworks:  us.Homeworks,
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

func (r *CourseRepository) UpdateInviteCode(ctx context.Context, courseID uuid.UUID, code *string) error {
	return r.rw.WriteDB().WithContext(ctx).
		Model(&models.Course{}).
		Where("id = ?", courseID).
		Update("invite_code", code).Error
}
