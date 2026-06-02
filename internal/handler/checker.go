package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/service"
)

type CheckerHandler struct {
	checkerService ICheckerService
}

func NewCheckerHandler(checkerService ICheckerService) *CheckerHandler {
	return &CheckerHandler{checkerService: checkerService}
}

type submitGradeRequest struct {
	StudentID   uuid.UUID `json:"student_id"`
	TaskID      uuid.UUID `json:"task_id"`
	CourseID    uuid.UUID `json:"course_id"`
	Status      string    `json:"status"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// POST /api/grades
func (h *CheckerHandler) SubmitGrade(c echo.Context) error {
	var req submitGradeRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.StudentID == uuid.Nil {
		return badRequest(c, "student_id is required")
	}
	if req.TaskID == uuid.Nil {
		return badRequest(c, "task_id is required")
	}
	if req.CourseID == uuid.Nil {
		return badRequest(c, "course_id is required")
	}
	if req.SubmittedAt.IsZero() {
		return badRequest(c, "submitted_at is required")
	}
	if req.SubmittedAt.After(time.Now()) {
		return badRequest(c, "submitted_at cannot be in the future")
	}

	submittedAt := req.SubmittedAt
	if submittedAt.IsZero() {
		submittedAt = time.Now()
	}

	score, err := h.checkerService.SubmitGrade(c.Request().Context(), service.SubmitGradeInput{
		StudentID:   req.StudentID,
		TaskID:      req.TaskID,
		CourseID:    req.CourseID,
		Status:      req.Status,
		SubmittedAt: submittedAt,
	})
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, score)
}
