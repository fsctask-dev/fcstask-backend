package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Модели по контракту

type BoardDeadline struct {
	ID      string  `json:"id"`
	Label   string  `json:"label"`
	Percent float64 `json:"percent"`
	DueAt   string  `json:"dueAt"`
	Status  string  `json:"status"`
}

type BoardTask struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Score       int     `json:"score"`
	ScoreEarned int     `json:"scoreEarned"`
	Stats       float64 `json:"stats"`
	IsBonus     bool    `json:"isBonus,omitempty"`
	IsSpecial   bool    `json:"isSpecial,omitempty"`
	URL         string  `json:"url,omitempty"`
}

type BoardGroup struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	IsSpecial bool            `json:"isSpecial,omitempty"`
	StartedAt string          `json:"startedAt"`
	EndsAt    string          `json:"endsAt"`
	Deadlines []BoardDeadline `json:"deadlines"`
	Tasks     []BoardTask     `json:"tasks"`
}

type TaskBoardSummary struct {
	CourseName    string       `json:"courseName"`
	CourseStatus  string       `json:"courseStatus"`
	SolvedScore   int          `json:"solvedScore"`
	MaxScore      int          `json:"maxScore"`
	SolvedPercent int          `json:"solvedPercent"`
	Groups        []BoardGroup `json:"groups"`
}

// Пример данных
var boardData = map[string]TaskBoardSummary{
	"algorithms": {
		CourseName:    "Algorithms 101",
		CourseStatus:  "in_progress",
		SolvedScore:   126,
		MaxScore:      200,
		SolvedPercent: 63,
		Groups: []BoardGroup{
			{
				ID:        "week-1",
				Name:      "Week 1: Warmup",
				StartedAt: "2024-10-01T09:00:00Z",
				EndsAt:    "2024-10-14T18:00:00Z",
				Deadlines: []BoardDeadline{
					{ID: "d1", Label: "Checkpoint", Percent: 0.6, DueAt: "2024-09-20T18:00:00Z", Status: "expired"},
					{ID: "d2", Label: "Final", Percent: 1.0, DueAt: "2024-10-14T18:00:00Z", Status: "urgent"},
				},
				Tasks: []BoardTask{
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
				Deadlines: []BoardDeadline{
					{ID: "d3", Label: "Checkpoint", Percent: 0.5, DueAt: "2024-10-22T18:00:00Z", Status: "active"},
					{ID: "d4", Label: "Final", Percent: 1.0, DueAt: "2024-10-28T18:00:00Z", Status: "active"},
				},
				Tasks: []BoardTask{
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
		Groups: []BoardGroup{
			{
				ID:        "project-phase-1",
				Name:      "Project Phase 1",
				StartedAt: "2024-09-01T09:00:00Z",
				EndsAt:    "2024-10-15T18:00:00Z",
				Deadlines: []BoardDeadline{
					{ID: "mlops-d1", Label: "Proposal", Percent: 0.3, DueAt: "2024-09-15T18:00:00Z", Status: "expired"},
					{ID: "mlops-d2", Label: "MVP", Percent: 1.0, DueAt: "2024-10-15T18:00:00Z", Status: "expired"},
				},
				Tasks: []BoardTask{
					{ID: "mlops-t1", Name: "Data Pipeline", Score: 50, ScoreEarned: 45, Stats: 0.9},
					{ID: "mlops-t2", Name: "Model Training", Score: 50, ScoreEarned: 30, Stats: 0.6},
					{ID: "mlops-t3", Name: "Monitoring Setup", Score: 50, ScoreEarned: 20, Stats: 0.4},
				},
			},
		},
	},
}

// GET /api/courses/:courseId/board
func GetCourseBoardHandler(c echo.Context) error {
	courseID := c.Param("courseId")

	// Проверка: courseID не может быть пустым (иначе это ошибка маршрутизации)
	if courseID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "course ID is required",
		})
	}

	// Проверка существования курса
	courseMu.RLock()
	course, exists := courseDB[courseID]
	courseMu.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "course not found",
		})
	}

	// Возврат данных доски или пустой структуры
	if board, ok := boardData[courseID]; ok {
		return c.JSON(http.StatusOK, board)
	}

	return c.JSON(http.StatusOK, TaskBoardSummary{
		CourseName:   course.Name,
		CourseStatus: course.Status,
		Groups:       []BoardGroup{},
	})
}