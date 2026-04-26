package repo

import (
	"context"
	"sync"

	models "fcstask-backend/internal/db/model"
)

type CourseRepositoryInterface interface {
	GetCourses(ctx context.Context) ([]models.Course, error)
	GetCourseByID(ctx context.Context, courseID string) (*models.Course, error)
	CreateCourse(ctx context.Context, course models.Course) (*models.Course, error)
	UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error)
	GetCourseBoard(ctx context.Context, courseID string) (*models.TaskBoardSummary, bool, error)
}

type InMemoryCourseRepository struct {
	mu      sync.RWMutex
	courses map[string]models.Course
	boards  map[string]models.TaskBoardSummary
}

func NewInMemoryCourseRepository() CourseRepositoryInterface {
	return &InMemoryCourseRepository{
		courses: defaultCourses(),
		boards:  defaultCourseBoards(),
	}
}

func (r *InMemoryCourseRepository) GetCourses(ctx context.Context) ([]models.Course, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	courses := make([]models.Course, 0, len(r.courses))
	for _, course := range r.courses {
		courses = append(courses, course)
	}
	return courses, nil
}

func (r *InMemoryCourseRepository) GetCourseByID(ctx context.Context, courseID string) (*models.Course, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	course, ok := r.courses[courseID]
	if !ok {
		return nil, nil
	}
	return &course, nil
}

func (r *InMemoryCourseRepository) CreateCourse(ctx context.Context, course models.Course) (*models.Course, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.courses[course.ID] = course
	return &course, nil
}

func (r *InMemoryCourseRepository) UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.courses[courseID] = course
	return &course, nil
}

func (r *InMemoryCourseRepository) GetCourseBoard(ctx context.Context, courseID string) (*models.TaskBoardSummary, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	board, ok := r.boards[courseID]
	if !ok {
		return nil, false, nil
	}
	return &board, true, nil
}

func defaultCourses() map[string]models.Course {
	return map[string]models.Course{
		"algorithms": {
			ID:           "algorithms",
			Name:         "Algorithms 101",
			Status:       "in_progress",
			StartDate:    "2024-10-01",
			EndDate:      "2024-12-20",
			RepoTemplate: "git@gitlab.local/algorithms-template.git",
			Description:  "Основы алгоритмов и структур данных",
			URL:          "/course/algorithms",
		},
		"mlops": {
			ID:           "mlops",
			Name:         "MLOps Studio",
			Status:       "all_tasks_issued",
			StartDate:    "2024-09-01",
			EndDate:      "2024-11-30",
			RepoTemplate: "git@gitlab.local/mlops-template.git",
			Description:  "Продвинутые практики MLOps",
			URL:          "/course/mlops",
		},
		"rust": {
			ID:           "rust",
			Name:         "Rust Core",
			Status:       "created",
			StartDate:    "2024-10-15",
			EndDate:      "2025-01-15",
			RepoTemplate: "git@gitlab.local/rust-template.git",
			Description:  "Основы системного программирования на Rust",
			URL:          "/course/rust",
		},
		"golang": {
			ID:           "golang",
			Name:         "Go Lab",
			Status:       "finished",
			StartDate:    "2024-08-01",
			EndDate:      "2024-10-31",
			RepoTemplate: "git@gitlab.local/golang-template.git",
			Description:  "Практикум по языку Go",
			URL:          "/course/golang",
		},
		"advanced-cpp": {
			ID:           "advanced-cpp",
			Name:         "Advanced C++",
			Status:       "in_progress",
			StartDate:    "2024-10-01",
			EndDate:      "2024-12-20",
			RepoTemplate: "git@gitlab.local/advanced-cpp-template.git",
			Description:  "Продвинутые концепции C++",
			URL:          "/course/advanced-cpp",
		},
		"advanced-python": {
			ID:           "advanced-python",
			Name:         "Advanced Python",
			Status:       "created",
			StartDate:    "2024-11-01",
			EndDate:      "2025-02-28",
			RepoTemplate: "git@gitlab.local/advanced-python-template.git",
			Description:  "Продвинутый анализ данных на Python",
			URL:          "/course/advanced-python",
		},
	}
}

