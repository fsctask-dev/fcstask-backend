package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type AdminHomeworkHandler struct {
	homeworkRepo repo.IHomeworkRepo
	deadlineRepo repo.IDeadlineRepo
}

func NewAdminHomeworkHandler(homeworkRepo repo.IHomeworkRepo, deadlineRepo repo.IDeadlineRepo) *AdminHomeworkHandler {
	return &AdminHomeworkHandler{
		homeworkRepo: homeworkRepo,
		deadlineRepo: deadlineRepo,
	}
}

// По просмотру посылок. У нас вообще есть где-то модель и миграции для submitions? не нашел их

type CreateHomeworkRequest struct {
	CourseID  uuid.UUID `json:"course_id"`
	StartDate *string   `json:"start_date"`
	EndDate   *string   `json:"end_date"`
}

type UpdateHomeworkRequest struct {
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

type PublishHomeworkRequest struct {
	IsPublic bool `json:"is_public"`
}

type SetHomeworkDeadlineRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	DueDate     string  `json:"due_date"` // RFC3339
}

// POST /admin/courses/:courseId/homework
func (h *AdminHomeworkHandler) AdminCreateHomeworkHandler(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	var req CreateHomeworkRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	hw := &model.Homework{
		CourseID: courseID,
	}

	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return badRequest(c, "start_date must be in format YYYY-MM-DD")
		}
		hw.StartDate = &t
	}
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return badRequest(c, "end_date must be in format YYYY-MM-DD")
		}
		hw.EndDate = &t
	}
	if hw.StartDate != nil && hw.EndDate != nil && !hw.EndDate.After(*hw.StartDate) {
		return badRequest(c, "end_date must be after start_date")
	}

	if err := h.homeworkRepo.Create(c.Request().Context(), hw); err != nil {
		return internalError(c, "Failed to create homework")
	}

	return c.JSON(http.StatusCreated, hw)
}

// GET /admin/courses/:courseId/homework/:hwId
func (h *AdminHomeworkHandler) AdminGetHomeworkHandler(c echo.Context) error {
	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	hw, err := h.homeworkRepo.GetByID(c.Request().Context(), hwID)
	if err != nil {
		return notFound(c, "Homework not found")
	}

	return c.JSON(http.StatusOK, hw)
}

// GET /admin/courses/:courseId/homework
func (h *AdminHomeworkHandler) AdminListHomeworkHandler(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	hws, err := h.homeworkRepo.GetByCourseID(c.Request().Context(), courseID)
	if err != nil {
		return internalError(c, "Failed to fetch homework list")
	}

	return c.JSON(http.StatusOK, hws)
}

// PATCH /admin/courses/:courseId/homework/:hwId
func (h *AdminHomeworkHandler) AdminUpdateHomeworkHandler(c echo.Context) error {
	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	hw, err := h.homeworkRepo.GetByID(c.Request().Context(), hwID)
	if err != nil {
		return notFound(c, "Homework not found")
	}

	var req UpdateHomeworkRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return badRequest(c, "start_date must be in format YYYY-MM-DD")
		}
		hw.StartDate = &t
	}
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return badRequest(c, "end_date must be in format YYYY-MM-DD")
		}
		hw.EndDate = &t
	}
	if hw.StartDate != nil && hw.EndDate != nil && !hw.EndDate.After(*hw.StartDate) {
		return badRequest(c, "end_date must be after start_date")
	}

	if err := h.homeworkRepo.Update(c.Request().Context(), hw); err != nil {
		return internalError(c, "Failed to update homework")
	}

	return c.JSON(http.StatusOK, hw)
}

// DELETE /admin/courses/:courseId/homework/:hwId
func (h *AdminHomeworkHandler) AdminDeleteHomeworkHandler(c echo.Context) error {
	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	if _, err := h.homeworkRepo.GetByID(c.Request().Context(), hwID); err != nil {
		return notFound(c, "Homework not found")
	}

	if err := h.homeworkRepo.Delete(c.Request().Context(), hwID); err != nil {
		return internalError(c, "Failed to delete homework")
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /admin/courses/:courseId/homework/:hwId/publish
func (h *AdminHomeworkHandler) AdminPublishHomeworkHandler(c echo.Context) error {
	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	hw, err := h.homeworkRepo.GetByID(c.Request().Context(), hwID)
	if err != nil {
		return notFound(c, "Homework not found")
	}

	var req PublishHomeworkRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	hw.IsPublic = &req.IsPublic

	if err := h.homeworkRepo.Update(c.Request().Context(), hw); err != nil {
		return internalError(c, "Failed to publish homework")
	}

	return c.JSON(http.StatusOK, hw)
}

// PUT /admin/courses/:courseId/homework/:hwId/deadline
func (h *AdminHomeworkHandler) AdminSetHomeworkDeadlineHandler(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	if _, err := h.homeworkRepo.GetByID(c.Request().Context(), hwID); err != nil {
		return notFound(c, "Homework not found")
	}

	var req SetHomeworkDeadlineRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.Title == "" {
		return badRequest(c, "Title is required")
	}

	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		return badRequest(c, "due_date must be in RFC3339 format")
	}

	user, _ := c.Get(UserContextKey).(*model.User)
	var assignedBy *uuid.UUID
	if user != nil {
		assignedBy = &user.ID
	}

	deadline := &model.Deadline{
		Title:       req.Title,
		Description: req.Description,
		CourseID:    courseID,
		DueDate:     dueDate,
		AssignedBy:  assignedBy,
	}

	if err := h.deadlineRepo.Create(c.Request().Context(), deadline); err != nil {
		return internalError(c, "Failed to set deadline")
	}

	return c.JSON(http.StatusCreated, deadline)
}

// PATCH /admin/deadlines/:deadlineId
func (h *AdminHomeworkHandler) AdminUpdateDeadlineHandler(c echo.Context) error {
	deadlineID, err := uuid.Parse(c.Param("deadlineId"))
	if err != nil {
		return badRequest(c, "Invalid deadline ID")
	}

	deadline, err := h.deadlineRepo.GetByID(c.Request().Context(), deadlineID)
	if err != nil {
		return notFound(c, "Deadline not found")
	}

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		DueDate     *string `json:"due_date"`
	}

	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.Title != nil {
		if *req.Title == "" {
			return badRequest(c, "Title cannot be empty")
		}
		deadline.Title = *req.Title
	}
	if req.Description != nil {
		deadline.Description = req.Description
	}
	if req.DueDate != nil {
		dueDate, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			return badRequest(c, "due_date must be in RFC3339 format")
		}
		deadline.DueDate = dueDate
	}

	if err := h.deadlineRepo.Update(c.Request().Context(), deadline); err != nil {
		return internalError(c, "Failed to update deadline")
	}

	return c.JSON(http.StatusOK, deadline)
}

// DELETE /admin/deadlines/:deadlineId
func (h *AdminHomeworkHandler) AdminDeleteDeadlineHandler(c echo.Context) error {
	deadlineID, err := uuid.Parse(c.Param("deadlineId"))
	if err != nil {
		return badRequest(c, "Invalid deadline ID")
	}

	if _, err := h.deadlineRepo.GetByID(c.Request().Context(), deadlineID); err != nil {
		return notFound(c, "Deadline not found")
	}

	if err := h.deadlineRepo.Delete(c.Request().Context(), deadlineID); err != nil {
		return internalError(c, "Failed to delete deadline")
	}

	return c.NoContent(http.StatusNoContent)
}
