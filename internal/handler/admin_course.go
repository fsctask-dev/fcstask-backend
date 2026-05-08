package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type AdminCourseHandler struct {
	courseRepo repo.CourseRepositoryInterface
}

func NewAdminCourseHandler(courseRepo repo.CourseRepositoryInterface) *AdminCourseHandler {
	return &AdminCourseHandler{courseRepo: courseRepo}
}

type UpdateCourseInfoRequest struct {
	Name         *string           `json:"name"`
	Description  *string           `json:"description"`
	RepoTemplate *string           `json:"repo_template"`
	Type         *model.CourseType `json:"type"`
	StartDate    *string           `json:"start_date"`
	EndDate      *string           `json:"end_date"`
}

type UpdateCourseStatusRequest struct {
	Status string `json:"status"`
}

const dateLayout = "2006-01-02"

func parseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}

// POST /admin/courses
func (h *AdminCourseHandler) AdminCreateCourseHandler(c echo.Context) error {
	var req struct {
		Name         string           `json:"name"`
		Slug         string           `json:"slug"`
		Description  *string          `json:"description"`
		Status       string           `json:"status"`
		Type         model.CourseType `json:"type"`
		StartDate    *string          `json:"start_date"`
		EndDate      *string          `json:"end_date"`
		RepoTemplate *string          `json:"repo_template"`
		URL          string           `json:"url"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.Name == "" {
		return badRequest(c, "Name is required")
	}
	if req.Slug == "" {
		return badRequest(c, "Slug is required")
	}
	if req.Status != "" && !isValidCourseStatus(req.Status) {
		return badRequest(c, "Invalid status value")
	}
	if req.Type != "" && req.Type != model.CourseTypePublic && req.Type != model.CourseTypePrivate {
		return badRequest(c, "Type must be 'public' or 'private'")
	}

	course := model.Course{
		Name:         req.Name,
		Slug:         req.Slug,
		Description:  req.Description,
		Status:       req.Status,
		Type:         req.Type,
		RepoTemplate: req.RepoTemplate,
		URL:          req.URL,
	}

	if req.Status == "" {
		course.Status = "created"
	}
	if req.Type == "" {
		course.Type = model.CourseTypePrivate
	}

	if req.StartDate != nil {
		t, err := parseDate(*req.StartDate)
		if err != nil {
			return badRequest(c, "start_date must be in format YYYY-MM-DD")
		}
		course.StartDate = &t
	}
	if req.EndDate != nil {
		t, err := parseDate(*req.EndDate)
		if err != nil {
			return badRequest(c, "end_date must be in format YYYY-MM-DD")
		}
		course.EndDate = &t
	}
	if course.StartDate != nil && course.EndDate != nil && !course.EndDate.After(*course.StartDate) {
		return badRequest(c, "end_date must be after start_date")
	}

	existing, err := h.courseRepo.GetCourseByID(c.Request().Context(), req.Slug)
	if err != nil {
		return internalError(c, "Failed to check existing course")
	}
	if existing != nil {
		return conflict(c, "Course with this slug already exists")
	}

	created, err := h.courseRepo.CreateCourse(c.Request().Context(), course)
	if err != nil {
		return internalError(c, "Failed to create course")
	}

	return c.JSON(http.StatusCreated, created)
}

// GET /admin/courses
func (h *AdminCourseHandler) AdminGetAllCoursesHandler(c echo.Context) error {
	statusFilter := c.QueryParam("status")
	if statusFilter != "" && !isValidCourseStatus(statusFilter) {
		return badRequest(c, "Invalid status value")
	}

	courses, err := h.courseRepo.GetCourses(c.Request().Context())
	if err != nil {
		return internalError(c, "Failed to fetch courses")
	}

	if statusFilter != "" {
		filtered := make([]model.Course, 0)
		for _, course := range courses {
			if course.Status == statusFilter {
				filtered = append(filtered, course)
			}
		}
		courses = filtered
	}

	return c.JSON(http.StatusOK, courses)
}

// GET /admin/courses/:courseId
func (h *AdminCourseHandler) AdminGetCourseByIDHandler(c echo.Context) error {
	courseID := c.Param("courseId")
	if _, err := uuid.Parse(courseID); err != nil {
		return badRequest(c, "Invalid course ID")
	}

	course, err := h.courseRepo.GetCourseByID(c.Request().Context(), courseID)
	if err != nil {
		return internalError(c, "Failed to fetch course")
	}
	if course == nil {
		return notFound(c, "Course not found")
	}

	return c.JSON(http.StatusOK, course)
}

// GET /admin/courses/slug/:slug
func (h *AdminCourseHandler) AdminGetCourseBySlugHandler(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return badRequest(c, "Slug is required")
	}

	course, err := h.courseRepo.GetCourseByID(c.Request().Context(), slug)
	if err != nil {
		return internalError(c, "Failed to fetch course")
	}
	if course == nil {
		return notFound(c, "Course not found")
	}

	return c.JSON(http.StatusOK, course)
}

// DELETE /admin/courses/:courseId
func (h *AdminCourseHandler) AdminDeleteCourseHandler(c echo.Context) error {
	courseID := c.Param("courseId")
	if _, err := uuid.Parse(courseID); err != nil {
		return badRequest(c, "Invalid course ID")
	}

	course, err := h.courseRepo.GetCourseByID(c.Request().Context(), courseID)
	if err != nil {
		return internalError(c, "Failed to fetch course")
	}
	if course == nil {
		return notFound(c, "Course not found")
	}

	if err := h.courseRepo.DeleteCourse(c.Request().Context(), courseID); err != nil {
		return internalError(c, "Failed to delete course")
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /admin/courses/:courseId
func (h *AdminCourseHandler) AdminEditCourseHandler(c echo.Context) error {
	courseID := c.Param("courseId")
	if _, err := uuid.Parse(courseID); err != nil {
		return badRequest(c, "Invalid course ID")
	}

	course, err := h.courseRepo.GetCourseByID(c.Request().Context(), courseID)
	if err != nil {
		return internalError(c, "Failed to fetch course")
	}
	if course == nil {
		return notFound(c, "Course not found")
	}

	var req UpdateCourseInfoRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.Type != nil && *req.Type != model.CourseTypePublic && *req.Type != model.CourseTypePrivate {
		return badRequest(c, "Type must be 'public' or 'private'")
	}

	var newStart, newEnd *time.Time

	if req.StartDate != nil {
		t, err := parseDate(*req.StartDate)
		if err != nil {
			return badRequest(c, "start_date must be in format YYYY-MM-DD")
		}
		newStart = &t
	}
	if req.EndDate != nil {
		t, err := parseDate(*req.EndDate)
		if err != nil {
			return badRequest(c, "end_date must be in format YYYY-MM-DD")
		}
		newEnd = &t
	}

	effectiveStart := course.StartDate
	if newStart != nil {
		effectiveStart = newStart
	}
	effectiveEnd := course.EndDate
	if newEnd != nil {
		effectiveEnd = newEnd
	}

	if effectiveStart != nil && effectiveEnd != nil && !effectiveEnd.After(*effectiveStart) {
		return badRequest(c, "end_date must be after start_date")
	}

	if req.Name != nil {
		course.Name = *req.Name
	}
	if req.Description != nil {
		course.Description = req.Description
	}
	if req.RepoTemplate != nil {
		course.RepoTemplate = req.RepoTemplate
	}
	if req.Type != nil {
		course.Type = *req.Type
	}
	if newStart != nil {
		course.StartDate = newStart
	}
	if newEnd != nil {
		course.EndDate = newEnd
	}

	updated, err := h.courseRepo.UpdateCourse(c.Request().Context(), courseID, *course)
	if err != nil {
		return internalError(c, "Failed to update course")
	}

	return c.JSON(http.StatusOK, updated)
}

// PATCH /admin/courses/:courseId/status
func (h *AdminCourseHandler) AdminUpdateCourseStatusHandler(c echo.Context) error {
	courseID := c.Param("courseId")
	if _, err := uuid.Parse(courseID); err != nil {
		return badRequest(c, "Invalid course ID")
	}

	course, err := h.courseRepo.GetCourseByID(c.Request().Context(), courseID)
	if err != nil {
		return internalError(c, "Failed to fetch course")
	}
	if course == nil {
		return notFound(c, "Course not found")
	}

	var req UpdateCourseStatusRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.Status == "" {
		return badRequest(c, "Status is required")
	}
	if !isValidCourseStatus(req.Status) {
		return badRequest(c, "Invalid status value")
	}

	course.Status = req.Status

	updated, err := h.courseRepo.UpdateCourse(c.Request().Context(), courseID, *course)
	if err != nil {
		return internalError(c, "Failed to update course status")
	}

	return c.JSON(http.StatusOK, updated)
}
