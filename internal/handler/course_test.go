package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type testCourseRepository struct {
	courses map[string]models.Course
	boards  map[string]models.TaskBoardSummary
}

func (r *testCourseRepository) GetCourses(ctx context.Context) ([]models.Course, error) {
	courses := make([]models.Course, 0, len(r.courses))
	for _, course := range r.courses {
		courses = append(courses, course)
	}
	return courses, nil
}

func (r *testCourseRepository) GetCourseByID(ctx context.Context, courseID string) (*models.Course, error) {
	course, ok := r.courses[courseID]
	if !ok {
		return nil, nil
	}
	return &course, nil
}

func (r *testCourseRepository) CreateCourse(ctx context.Context, course models.Course) (*models.Course, error) {
	if course.ID == uuid.Nil {
		course.ID = uuid.New()
	}
	r.courses[course.Slug] = course
	return &course, nil
}

func (r *testCourseRepository) UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error) {
	r.courses[courseID] = course
	return &course, nil
}

func (r *testCourseRepository) DeleteCourse(ctx context.Context, courseID string) error {
	delete(r.courses, courseID)
	return nil
}

func (r *testCourseRepository) GetCourseBoard(ctx context.Context, courseID string) (*models.TaskBoardSummary, bool, error) {
	board, ok := r.boards[courseID]
	if !ok {
		return nil, false, nil
	}
	return &board, true, nil
}

var testCourses *testCourseRepository

func setupEcho() *echo.Echo {
	e := echo.New()
	api := e.Group("/api")
	if testCourses == nil {
		resetDB()
	}
	courseHandler := NewCourseHandler(service.NewCourseService(testCourses))

	api.GET("/courses", courseHandler.GetCourses)
	api.GET("/courses/:courseId", courseHandler.GetCourse)
	api.POST("/courses", courseHandler.CreateCourse)
	api.PUT("/courses/:courseId", courseHandler.UpdateCourse)
	api.GET("/courses/:courseId/board", courseHandler.GetCourseBoard)

	return e
}

func plainReq(method, path string, body []byte) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func resetDB() {
	testCourses = &testCourseRepository{
		courses: map[string]models.Course{
			"algorithms": {
				ID:           testCourseUUID(),
				Name:         "Algorithms",
				Slug:         "algorithms",
				Status:       "created",
				Type:         models.CourseTypePublic,
				StartDate:    testCourseDate("2024-01-01"),
				EndDate:      testCourseDate("2024-02-01"),
				RepoTemplate: testStringPtr("git@test/repo.git"),
				Description:  testStringPtr("test"),
				URL:          "/course/algorithms",
			},
			"hidden": {
				ID:           testCourseUUID(),
				Name:         "Hidden",
				Slug:         "hidden",
				Status:       "hidden",
				Type:         models.CourseTypePrivate,
				StartDate:    testCourseDate("2024-01-01"),
				EndDate:      testCourseDate("2024-02-01"),
				RepoTemplate: testStringPtr("git@test/repo.git"),
				Description:  testStringPtr("hidden"),
				URL:          "/course/hidden",
			},
		},
		boards: map[string]models.TaskBoardSummary{},
	}
}

func TestCourseValidation(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() bool
		expected bool
	}{
		{"valid status", func() bool { return isValidCourseStatus("created") }, true},
		{"invalid status", func() bool { return isValidCourseStatus("broken") }, false},
		{"valid date", func() bool { return isValidDate("2024-01-01") }, true},
		{"invalid date format", func() bool { return isValidDate("01-01-2024") }, false},
		{"valid date range", func() bool { return isValidDateRange("2024-01-01", "2024-01-02") }, true},
		{"invalid date range", func() bool { return isValidDateRange("2024-01-02", "2024-01-01") }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestCourseHandler_GetCourses(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var courses []Course
	if err := json.Unmarshal(rec.Body.Bytes(), &courses); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(courses) != 2 {
		t.Fatalf("expected 2 courses, got %d", len(courses))
	}
}

func TestCourseHandler_GetCourses_Filter(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses?status=hidden", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var courses []Course
	if err := json.Unmarshal(rec.Body.Bytes(), &courses); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(courses) != 1 || courses[0].Status != "hidden" {
		t.Fatalf("unexpected filtered courses: %#v", courses)
	}
}

func TestCourseHandler_GetCourse_OK(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses/algorithms", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCourseHandler_GetCourse_NotFound(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses/unknown", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCourseHandler_CreateCourse_Success(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"Go Course",
		"slug":"go-course",
		"status":"created",
		"type":"public",
		"startDate":"2024-03-01",
		"endDate":"2024-04-01",
		"repoTemplate":"git@test/go.git",
		"description":"test"
	}`)
	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var course Course
	if err := json.Unmarshal(rec.Body.Bytes(), &course); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if course.ID == uuid.Nil {
		t.Fatal("expected generated UUID")
	}
	if course.Slug != "go-course" || course.URL != "/course/go-course" {
		t.Fatalf("unexpected course: %#v", course)
	}
}

func TestCourseHandler_CreateCourse_Conflict(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"Algorithms",
		"slug":"algorithms",
		"status":"created",
		"startDate":"2024-03-01",
		"endDate":"2024-04-01",
		"repoTemplate":"git@test/go.git",
		"description":"test"
	}`)
	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestCourseHandler_CreateCourse_InvalidDateRange(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"Bad Course",
		"slug":"bad-course",
		"status":"created",
		"startDate":"2024-04-01",
		"endDate":"2024-03-01",
		"repoTemplate":"git@test/go.git",
		"description":"test"
	}`)
	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCourseHandler_UpdateCourse_PartialUpdate(t *testing.T) {
	resetDB()
	e := setupEcho()
	original := testCourses.courses["algorithms"]

	body := []byte(`{"name":"New Name Only","description":"New desc only"}`)
	req := plainReq(http.MethodPut, "/api/courses/algorithms", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var updated Course
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if updated.Name != "New Name Only" {
		t.Errorf("name not updated, got %q", updated.Name)
	}
	if updated.Description == nil || *updated.Description != "New desc only" {
		t.Errorf("description not updated, got %v", updated.Description)
	}
	if !updated.StartDate.Equal(*original.StartDate) {
		t.Error("startDate should not change")
	}
}

func TestCourseHandler_UpdateCourse_InvalidStatus(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodPut, "/api/courses/algorithms", []byte(`{"status":"bad"}`))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCourseHandler_GetCourseBoard_Empty(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses/algorithms/board", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var board TaskBoardSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &board); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if board.CourseName != "Algorithms" || len(board.Groups) != 0 {
		t.Fatalf("unexpected board: %#v", board)
	}
}

func testCourseDate(value string) *time.Time {
	parsed, _ := time.Parse("2006-01-02", value)
	return &parsed
}

func testStringPtr(value string) *string {
	return &value
}

func testCourseUUID() uuid.UUID {
	return uuid.New()
}
