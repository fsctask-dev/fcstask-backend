package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func setupEcho() *echo.Echo {
	e := echo.New()
	api := e.Group("/api")

	api.GET("/courses", GetCoursesHandler)
	api.GET("/courses/:courseId", GetCourseHandler)
	api.POST("/courses", CreateCourseHandler)
	api.PUT("/courses/:courseId", UpdateCourseHandler)

	return e
}

// plainReq - запрос БЕЗ авторизации
func plainReq(method, path string, body []byte) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func resetDB() {
	courseMu.Lock()
	defer courseMu.Unlock()

	courseDB = map[string]Course{
		"algorithms": {
			ID:           "algorithms",
			Name:         "Algorithms",
			Status:       "created",
			StartDate:    "2024-01-01",
			EndDate:      "2024-02-01",
			RepoTemplate: "git@test/repo.git",
			Description:  "test",
			URL:          "/course/algorithms",
		},
		"hidden": {
			ID:           "hidden",
			Name:         "Hidden",
			Status:       "hidden",
			StartDate:    "2024-01-01",
			EndDate:      "2024-02-01",
			RepoTemplate: "git@test/repo.git",
			Description:  "hidden",
			URL:          "/course/hidden",
		},
	}
}

func TestValidators(t *testing.T) {
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

func TestGetCourses_EmptyFilterResult(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses?status=finished", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var courses []Course
	json.Unmarshal(rec.Body.Bytes(), &courses)

	if len(courses) != 0 {
		t.Fatalf("expected 0 courses, got %d", len(courses))
	}
}

func TestGetCourses(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

func TestGetCourses_Filter(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses?status=hidden", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var courses []Course
	_ = json.Unmarshal(rec.Body.Bytes(), &courses)

	if len(courses) != 1 {
		t.Fatalf("expected 1 filtered course")
	}
}

func TestGetCourses_NoFilterAllVisible(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var courses []Course
	_ = json.Unmarshal(rec.Body.Bytes(), &courses)

	if len(courses) != 2 {
		t.Fatalf("expected 2 courses without filter")
	}
}

func TestGetCourse_OK(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses/algorithms", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

func TestGetCourse_NotFound(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses/unknown", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404")
	}
}

func TestCreateCourse_EmptyRepoTemplate(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"Test",
		"slug":"test",
		"status":"created",
		"startDate":"2025-01-01",
		"endDate":"2025-02-01",
		"description":"x"
	}`)

	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateCourse_Success(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"Go Course",
		"slug":"go-course",
		"status":"created",
		"startDate":"2024-03-01",
		"endDate":"2024-04-01",
		"repoTemplate":"git@test/go.git",
		"description":"Go basics"
	}`)

	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreateCourse_ValidationError(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodPost, "/api/courses", []byte(`{"slug":"a"}`))
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}
}

func TestCreateCourse_Conflict(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"Algorithms",
		"slug":"algorithms",
		"status":"created",
		"startDate":"2024-01-01",
		"endDate":"2024-02-01",
		"repoTemplate":"git@test",
		"description":"dup"
	}`)

	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409")
	}
}

func TestCreateCourse_InvalidDateRange(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"Bad",
		"slug":"bad-course",
		"status":"created",
		"startDate":"2024-02-01",
		"endDate":"2024-01-01",
		"repoTemplate":"git@test",
		"description":"x"
	}`)

	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}
}

func TestCreateCourse_MissingRequiredFields(t *testing.T) {
	resetDB()
	e := setupEcho()

	cases := []struct {
		name         string
		body         string
		wantErrField string
	}{
		{"no name", `{"slug":"test","status":"created","startDate":"2025-01-01","endDate":"2025-02-01","repoTemplate":"git@a","description":"x"}`, "name"},
		{"no slug", `{"name":"Test","status":"created","startDate":"2025-01-01","endDate":"2025-02-01","repoTemplate":"git@a","description":"x"}`, "slug"},
		{"no status", `{"name":"Test","slug":"test","startDate":"2025-01-01","endDate":"2025-02-01","repoTemplate":"git@a","description":"x"}`, "status"},
		{"no repoTemplate", `{"name":"Test","slug":"test","status":"created","startDate":"2025-01-01","endDate":"2025-02-01","description":"x"}`, "repoTemplate"},
		{"no description", `{"name":"Test","slug":"test","status":"created","startDate":"2025-01-01","endDate":"2025-02-01","repoTemplate":"git@a"}`, "description"},
		{"no startDate", `{"name":"Test","slug":"test","status":"created","endDate":"2025-02-01","repoTemplate":"git@a","description":"x"}`, "startDate"},
		{"no endDate", `{"name":"Test","slug":"test","status":"created","startDate":"2025-01-01","repoTemplate":"git@a","description":"x"}`, "endDate"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := plainReq(http.MethodPost, "/api/courses", []byte(tc.body))
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rec.Code)
			}

			var resp map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &resp)
			details, ok := resp["details"].([]interface{})
			if !ok {
				t.Fatal("expected details array")
			}
			found := false
			for _, d := range details {
				errMap := d.(map[string]interface{})
				if errMap["field"] == tc.wantErrField {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected validation error for field %q", tc.wantErrField)
			}
		})
	}
}

