package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
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

	var response PaginatedResponse
	json.Unmarshal(rec.Body.Bytes(), &response)
	
	courses := response.Data.([]interface{})
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

	var response PaginatedResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	
	courses := response.Data.([]interface{})
	if len(courses) != 1 {
		t.Fatalf("expected 1 filtered course, got %d", len(courses))
	}
}

func TestGetCourses_NoFilterAllVisible(t *testing.T) {
	resetDB()
	e := setupEcho()

	req := plainReq(http.MethodGet, "/api/courses", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var response PaginatedResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	
	courses := response.Data.([]interface{})
	if len(courses) != 2 {
		t.Fatalf("expected 2 courses without filter, got %d", len(courses))
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


func TestGetCourses_WithPagination(t *testing.T) {
	resetDB()
	courseMu.Lock()
	for i := 1; i <= 25; i++ {
		courseDB[fmt.Sprintf("course-%d", i)] = Course{
			ID:           fmt.Sprintf("course-%d", i),
			Name:         fmt.Sprintf("Course %d", i),
			Status:       "created",
			StartDate:    "2024-01-01",
			EndDate:      "2024-02-01",
			RepoTemplate: "git@test/repo.git",
			Description:  "test",
			URL:          fmt.Sprintf("/course/course-%d", i),
		}
	}
	courseMu.Unlock()
	
	e := setupEcho()
	
	tests := []struct {
		name           string
		limit          string
		offset         string
		expectedCount  int
		expectedTotal  int
		expectedStatus int
	}{
		{
			name:           "default pagination (limit 20, offset 0)",
			limit:          "",
			offset:         "",
			expectedCount:  20,
			expectedTotal:  27, // 2 из resetDB + 25 добавленных
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom limit 5",
			limit:          "5",
			offset:         "0",
			expectedCount:  5,
			expectedTotal:  27,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom limit 5 offset 10",
			limit:          "5",
			offset:         "10",
			expectedCount:  5,
			expectedTotal:  27,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "offset beyond total",
			limit:          "10",
			offset:         "100",
			expectedCount:  0,
			expectedTotal:  27,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "limit exceeds max (should be capped at 100)",
			limit:          "200",
			offset:         "0",
			expectedCount:  27, // всего 27 курсов
			expectedTotal:  27,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid limit (should use default)",
			limit:          "invalid",
			offset:         "0",
			expectedCount:  20,
			expectedTotal:  27,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "negative offset (should use 0)",
			limit:          "10",
			offset:         "-5",
			expectedCount:  10,
			expectedTotal:  27,
			expectedStatus: http.StatusOK,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/courses"
			if tt.limit != "" || tt.offset != "" {
				url += "?"
				if tt.limit != "" {
					url += "limit=" + tt.limit
				}
				if tt.offset != "" {
					if tt.limit != "" {
						url += "&"
					}
					url += "offset=" + tt.offset
				}
			}
			
			req := plainReq(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			var response PaginatedResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			
			data, ok := response.Data.([]interface{})
			assert.True(t, ok, "response should have 'data' field")
			assert.Equal(t, tt.expectedCount, len(data), "unexpected data count")
			
			assert.Equal(t, tt.expectedTotal, response.Pagination.Total, "unexpected total")
			
			if tt.limit == "200" {
				assert.Equal(t, 100, response.Pagination.Limit, "limit should be capped at 100")
			} else if tt.limit == "invalid" {
				assert.Equal(t, 20, response.Pagination.Limit, "invalid limit should use default 20")
			} else if tt.limit != "" {
				expectedLimit := tt.expectedCount
				if tt.expectedCount > 20 && tt.limit != "200" {
					expectedLimit, _ = strconv.Atoi(tt.limit)
				}
				if expectedLimit <= 100 {
					assert.Equal(t, expectedLimit, response.Pagination.Limit)
				}
			}
		})
	}
}

func TestGetCourses_WithPaginationAndFilter(t *testing.T) {
	resetDB()
	e := setupEcho()
	
	courseMu.Lock()
	courseDB["filtered-1"] = Course{
		ID:           "filtered-1",
		Name:         "Filtered 1",
		Status:       "in_progress",
		StartDate:    "2024-01-01",
		EndDate:      "2024-02-01",
		RepoTemplate: "git@test/repo.git",
		Description:  "test",
		URL:          "/course/filtered-1",
	}
	courseDB["filtered-2"] = Course{
		ID:           "filtered-2",
		Name:         "Filtered 2",
		Status:       "in_progress",
		StartDate:    "2024-01-01",
		EndDate:      "2024-02-01",
		RepoTemplate: "git@test/repo.git",
		Description:  "test",
		URL:          "/course/filtered-2",
	}
	courseMu.Unlock()
	
	req := plainReq(http.MethodGet, "/api/courses?status=in_progress&limit=1&offset=1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	
	assert.Equal(t, http.StatusOK, rec.Code)
	
	var response PaginatedResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	data := response.Data.([]interface{})
	assert.Equal(t, 1, len(data), "should return 1 course with offset=1")
	assert.Equal(t, 3, response.Pagination.Total, "total should be 3 courses with status 'in_progress'")
}

func TestGetCourses_PaginationNextPrev(t *testing.T) {
	resetDB()
	courseMu.Lock()
	for i := 1; i <= 15; i++ {
		courseDB[fmt.Sprintf("pagination-test-%d", i)] = Course{
			ID:           fmt.Sprintf("pagination-test-%d", i),
			Name:         fmt.Sprintf("Pagination Test %d", i),
			Status:       "created",
			StartDate:    "2024-01-01",
			EndDate:      "2024-02-01",
			RepoTemplate: "git@test/repo.git",
			Description:  "test",
			URL:          fmt.Sprintf("/course/pagination-test-%d", i),
		}
	}
	courseMu.Unlock()
	
	e := setupEcho()
	
	// Тест для первой страницы (next должен быть)
	req := plainReq(http.MethodGet, "/api/courses?limit=5&offset=0", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	
	var response PaginatedResponse
	json.Unmarshal(rec.Body.Bytes(), &response)
	
	assert.NotNil(t, response.Pagination.Next, "next should exist for first page")
	assert.Nil(t, response.Pagination.Prev, "prev should be nil for first page")
	assert.Equal(t, 5, *response.Pagination.Next, "next offset should be 5")
	
	// Тест для последней страницы (next должен быть nil)
	req2 := plainReq(http.MethodGet, "/api/courses?limit=5&offset=15", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	
	var response2 PaginatedResponse
	json.Unmarshal(rec2.Body.Bytes(), &response2)
	
	assert.Nil(t, response2.Pagination.Next, "next should be nil for last page")
	assert.NotNil(t, response2.Pagination.Prev, "prev should exist for last page")
	assert.Equal(t, 10, *response2.Pagination.Prev, "prev offset should be 10")
}