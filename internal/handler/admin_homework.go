package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/service"
)

type AdminHomeworkHandler struct {
	homeworkService IAdminHomeworkService
}

func NewAdminHomeworkHandler(homeworkService IAdminHomeworkService) *AdminHomeworkHandler {
	return &AdminHomeworkHandler{homeworkService: homeworkService}
}

type CreateHomeworkRequest struct {
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

type UpdateHomeworkRequest struct {
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

type PublishHomeworkRequest struct {
	IsPublic bool `json:"is_public"`
}

type SetDeadlineRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	DueDate     string  `json:"due_date"`
}

type UpdateDeadlineRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	DueDate     *string `json:"due_date"`
}

// POST /admin/courses/:courseId/homework
func (h *AdminHomeworkHandler) CreateHomework(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}

	var req CreateHomeworkRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	input := service.CreateHomeworkInput{
		CourseID: courseID,
	}
	if req.StartDate != nil {
		input.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		input.EndDate = *req.EndDate
	}

	hw, err := h.homeworkService.CreateHomework(c.Request().Context(), user.ID, input)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusCreated, hw)
}

// GET /admin/courses/:courseId/homework/:hwId
func (h *AdminHomeworkHandler) GetHomework(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	hwID, ok := parseUUIDParam(c, "hwId", "Invalid homework ID")
	if !ok {
		return nil
	}
	hw, err := h.homeworkService.GetHomework(c.Request().Context(), user.ID, hwID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, hw)
}

// GET /admin/courses/:courseId/homework
func (h *AdminHomeworkHandler) ListHomework(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}
	hws, err := h.homeworkService.ListHomework(c.Request().Context(), user.ID, courseID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, hws)
}

// PATCH /admin/courses/:courseId/homework/:hwId
func (h *AdminHomeworkHandler) UpdateHomework(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	hwID, ok := parseUUIDParam(c, "hwId", "Invalid homework ID")
	if !ok {
		return nil
	}

	var req UpdateHomeworkRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	input := service.UpdateHomeworkInput{}
	if req.StartDate != nil {
		input.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		input.EndDate = *req.EndDate
	}

	hw, err := h.homeworkService.UpdateHomework(c.Request().Context(), user.ID, hwID, input)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, hw)
}

// DELETE /admin/courses/:courseId/homework/:hwId
func (h *AdminHomeworkHandler) DeleteHomework(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	hwID, ok := parseUUIDParam(c, "hwId", "Invalid homework ID")
	if !ok {
		return nil
	}
	if err := h.homeworkService.DeleteHomework(c.Request().Context(), user.ID, hwID); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /admin/courses/:courseId/homework/:hwId/publish
func (h *AdminHomeworkHandler) PublishHomework(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	hwID, ok := parseUUIDParam(c, "hwId", "Invalid homework ID")
	if !ok {
		return nil
	}

	var req PublishHomeworkRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	hw, err := h.homeworkService.PublishHomework(c.Request().Context(), user.ID, hwID, req.IsPublic)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, hw)
}

// PUT /admin/courses/:courseId/homework/:hwId/deadline
func (h *AdminHomeworkHandler) SetDeadline(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}

	hwID, ok := parseUUIDParam(c, "hwId", "Invalid homework ID")
	if !ok {
		return nil
	}

	var req SetDeadlineRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	assignedBy := &user.ID

	input := service.SetDeadlineInput{
		CourseID:   courseID,
		HomeworkID: hwID,
		Title:      req.Title,
		DueDate:    req.DueDate,
		AssignedBy: assignedBy,
	}
	if req.Description != nil {
		input.Description = *req.Description
	}

	deadline, err := h.homeworkService.SetDeadline(c.Request().Context(), user.ID, input)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusCreated, deadline)
}

// PATCH /admin/deadlines/:deadlineId
func (h *AdminHomeworkHandler) UpdateDeadline(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	deadlineID, ok := parseUUIDParam(c, "deadlineId", "Invalid deadline ID")
	if !ok {
		return nil
	}

	var req UpdateDeadlineRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	input := service.UpdateDeadlineInput{}
	if req.Title != nil {
		input.Title = *req.Title
	}
	if req.Description != nil {
		input.Description = *req.Description
	}
	if req.DueDate != nil {
		input.DueDate = *req.DueDate
	}

	deadline, err := h.homeworkService.UpdateDeadline(c.Request().Context(), user.ID, deadlineID, input)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, deadline)
}

// DELETE /admin/deadlines/:deadlineId
func (h *AdminHomeworkHandler) DeleteDeadline(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	deadlineID, ok := parseUUIDParam(c, "deadlineId", "Invalid deadline ID")
	if !ok {
		return nil
	}
	if err := h.homeworkService.DeleteDeadline(c.Request().Context(), user.ID, deadlineID); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
