package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// setupEchoBoard — настраивает Echo с нужным хендлером
func setupEchoBoard() *echo.Echo {
	e := echo.New()
	e.GET("/api/courses/:courseId/board", GetCourseBoardHandler)
	return e
}

// resetBoardDB — сбрасывает boardData к исходному состоянию
func resetBoardDB() {
	boardData = map[string]TaskBoardSummary{
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
					},
					Tasks: []BoardTask{
						{ID: "t1", Name: "Arrays Sprint", Score: 20, ScoreEarned: 20, Stats: 0.82},
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
					},
					Tasks: []BoardTask{
						{ID: "mlops-t1", Name: "Data Pipeline", Score: 50, ScoreEarned: 45, Stats: 0.9},
					},
				},
			},
		},
	}
}

/* ============================================================
   Тесты GetCourseBoardHandler
   ============================================================ */

func TestGetCourseBoardHandler_ValidCourse(t *testing.T) {
	resetBoardDB()
	e := setupEchoBoard()

	req := httptest.NewRequest(http.MethodGet, "/api/courses/algorithms/board", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp TaskBoardSummary
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "Algorithms 101", resp.CourseName)
	assert.Equal(t, "in_progress", resp.CourseStatus)
	assert.Equal(t, 126, resp.SolvedScore)
	assert.Len(t, resp.Groups, 1)
	assert.Equal(t, "week-1", resp.Groups[0].ID)
	assert.Equal(t, "Checkpoint", resp.Groups[0].Deadlines[0].Label)
	assert.Equal(t, "Arrays Sprint", resp.Groups[0].Tasks[0].Name)
}

func TestGetCourseBoardHandler_CourseWithoutBoard(t *testing.T) {
	resetBoardDB()
	e := setupEchoBoard()

	// Добавим курс, которого нет в boardData, но есть в courseDB
	courseMu.Lock()
	courseDB["rust"] = Course{
		ID:           "rust",
		Name:         "Rust Core",
		Status:       "created",
		StartDate:    "2024-10-15",
		EndDate:      "2025-01-15",
		RepoTemplate: "git@test/rust.git",
		Description:  "Rust basics",
		URL:          "/course/rust",
	}
	courseMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/courses/rust/board", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp TaskBoardSummary
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "Rust Core", resp.CourseName)
	assert.Equal(t, "created", resp.CourseStatus)
	assert.Empty(t, resp.Groups)
}

func TestGetCourseBoardHandler_CourseNotFound(t *testing.T) {
	resetBoardDB()
	e := setupEchoBoard()

	req := httptest.NewRequest(http.MethodGet, "/api/courses/nonexistent/board", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "course not found", resp["error"])
}

func TestGetCourseBoardHandler_EmptyBoardData(t *testing.T) {
	resetDB()
	// Сохраним оригинальное состояние
	originalBoardData := boardData
	defer func() { boardData = originalBoardData }()

	// Очистим boardData
	boardData = make(map[string]TaskBoardSummary)

	e := setupEchoBoard()

	// Убедимся, что курс существует
	courseMu.RLock()
	_, exists := courseDB["algorithms"]
	courseMu.RUnlock()
	assert.True(t, exists, "course 'algorithms' must exist")

	req := httptest.NewRequest(http.MethodGet, "/api/courses/algorithms/board", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp TaskBoardSummary
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "Algorithms", resp.CourseName)
	assert.Empty(t, resp.Groups)
}

func TestGetCourseBoardHandler_JSONStructure(t *testing.T) {
	resetDB()
	resetBoardDB()
	
	// Добавим курс mlops в courseDB
	courseMu.Lock()
	courseDB["mlops"] = Course{
		ID:           "mlops",
		Name:         "MLOps Studio",
		Status:       "all_tasks_issued",
		StartDate:    "2024-09-01",
		EndDate:      "2024-11-30",
		RepoTemplate: "git@test/mlops.git",
		Description:  "MLOps course",
		URL:          "/course/mlops",
	}
	courseMu.Unlock()
	
	e := setupEchoBoard()

	req := httptest.NewRequest(http.MethodGet, "/api/courses/mlops/board", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Проверим, что ответ — валидный JSON
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Проверим наличие всех обязательных полей
	requiredFields := []string{"courseName", "courseStatus", "solvedScore", "maxScore", "solvedPercent", "groups"}
	for _, field := range requiredFields {
		assert.Contains(t, resp, field, "missing field: %s", field)
	}

	groups, ok := resp["groups"].([]interface{})
	assert.True(t, ok, "groups should be array")
	assert.GreaterOrEqual(t, len(groups), 1, "should have at least one group")

	group := groups[0].(map[string]interface{})
	assert.Contains(t, group, "id")
	assert.Contains(t, group, "name")
	assert.Contains(t, group, "deadlines")
	assert.Contains(t, group, "tasks")

	deadlines := group["deadlines"].([]interface{})
	assert.GreaterOrEqual(t, len(deadlines), 1)

	deadline := deadlines[0].(map[string]interface{})
	assert.Contains(t, deadline, "id")
	assert.Contains(t, deadline, "label")
	assert.Contains(t, deadline, "percent")
	assert.Contains(t, deadline, "dueAt")
	assert.Contains(t, deadline, "status")
}