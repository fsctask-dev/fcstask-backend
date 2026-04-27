package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type CourseHandler struct {
	courseRepo repo.CourseRepositoryInterface
}

func NewCourseHandler(courseRepo repo.CourseRepositoryInterface) *CourseHandler {
	return &CourseHandler{courseRepo: courseRepo}
}

type PostCourseRequest struct {
	Name         string           `json:"name"`
	Slug         string           `json:"slug"`
	Status       string           `json:"status"`
	Type         model.CourseType `json:"type"`
	StartDate    string           `json:"startDate"`
	EndDate      string           `json:"endDate"`
	RepoTemplate string           `json:"repoTemplate"`
	Description  string           `json:"description"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func isValidCourseStatus(status string) bool {
	valid := map[string]bool{
		"created":          true,
		"hidden":           true,
		"in_progress":      true,
		"all_tasks_issued": true,
		"doreshka":         true,
		"finished":         true,
	}
	return valid[status]
}

func isValidDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

func isValidDateRange(start, end string) bool {
	startDate, _ := time.Parse("2006-01-02", start)
	endDate, _ := time.Parse("2006-01-02", end)
	return endDate.After(startDate)
}

func (req *PostCourseRequest) Validate(c echo.Context) error {
	var errs []ValidationError

	if req.Name == "" {
		errs = append(errs, ValidationError{"name", "name is required"})
	}

	if req.Slug == "" {
		errs = append(errs, ValidationError{"slug", "slug is required"})
	}

	if req.Status == "" {
		errs = append(errs, ValidationError{"status", "status is required"})
	} else if !isValidCourseStatus(req.Status) {
		errs = append(errs, ValidationError{"status", "invalid status value"})
	}

	if req.Type != "" && req.Type != model.CourseTypePublic && req.Type != model.CourseTypePrivate {
		errs = append(errs, ValidationError{"type", "type must be 'public' or 'private'"})
	}

	if req.StartDate == "" {
		errs = append(errs, ValidationError{"startDate", "startDate is required"})
	} else if !isValidDate(req.StartDate) {
		errs = append(errs, ValidationError{"startDate", "startDate must be in format YYYY-MM-DD"})
	}

	if req.EndDate == "" {
		errs = append(errs, ValidationError{"endDate", "endDate is required"})
	} else if !isValidDate(req.EndDate) {
		errs = append(errs, ValidationError{"endDate", "endDate must be in format YYYY-MM-DD"})
	}

	if req.StartDate != "" && req.EndDate != "" && !isValidDateRange(req.StartDate, req.EndDate) {
		errs = append(errs, ValidationError{"dateRange", "endDate must be after startDate"})
	}

	if req.RepoTemplate == "" {
		errs = append(errs, ValidationError{"repoTemplate", "repoTemplate is required"})
	}

	if req.Description == "" {
		errs = append(errs, ValidationError{"description", "description is required"})
	}

	if len(errs) > 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "validation failed", "details": errs})
	}

	return nil
}

func (h *CourseHandler) GetCoursesHandler(c echo.Context) error {
	statusFilter := c.QueryParam("status")

	courses, err := h.courseRepo.GetAll(c.Request().Context(), statusFilter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch courses"})
	}

	return c.JSON(http.StatusOK, courses)
}

func (h *CourseHandler) GetCourseHandler(c echo.Context) error {
	courseID := c.Param("courseId")

	course, err := h.courseRepo.GetByID(c.Request().Context(), courseID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "course not found"})
	}

	return c.JSON(http.StatusOK, course)
}

func (h *CourseHandler) CreateCourseHandler(c echo.Context) error {
	var req PostCourseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
	}

	if err := req.Validate(c); err != nil {
		return err
	}

	_, err := h.courseRepo.GetBySlug(c.Request().Context(), req.Slug)
	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "course with this slug already exists"})
	}

	courseType := req.Type
	if courseType == "" {
		courseType = model.CourseTypePrivate
	}

	course := model.Course{
		ID:           req.Slug,
		Name:         req.Name,
		Slug:         req.Slug,
		Status:       req.Status,
		Type:         courseType,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		RepoTemplate: req.RepoTemplate,
		Description:  req.Description,
		URL:          "/course/" + req.Slug,
	}

	if err := h.courseRepo.Create(c.Request().Context(), &course); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create course"})
	}

	return c.JSON(http.StatusCreated, course)
}

func (h *CourseHandler) UpdateCourseHandler(c echo.Context) error {
	courseID := c.Param("courseId")

	course, err := h.courseRepo.GetByID(c.Request().Context(), courseID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "course not found"})
	}

	var req PostCourseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
	}

	if req.Status != "" && !isValidCourseStatus(req.Status) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid status value"})
	}

	if req.Type != "" && req.Type != model.CourseTypePublic && req.Type != model.CourseTypePrivate {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "type must be 'public' or 'private'"})
	}

	if req.StartDate != "" && !isValidDate(req.StartDate) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "startDate must be in format YYYY-MM-DD"})
	}

	if req.EndDate != "" && !isValidDate(req.EndDate) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "endDate must be in format YYYY-MM-DD"})
	}

	if req.Name != "" {
		course.Name = req.Name
	}
	if req.Status != "" {
		course.Status = req.Status
	}
	if req.Type != "" {
		course.Type = req.Type
	}
	if req.StartDate != "" {
		course.StartDate = req.StartDate
	}
	if req.EndDate != "" {
		course.EndDate = req.EndDate
	}
	if req.RepoTemplate != "" {
		course.RepoTemplate = req.RepoTemplate
	}
	if req.Description != "" {
		course.Description = req.Description
	}

	if !isValidDateRange(course.StartDate, course.EndDate) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "endDate must be after startDate"})
	}

	if err := h.courseRepo.Update(c.Request().Context(), course); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update course"})
	}

	return c.JSON(http.StatusOK, course)
}
