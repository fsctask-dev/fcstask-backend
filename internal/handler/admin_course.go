package handler

import (
	"net/http"

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

// PATCH /admin/courses/:courseId
func (h *AdminCourseHandler) AdminEditCourseHandler(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	course, err := h.courseRepo.GetByID(c.Request().Context(), courseID.String())
	if err != nil {
		return notFound(c, "Course not found")
	}

	var req UpdateCourseInfoRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.Type != nil && *req.Type != model.CourseTypePublic && *req.Type != model.CourseTypePrivate {
		return badRequest(c, "Type must be 'public' or 'private'")
	}

	if req.StartDate != nil && !isValidDate(*req.StartDate) {
		return badRequest(c, "start_date must be in format YYYY-MM-DD")
	}
	if req.EndDate != nil && !isValidDate(*req.EndDate) {
		return badRequest(c, "end_date must be in format YYYY-MM-DD")
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
	if req.StartDate != nil {
		course.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		course.EndDate = req.EndDate
	}

	if course.StartDate != nil && course.EndDate != nil {
		if !isValidDateRange(*course.StartDate, *course.EndDate) {
			return badRequest(c, "end_date must be after start_date")
		}
	}

	if err := h.courseRepo.Update(c.Request().Context(), course); err != nil {
		return internalError(c, "Failed to update course")
	}

	return c.JSON(http.StatusOK, course)
}

// PATCH /admin/courses/:courseId/status
func (h *AdminCourseHandler) AdminUpdateCourseStatusHandler(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	course, err := h.courseRepo.GetByID(c.Request().Context(), courseID.String())
	if err != nil {
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

	if err := h.courseRepo.Update(c.Request().Context(), course); err != nil {
		return internalError(c, "Failed to update course status")
	}

	return c.JSON(http.StatusOK, course)
}

// POST /admin/courses
func (h *AdminCourseHandler) AdminCreateCourseHandler(c echo.Context) error {
	var req struct {
		Name         string           `json:"name"`
		Description  *string          `json:"description"`
		RepoTemplate *string          `json:"repo_template"`
		Type         model.CourseType `json:"type"`
		StartDate    *string          `json:"start_date"`
		EndDate      *string          `json:"end_date"`
		Status       string           `json:"status"`
	}

	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.Name == "" {
		return badRequest(c, "Name is required")
	}

	if !isValidCourseStatus(req.Status) {
		return badRequest(c, "Invalid status")
	}

	course := &model.Course{
		Name:         req.Name,
		Description:  req.Description,
		RepoTemplate: req.RepoTemplate,
		Type:         req.Type,
		Status:       req.Status,
	}

	if req.StartDate != nil && isValidDate(*req.StartDate) {
		course.StartDate = req.StartDate
	}
	if req.EndDate != nil && isValidDate(*req.EndDate) {
		course.EndDate = req.EndDate
	}

	if err := h.courseRepo.Create(c.Request().Context(), course); err != nil {
		return internalError(c, "Failed to create course")
	}

	return c.JSON(http.StatusCreated, course)
}

// DELETE /admin/courses/:courseId
func (h *AdminCourseHandler) AdminDeleteCourseHandler(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	if _, err := h.courseRepo.GetByID(c.Request().Context(), courseID.String()); err != nil {
		return notFound(c, "Course not found")
	}

	if err := h.courseRepo.Delete(c.Request().Context(), courseID.String()); err != nil {
		return internalError(c, "Failed to delete course")
	}

	return c.NoContent(http.StatusNoContent)
}
