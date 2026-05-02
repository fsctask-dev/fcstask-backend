package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
)

// MockCourseRepo - мок для репозитория курсов
type MockCourseRepo struct {
	mock.Mock
}

func (m *MockCourseRepo) Create(ctx context.Context, course *model.Course) error {
	args := m.Called(ctx, course)
	return args.Error(0)
}

func (m *MockCourseRepo) GetByID(ctx context.Context, id string) (*model.Course, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

func (m *MockCourseRepo) GetBySlug(ctx context.Context, slug string) (*model.Course, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

func (m *MockCourseRepo) GetAll(ctx context.Context, statusFilter string) ([]model.Course, error) {
	args := m.Called(ctx, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Course), args.Error(1)
}

func (m *MockCourseRepo) Update(ctx context.Context, course *model.Course) error {
	args := m.Called(ctx, course)
	return args.Error(0)
}

func (m *MockCourseRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestGetCourses_Success(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	courses := []model.Course{
		{ID: "algorithms", Name: "Algorithms", Status: "created", Type: model.CourseTypePublic},
		{ID: "hidden", Name: "Hidden", Status: "hidden", Type: model.CourseTypePrivate},
	}

	mockRepo.On("GetAll", mock.Anything, "").Return(courses, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/courses", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.GetCoursesHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []model.Course
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, "algorithms", resp[0].ID)
	assert.Equal(t, model.CourseTypePublic, resp[0].Type)

	mockRepo.AssertExpectations(t)
}

func TestGetCourses_WithStatusFilter(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	courses := []model.Course{
		{ID: "hidden", Name: "Hidden", Status: "hidden", Type: model.CourseTypePrivate},
	}

	mockRepo.On("GetAll", mock.Anything, "hidden").Return(courses, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/courses?status=hidden", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.GetCoursesHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []model.Course
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "hidden", resp[0].Status)

	mockRepo.AssertExpectations(t)
}

func TestGetCourses_GetAllError(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	mockRepo.On("GetAll", mock.Anything, "").Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/courses", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.GetCoursesHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "failed to fetch courses", resp["error"])

	mockRepo.AssertExpectations(t)
}

func TestGetCourse_Success(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	course := &model.Course{
		ID:           "algorithms",
		Name:         "Algorithms",
		Slug:         "algorithms",
		Status:       "created",
		Type:         model.CourseTypePublic,
		StartDate:    "2024-01-01",
		EndDate:      "2024-02-01",
		RepoTemplate: "git@test/repo.git",
		Description:  "test",
		URL:          "/course/algorithms",
	}

	mockRepo.On("GetByID", mock.Anything, "algorithms").Return(course, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/courses/algorithms", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("algorithms")

	handler := NewCourseHandler(mockRepo)
	err := handler.GetCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp model.Course
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "algorithms", resp.ID)
	assert.Equal(t, model.CourseTypePublic, resp.Type)

	mockRepo.AssertExpectations(t)
}

func TestGetCourse_NotFound(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	mockRepo.On("GetByID", mock.Anything, "unknown").Return(nil, fmt.Errorf("not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/courses/unknown", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("unknown")

	handler := NewCourseHandler(mockRepo)
	err := handler.GetCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "course not found", resp["error"])

	mockRepo.AssertExpectations(t)
}

func TestCreateCourse_Success(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	mockRepo.On("GetBySlug", mock.Anything, "go-course").Return(nil, fmt.Errorf("not found"))
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Course")).Return(nil)

	body := `{
		"name":"Go Course",
		"slug":"go-course",
		"status":"created",
		"type":"public",
		"startDate":"2024-03-01",
		"endDate":"2024-04-01",
		"repoTemplate":"git@test/go.git",
		"description":"Go basics"
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/courses", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.CreateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp model.Course
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "go-course", resp.ID)
	assert.Equal(t, model.CourseTypePublic, resp.Type)
	assert.Equal(t, "/course/go-course", resp.URL)

	mockRepo.AssertExpectations(t)
}

func TestCreateCourse_DefaultTypePrivate(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	mockRepo.On("GetBySlug", mock.Anything, "test").Return(nil, fmt.Errorf("not found"))
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(course *model.Course) bool {
		return course.Type == model.CourseTypePrivate
	})).Return(nil)

	body := `{
		"name":"Test",
		"slug":"test",
		"status":"created",
		"startDate":"2024-01-01",
		"endDate":"2024-02-01",
		"repoTemplate":"git@test",
		"description":"test"
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/courses", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.CreateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp model.Course
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, model.CourseTypePrivate, resp.Type)

	mockRepo.AssertExpectations(t)
}

func TestCreateCourse_ValidationError(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	body := `{"slug":"a"}`

	req := httptest.NewRequest(http.MethodPost, "/api/courses", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.CreateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "validation failed", resp["error"])
	assert.NotNil(t, resp["details"])
}

func TestCreateCourse_Conflict(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	existingCourse := &model.Course{ID: "algorithms", Slug: "algorithms"}
	mockRepo.On("GetBySlug", mock.Anything, "algorithms").Return(existingCourse, nil)

	body := `{
		"name":"Algorithms",
		"slug":"algorithms",
		"status":"created",
		"startDate":"2024-01-01",
		"endDate":"2024-02-01",
		"repoTemplate":"git@test",
		"description":"dup"
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/courses", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.CreateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "course with this slug already exists", resp["error"])

	mockRepo.AssertExpectations(t)
}

func TestCreateCourse_InvalidJSON(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	body := `{ "name": "test"`

	req := httptest.NewRequest(http.MethodPost, "/api/courses", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.CreateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "invalid JSON payload", resp["error"])
}

func TestCreateCourse_MissingRequiredFields(t *testing.T) {
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
			e := echo.New()
			mockRepo := new(MockCourseRepo)

			req := httptest.NewRequest(http.MethodPost, "/api/courses", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			handler := NewCourseHandler(mockRepo)
			err := handler.CreateCourseHandler(ctx)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var resp map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &resp)
			details, ok := resp["details"].([]interface{})
			assert.True(t, ok, "expected details array")

			found := false
			for _, d := range details {
				errMap := d.(map[string]interface{})
				if errMap["field"] == tc.wantErrField {
					found = true
					break
				}
			}
			assert.True(t, found, "expected validation error for field %q", tc.wantErrField)
		})
	}
}

func TestCreateCourse_InvalidType(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	body := `{"name":"Test","slug":"test","status":"created","type":"invalid","startDate":"2025-01-01","endDate":"2025-02-01","repoTemplate":"git@a","description":"x"}`

	req := httptest.NewRequest(http.MethodPost, "/api/courses", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.CreateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	details, ok := resp["details"].([]interface{})
	assert.True(t, ok)

	found := false
	for _, d := range details {
		errMap := d.(map[string]interface{})
		if errMap["field"] == "type" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestCreateCourse_CreateError(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	mockRepo.On("GetBySlug", mock.Anything, "test").Return(nil, fmt.Errorf("not found"))
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Course")).Return(errors.New("db error"))

	body := `{
		"name":"Test",
		"slug":"test",
		"status":"created",
		"startDate":"2024-01-01",
		"endDate":"2024-02-01",
		"repoTemplate":"git@test",
		"description":"test"
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/courses", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := NewCourseHandler(mockRepo)
	err := handler.CreateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "failed to create course", resp["error"])

	mockRepo.AssertExpectations(t)
}

func TestUpdateCourse_Success(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	existingCourse := &model.Course{
		ID:           "algorithms",
		Name:         "Algorithms",
		Slug:         "algorithms",
		Status:       "created",
		Type:         model.CourseTypePrivate,
		StartDate:    "2024-01-01",
		EndDate:      "2024-02-01",
		RepoTemplate: "git@test/repo.git",
		Description:  "test",
	}

	mockRepo.On("GetByID", mock.Anything, "algorithms").Return(existingCourse, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Course")).Return(nil)

	body := `{
		"name":"Updated Algorithms",
		"type":"public"
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/courses/algorithms", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("algorithms")

	handler := NewCourseHandler(mockRepo)
	err := handler.UpdateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp model.Course
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Updated Algorithms", resp.Name)
	assert.Equal(t, model.CourseTypePublic, resp.Type)

	mockRepo.AssertExpectations(t)
}

func TestUpdateCourse_NotFound(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	mockRepo.On("GetByID", mock.Anything, "unknown").Return(nil, fmt.Errorf("not found"))

	body := `{"name":"x"}`

	req := httptest.NewRequest(http.MethodPut, "/api/courses/unknown", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("unknown")

	handler := NewCourseHandler(mockRepo)
	err := handler.UpdateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockRepo.AssertExpectations(t)
}

func TestUpdateCourse_InvalidStatus(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	existingCourse := &model.Course{
		ID:     "algorithms",
		Status: "created",
	}

	mockRepo.On("GetByID", mock.Anything, "algorithms").Return(existingCourse, nil)

	body := `{"status":"invalid"}`

	req := httptest.NewRequest(http.MethodPut, "/api/courses/algorithms", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("algorithms")

	handler := NewCourseHandler(mockRepo)
	err := handler.UpdateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	mockRepo.AssertExpectations(t)
}

func TestUpdateCourse_InvalidType(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	existingCourse := &model.Course{
		ID:   "algorithms",
		Type: model.CourseTypePrivate,
	}

	mockRepo.On("GetByID", mock.Anything, "algorithms").Return(existingCourse, nil)

	body := `{"type":"invalid"}`

	req := httptest.NewRequest(http.MethodPut, "/api/courses/algorithms", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("algorithms")

	handler := NewCourseHandler(mockRepo)
	err := handler.UpdateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	mockRepo.AssertExpectations(t)
}

func TestUpdateCourse_InvalidDateRange(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	existingCourse := &model.Course{
		ID:        "algorithms",
		StartDate: "2024-01-01",
		EndDate:   "2024-02-01",
	}

	mockRepo.On("GetByID", mock.Anything, "algorithms").Return(existingCourse, nil)

	body := `{"startDate":"2025-03-01","endDate":"2025-02-01"}`

	req := httptest.NewRequest(http.MethodPut, "/api/courses/algorithms", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("algorithms")

	handler := NewCourseHandler(mockRepo)
	err := handler.UpdateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	mockRepo.AssertExpectations(t)
}

func TestUpdateCourse_PartialUpdate(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	existingCourse := &model.Course{
		ID:           "algorithms",
		Name:         "Algorithms",
		Status:       "created",
		Type:         model.CourseTypePrivate,
		StartDate:    "2024-01-01",
		EndDate:      "2024-02-01",
		RepoTemplate: "git@test/repo.git",
		Description:  "test",
	}

	mockRepo.On("GetByID", mock.Anything, "algorithms").Return(existingCourse, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Course")).Return(nil)

	body := `{
		"name": "New Name Only",
		"description": "New desc only"
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/courses/algorithms", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("algorithms")

	handler := NewCourseHandler(mockRepo)
	err := handler.UpdateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp model.Course
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "New Name Only", resp.Name)
	assert.Equal(t, "New desc only", resp.Description)
	assert.Equal(t, "created", resp.Status) // не должно измениться

	mockRepo.AssertExpectations(t)
}

func TestUpdateCourse_EmptyFieldsIgnored(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	existingCourse := &model.Course{
		ID:          "algorithms",
		Name:        "Algorithms",
		Status:      "created",
		Description: "test",
	}

	mockRepo.On("GetByID", mock.Anything, "algorithms").Return(existingCourse, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Course")).Return(nil)

	body := `{
		"name":"",
		"status":"",
		"description":""
	}`

	req := httptest.NewRequest(http.MethodPut, "/api/courses/algorithms", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("algorithms")

	handler := NewCourseHandler(mockRepo)
	err := handler.UpdateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp model.Course
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Algorithms", resp.Name)  // не должно измениться
	assert.Equal(t, "created", resp.Status)   // не должно измениться
	assert.Equal(t, "test", resp.Description) // не должно измениться

	mockRepo.AssertExpectations(t)
}

func TestUpdateCourse_UpdateError(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	existingCourse := &model.Course{
		ID:        "algorithms",
		StartDate: "2024-01-01",
		EndDate:   "2024-02-01",
	}

	mockRepo.On("GetByID", mock.Anything, "algorithms").Return(existingCourse, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Course")).Return(errors.New("db error"))

	body := `{"name":"Updated"}`

	req := httptest.NewRequest(http.MethodPut, "/api/courses/algorithms", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("algorithms")

	handler := NewCourseHandler(mockRepo)
	err := handler.UpdateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "failed to update course", resp["error"])

	mockRepo.AssertExpectations(t)
}

func TestUpdateCourse_InvalidJSON(t *testing.T) {
	e := echo.New()
	mockRepo := new(MockCourseRepo)

	body := `{ "name": "test"`

	req := httptest.NewRequest(http.MethodPut, "/api/courses/algorithms", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("algorithms")

	handler := NewCourseHandler(mockRepo)
	err := handler.UpdateCourseHandler(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "invalid JSON payload", resp["error"])
}

func TestIsValidDateRange_EqualDates(t *testing.T) {
	assert.False(t, isValidDateRange("2024-01-01", "2024-01-01"))
}
