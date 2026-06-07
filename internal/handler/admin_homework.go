package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type AdminHomeworkHandler struct {
	homeworkService IAdminHomeworkService
}

func NewAdminHomeworkHandler(homeworkService IAdminHomeworkService) *AdminHomeworkHandler {
	return &AdminHomeworkHandler{homeworkService: homeworkService}
}

type CreateHomeworkRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Position    *int    `json:"position"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

type UpdateHomeworkRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Position    *int    `json:"position"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

type PublishHomeworkRequest struct {
	IsPublic bool `json:"is_public"`
}

type SetDeadlineRequest struct {
	CourseID     string    `json:"course_id"`
	Title        string    `json:"title"`
	Description  *string   `json:"description"`
	SoftDeadline time.Time `json:"soft_deadline"`
	HardDeadline time.Time `json:"hard_deadline"`
}

type UpdateDeadlineRequest struct {
	Title        *string    `json:"title"`
	Description  *string    `json:"description"`
	SoftDeadline *time.Time `json:"soft_deadline"`
	HardDeadline *time.Time `json:"hard_deadline"`
}

// POST /admin/courses/:courseId/homework
func (h *AdminHomeworkHandler) CreateHomework(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	var req CreateHomeworkRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	input := service.CreateHomeworkInput{
		CourseID: courseID,
	}
	if req.Title != nil {
		input.Title = *req.Title
	}
	if req.Description != nil {
		input.Description = *req.Description
	}
	if req.Position != nil {
		input.Position = *req.Position
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
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}
	hw, err := h.homeworkService.GetHomework(c.Request().Context(), user.ID, hwID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, hw)
}

// GET /admin/courses/:courseId/homework
func (h *AdminHomeworkHandler) ListHomework(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}
	hws, err := h.homeworkService.ListHomework(c.Request().Context(), user.ID, courseID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, hws)
}

// PATCH /admin/courses/:courseId/homework/:hwId
func (h *AdminHomeworkHandler) UpdateHomework(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	var req UpdateHomeworkRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	input := service.UpdateHomeworkInput{}
	input.Title = req.Title
	input.Description = req.Description
	input.Position = req.Position
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
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}
	if err := h.homeworkService.DeleteHomework(c.Request().Context(), user.ID, hwID); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /admin/courses/:courseId/homework/:hwId/publish
func (h *AdminHomeworkHandler) PublishHomework(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	var req PublishHomeworkRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	hw, err := h.homeworkService.PublishHomework(c.Request().Context(), user.ID, hwID, req.IsPublic)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, hw)
}

// GET /api/homework/:hwId/deadline
func (h *AdminHomeworkHandler) GetDeadlineByHomeworkID(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	deadline, err := h.homeworkService.GetDeadlineByHomeworkID(c.Request().Context(), user.ID, hwID)
	if err != nil {
		return serviceError(c, err)
	}
	return c.JSON(http.StatusOK, deadline)
}

// PUT /admin/courses/:courseId/homework/:hwId/deadline
// PUT /api/homework/:hwId/deadline
func (h *AdminHomeworkHandler) SetDeadline(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	var courseID uuid.UUID
	if c.Param("courseId") != "" {
		var err error
		courseID, err = uuid.Parse(c.Param("courseId"))
		if err != nil {
			return badRequest(c, "Invalid course ID")
		}
	}

	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	var req SetDeadlineRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if courseID == uuid.Nil {
		courseID, err = uuid.Parse(req.CourseID)
		if err != nil {
			return badRequest(c, "course_id is required")
		}
	}

	var assignedBy *uuid.UUID
	assignedBy = &user.ID

	input := service.SetDeadlineInput{
		CourseID:     courseID,
		HomeworkID:   hwID,
		Title:        req.Title,
		AssignedBy:   assignedBy,
		SoftDeadline: req.SoftDeadline,
		HardDeadline: req.HardDeadline,
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
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	deadlineID, err := uuid.Parse(c.Param("deadlineId"))
	if err != nil {
		return badRequest(c, "Invalid deadline ID")
	}

	var req UpdateDeadlineRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	input := service.UpdateDeadlineInput{}
	if req.Title != nil {
		input.Title = *req.Title
	}
	if req.Description != nil {
		input.Description = *req.Description
	}
	if req.SoftDeadline != nil && !req.SoftDeadline.IsZero() {
		input.SoftDeadline = *req.SoftDeadline
	}
	if req.HardDeadline != nil && !req.HardDeadline.IsZero() {
		input.HardDeadline = *req.HardDeadline
	}

	deadline, err := h.homeworkService.UpdateDeadline(c.Request().Context(), user.ID, deadlineID, input)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, deadline)
}

// DELETE /admin/deadlines/:deadlineId
func (h *AdminHomeworkHandler) DeleteDeadline(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	deadlineID, err := uuid.Parse(c.Param("deadlineId"))
	if err != nil {
		return badRequest(c, "Invalid deadline ID")
	}
	if err := h.homeworkService.DeleteDeadline(c.Request().Context(), user.ID, deadlineID); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
