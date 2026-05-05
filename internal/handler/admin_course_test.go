package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/handler"
)

type MockCourseRepo struct {
	mock.Mock
}

func (m *MockCourseRepo) GetCourses(ctx context.Context) ([]model.Course, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Course), args.Error(1)
}

func (m *MockCourseRepo) GetCourseByID(ctx context.Context, courseID string) (*model.Course, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

func (m *MockCourseRepo) CreateCourse(ctx context.Context, course model.Course) (*model.Course, error) {
	args := m.Called(ctx, course)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

func (m *MockCourseRepo) UpdateCourse(ctx context.Context, courseID string, course model.Course) (*model.Course, error) {
	args := m.Called(ctx, courseID, course)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Course), args.Error(1)
}

func (m *MockCourseRepo) DeleteCourse(ctx context.Context, courseID string) error {
	args := m.Called(ctx, courseID)
	return args.Error(0)
}

func (m *MockCourseRepo) GetCourseBoard(ctx context.Context, courseID string) (*model.TaskBoardSummary, bool, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*model.TaskBoardSummary), args.Bool(1), args.Error(2)
}

func setupEcho() *echo.Echo {
	e := echo.New()
	return e
}

func newRequest(method, path string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	return req, rec
}

func TestAdminCreateCourse_Success(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	body := map[string]interface{}{
		"name":   "Test Course",
		"slug":   "test-course",
		"status": "created",
		"type":   "public",
	}
	req, rec := newRequest(http.MethodPost, "/admin/courses", body)
	c := e.NewContext(req, rec)

	repo.On("GetCourseByID", mock.Anything, "test-course").Return(nil, nil)
	created := &model.Course{ID: uuid.New(), Name: "Test Course", Slug: "test-course"}
	repo.On("CreateCourse", mock.Anything, mock.AnythingOfType("model.Course")).Return(created, nil)

	err := h.AdminCreateCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestAdminCreateCourse_MissingName(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	body := map[string]interface{}{"slug": "test-course"}
	req, rec := newRequest(http.MethodPost, "/admin/courses", body)
	c := e.NewContext(req, rec)

	err := h.AdminCreateCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminCreateCourse_MissingSlug(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	body := map[string]interface{}{"name": "Test Course"}
	req, rec := newRequest(http.MethodPost, "/admin/courses", body)
	c := e.NewContext(req, rec)

	err := h.AdminCreateCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminCreateCourse_InvalidStatus(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	body := map[string]interface{}{"name": "X", "slug": "x", "status": "invalid"}
	req, rec := newRequest(http.MethodPost, "/admin/courses", body)
	c := e.NewContext(req, rec)

	err := h.AdminCreateCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminCreateCourse_InvalidType(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	body := map[string]interface{}{"name": "X", "slug": "x", "type": "invalid"}
	req, rec := newRequest(http.MethodPost, "/admin/courses", body)
	c := e.NewContext(req, rec)

	err := h.AdminCreateCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminCreateCourse_InvalidStartDate(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	body := map[string]interface{}{"name": "X", "slug": "x", "start_date": "not-a-date"}
	req, rec := newRequest(http.MethodPost, "/admin/courses", body)
	c := e.NewContext(req, rec)

	err := h.AdminCreateCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminCreateCourse_EndBeforeStart(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	body := map[string]interface{}{
		"name":       "X",
		"slug":       "x",
		"start_date": "2025-12-01",
		"end_date":   "2025-01-01",
	}
	req, rec := newRequest(http.MethodPost, "/admin/courses", body)
	c := e.NewContext(req, rec)

	err := h.AdminCreateCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminCreateCourse_SlugConflict(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	body := map[string]interface{}{"name": "X", "slug": "existing"}
	req, rec := newRequest(http.MethodPost, "/admin/courses", body)
	c := e.NewContext(req, rec)

	existing := &model.Course{Slug: "existing"}
	repo.On("GetCourseByID", mock.Anything, "existing").Return(existing, nil)

	err := h.AdminCreateCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAdminGetAllCourses_Success(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	req, rec := newRequest(http.MethodGet, "/admin/courses", nil)
	c := e.NewContext(req, rec)

	courses := []model.Course{{Name: "A"}, {Name: "B"}}
	repo.On("GetCourses", mock.Anything).Return(courses, nil)

	err := h.AdminGetAllCoursesHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminGetAllCourses_WithStatusFilter(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)
	req, rec := newRequest(http.MethodGet, "/admin/courses?status=hidden", nil)
	c := e.NewContext(req, rec)
	courses := []model.Course{
		{Name: "A", Status: "hidden"},
		{Name: "B", Status: "doreshka"},
	}
	repo.On("GetCourses", mock.Anything).Return(courses, nil)
	err := h.AdminGetAllCoursesHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var result []model.Course
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Len(t, result, 1)
	assert.Equal(t, "A", result[0].Name)
}

func TestAdminGetAllCourses_InvalidStatus(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)
	req, rec := newRequest(http.MethodGet, "/admin/courses?status=badstatus", nil)
	c := e.NewContext(req, rec)
	err := h.AdminGetAllCoursesHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminGetCourseByID_Success(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	id := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/courses/"+id.String(), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())

	course := &model.Course{ID: id, Name: "Test"}
	repo.On("GetCourseByID", mock.Anything, id.String()).Return(course, nil)

	err := h.AdminGetCourseByIDHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminGetCourseByID_InvalidUUID(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	req, rec := newRequest(http.MethodGet, "/admin/courses/not-a-uuid", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues("not-a-uuid")

	err := h.AdminGetCourseByIDHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminGetCourseByID_NotFound(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	id := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/courses/"+id.String(), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())

	repo.On("GetCourseByID", mock.Anything, id.String()).Return(nil, nil)

	err := h.AdminGetCourseByIDHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminDeleteCourse_Success(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	id := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/admin/courses/"+id.String(), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())

	course := &model.Course{ID: id}
	repo.On("GetCourseByID", mock.Anything, id.String()).Return(course, nil)
	repo.On("DeleteCourse", mock.Anything, id.String()).Return(nil)

	err := h.AdminDeleteCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAdminDeleteCourse_NotFound(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	id := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/admin/courses/"+id.String(), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())

	repo.On("GetCourseByID", mock.Anything, id.String()).Return(nil, nil)

	err := h.AdminDeleteCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminEditCourse_Success(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	id := uuid.New()
	newName := "Updated Name"
	body := map[string]interface{}{"name": newName}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+id.String(), body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())

	existing := &model.Course{ID: id, Name: "Old Name", Slug: "slug"}
	repo.On("GetCourseByID", mock.Anything, id.String()).Return(existing, nil)
	updated := &model.Course{ID: id, Name: newName, Slug: "slug"}
	repo.On("UpdateCourse", mock.Anything, id.String(), mock.AnythingOfType("model.Course")).Return(updated, nil)

	err := h.AdminEditCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminEditCourse_InvalidType(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	id := uuid.New()
	typeVal := model.CourseType("invalid")
	body := map[string]interface{}{"type": typeVal}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+id.String(), body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())

	existing := &model.Course{ID: id, Slug: "slug"}
	repo.On("GetCourseByID", mock.Anything, id.String()).Return(existing, nil)

	err := h.AdminEditCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminEditCourse_EndDateBeforeStart(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	id := uuid.New()
	body := map[string]interface{}{
		"start_date": "2025-12-01",
		"end_date":   "2025-06-01",
	}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+id.String(), body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())

	existing := &model.Course{ID: id, Slug: "slug"}
	repo.On("GetCourseByID", mock.Anything, id.String()).Return(existing, nil)

	err := h.AdminEditCourseHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminUpdateCourseStatus_Success(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)
	id := uuid.New()
	body := map[string]interface{}{"status": "doreshka"}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+id.String()+"/status", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())
	existing := &model.Course{ID: id, Slug: "slug", Status: "created"}
	repo.On("GetCourseByID", mock.Anything, id.String()).Return(existing, nil)
	updated := &model.Course{ID: id, Slug: "slug", Status: "started"}
	repo.On("UpdateCourse", mock.Anything, id.String(), mock.AnythingOfType("model.Course")).Return(updated, nil)
	err := h.AdminUpdateCourseStatusHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminUpdateCourseStatus_EmptyStatus(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	id := uuid.New()
	body := map[string]interface{}{"status": ""}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+id.String()+"/status", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())

	existing := &model.Course{ID: id, Slug: "slug"}
	repo.On("GetCourseByID", mock.Anything, id.String()).Return(existing, nil)

	err := h.AdminUpdateCourseStatusHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminUpdateCourseStatus_InvalidStatus(t *testing.T) {
	e := setupEcho()
	repo := new(MockCourseRepo)
	h := handler.NewAdminCourseHandler(repo)

	id := uuid.New()
	body := map[string]interface{}{"status": "flying"}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+id.String()+"/status", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId")
	c.SetParamValues(id.String())

	existing := &model.Course{ID: id, Slug: "slug"}
	repo.On("GetCourseByID", mock.Anything, id.String()).Return(existing, nil)

	err := h.AdminUpdateCourseStatusHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