func defaultCourseBoards() map[string]models.TaskBoardSummary {
	return map[string]models.TaskBoardSummary{
		"algorithms": {
			CourseName:    "Algorithms 101",
			CourseStatus:  "in_progress",
			SolvedScore:   126,
			MaxScore:      200,
			SolvedPercent: 63,
			Groups: []models.BoardGroup{
				{
					ID:        "week-1",
					Name:      "Week 1: Warmup",
					StartedAt: "2024-10-01T09:00:00Z",
					EndsAt:    "2024-10-14T18:00:00Z",
					Deadlines: []models.BoardDeadline{
						{ID: "d1", Label: "Checkpoint", Percent: 0.6, DueAt: "2024-09-20T18:00:00Z", Status: "expired"},
						{ID: "d2", Label: "Final", Percent: 1.0, DueAt: "2024-10-14T18:00:00Z", Status: "urgent"},
					},
					Tasks: []models.BoardTask{
						{ID: "t1", Name: "Arrays Sprint", Score: 20, ScoreEarned: 20, Stats: 0.82},
						{ID: "t2", Name: "Stack Trace", Score: 25, ScoreEarned: 10, Stats: 0.64},
						{ID: "t3", Name: "Sorting Arena", Score: 30, ScoreEarned: 0, Stats: 0.38, IsSpecial: true},
					},
				},
				{
					ID:        "week-2",
					Name:      "Week 2: Graphs",
					IsSpecial: true,
					StartedAt: "2024-10-15T09:00:00Z",
					EndsAt:    "2024-10-28T18:00:00Z",
					Deadlines: []models.BoardDeadline{
						{ID: "d3", Label: "Checkpoint", Percent: 0.5, DueAt: "2024-10-22T18:00:00Z", Status: "active"},
						{ID: "d4", Label: "Final", Percent: 1.0, DueAt: "2024-10-28T18:00:00Z", Status: "active"},
					},
					Tasks: []models.BoardTask{
						{ID: "t4", Name: "Bridge Builder", Score: 40, ScoreEarned: 25, Stats: 0.57},
						{ID: "t5", Name: "Shortest Path Lab", Score: 30, ScoreEarned: 0, Stats: 0.44},
						{ID: "t6", Name: "Bonus Relay", Score: 10, ScoreEarned: 12, Stats: 0.91, IsBonus: true},
					},
				},
			},
		},
		"mlops": {
			CourseName:    "MLOps Studio",
			CourseStatus:  "all_tasks_issued",
			SolvedScore:   95,
			MaxScore:      150,
			SolvedPercent: 63,
			Groups: []models.BoardGroup{
				{
					ID:        "project-phase-1",
					Name:      "Project Phase 1",
					StartedAt: "2024-09-01T09:00:00Z",
					EndsAt:    "2024-10-15T18:00:00Z",
					Deadlines: []models.BoardDeadline{
						{ID: "mlops-d1", Label: "Proposal", Percent: 0.3, DueAt: "2024-09-15T18:00:00Z", Status: "expired"},
						{ID: "mlops-d2", Label: "MVP", Percent: 1.0, DueAt: "2024-10-15T18:00:00Z", Status: "expired"},
					},
					Tasks: []models.BoardTask{
						{ID: "mlops-t1", Name: "Data Pipeline", Score: 50, ScoreEarned: 45, Stats: 0.9},
						{ID: "mlops-t2", Name: "Model Training", Score: 50, ScoreEarned: 30, Stats: 0.6},
						{ID: "mlops-t3", Name: "Monitoring Setup", Score: 50, ScoreEarned: 20, Stats: 0.4},
					},
				},
			},
		},
	}
}