func TestCreateCourse_InvalidDates(t *testing.T) {
	resetDB()
	e := setupEcho()

	badDates := []struct {
		name      string
		startDate string
		endDate   string
	}{
		{"invalid start format", "01-01-2025", "2025-02-01"},
		{"invalid end format", "2025-01-01", "01-02-2025"},
		{"end before start", "2025-02-01", "2025-01-01"},
	}

	for _, tc := range badDates {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"name":"Test","slug":"test","status":"created","startDate":"%s","endDate":"%s","repoTemplate":"git@a","description":"x"}`, tc.startDate, tc.endDate)
			req := plainReq(http.MethodPost, "/api/courses", []byte(body))
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rec.Code)
			}
		})
	}
}

func TestCreateCourse_InvalidStatus(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{"name":"Test","slug":"test","status":"invalid","startDate":"2025-01-01","endDate":"2025-02-01","repoTemplate":"git@a","description":"x"}`)
	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateCourse_InvalidJSON(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{ "name": "test"`) // malformed
	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateCourse_ExtraFieldsIgnored(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"Extra",
		"slug":"extra",
		"status":"created",
		"startDate":"2024-03-01",
		"endDate":"2024-04-01",
		"repoTemplate":"git@test",
		"description":"test",
		"id":"should-ignore",
		"url":"should-ignore"
	}`)

	req := plainReq(http.MethodPost, "/api/courses", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var course Course
	json.Unmarshal(rec.Body.Bytes(), &course)
	if course.ID != "extra" {
		t.Error("ID should be set from slug")
	}
	if course.URL != "/course/extra" {
		t.Error("URL should be generated")
	}
}

func TestUpdateCourse_UpdateRepoTemplate(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{"repoTemplate":"git@updated"}`)
	req := plainReq(http.MethodPut, "/api/courses/algorithms", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var updated Course
	json.Unmarshal(rec.Body.Bytes(), &updated)

	if updated.RepoTemplate != "git@updated" {
		t.Fatalf("repoTemplate not updated")
	}
}

func TestUpdateCourse_DateRangeValidAfterPartial(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{"endDate":"2024-03-01"}`)
	req := plainReq(http.MethodPut, "/api/courses/algorithms", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUpdateCourse_AllFields(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"Updated",
		"status":"finished",
		"startDate":"2024-01-10",
		"endDate":"2024-02-10",
		"repoTemplate":"git@new",
		"description":"updated"
	}`)

	req := plainReq(http.MethodPut, "/api/courses/algorithms", body)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

func TestUpdateCourse_InvalidStatus(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodPut, "/api/courses/algorithms", []byte(`{"status":"bad"}`))
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}
}

func TestUpdateCourse_NotFound(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodPut, "/api/courses/unknown", []byte(`{"name":"x"}`))
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404")
	}
}

func TestUpdateCourse_PartialUpdate(t *testing.T) {
	resetDB()
	e := setupEcho()

	courseMu.RLock()
	original := courseDB["algorithms"]
	courseMu.RUnlock()

	body := []byte(`{
        "name": "New Name Only",
        "description": "New desc only"
    }`)

	req := plainReq(http.MethodPut, "/api/courses/algorithms", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var updated Course
	json.Unmarshal(rec.Body.Bytes(), &updated)

	if updated.Name != "New Name Only" {
		t.Errorf("name not updated, got %q", updated.Name)
	}
	if updated.Description != "New desc only" {
		t.Errorf("description not updated, got %q", updated.Description)
	}
	if updated.Status != original.Status {
		t.Errorf("status changed unexpectedly: %q → %q", original.Status, updated.Status)
	}
	if updated.StartDate != original.StartDate {
		t.Error("startDate should not change")
	}
}

func TestUpdateCourse_EmptyFieldsIgnored(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"name":"",
		"status":"",
		"description":""
	}`)

	req := plainReq(http.MethodPut, "/api/courses/algorithms", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var updated Course
	json.Unmarshal(rec.Body.Bytes(), &updated)

	if updated.Name == "" {
		t.Error("name should not be updated to empty")
	}
	if updated.Status == "" {
		t.Error("status should not be updated to empty")
	}
	if updated.Description == "" {
		t.Error("description should not be updated to empty")
	}
}

func TestUpdateCourse_InvalidDateRange(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{
		"startDate": "2025-03-01",
		"endDate":   "2025-02-01"
	}`)

	req := plainReq(http.MethodPut, "/api/courses/algorithms", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateCourse_InvalidDateFormat(t *testing.T) {
	resetDB()
	e := setupEcho()

	cases := []string{
		`{"startDate": "01-03-2025"}`,
		`{"endDate": "01-04-2025"}`,
	}

	for _, bodyStr := range cases {
		req := plainReq(http.MethodPut, "/api/courses/algorithms", []byte(bodyStr))
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	}
}

func TestUpdateCourse_IgnoreSlugChange(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{"slug": "new-slug"}`)

	req := plainReq(http.MethodPut, "/api/courses/algorithms", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	courseMu.RLock()
	course := courseDB["algorithms"]
	courseMu.RUnlock()

	if course.ID != "algorithms" {
		t.Error("ID should not change")
	}
}

func TestUpdateCourse_InvalidJSON(t *testing.T) {
	resetDB()
	e := setupEcho()

	body := []byte(`{ "name": "test"`)
	req := plainReq(http.MethodPut, "/api/courses/algorithms", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestIsValidDateRange_EqualDates(t *testing.T) {
	if isValidDateRange("2024-01-01", "2024-01-01") {
		t.Fatal("expected false when dates are equal")
	}
}