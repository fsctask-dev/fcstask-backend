package handler

import (
	"fcstask-backend/internal/db/model"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/service"
)

type GradeUpdateHandler struct {
	gradeUpdateService IGradeUpdateService
}

func NewGradeUpdateHandler(svc IGradeUpdateService) *GradeUpdateHandler {
	return &GradeUpdateHandler{gradeUpdateService: svc}
}

type GradeUpdateRequest struct {
	StudentID uuid.UUID `json:"studentId"`
	Score     *int      `json:"score"`
}

// POST /admin/courses/:courseId/homework/:hwId/tasks/:taskId/update_grade
func (h *GradeUpdateHandler) UpdateGrade(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("course_id"))
	if err != nil {
		return badRequest(c, "invalid course_id")
	}
	taskID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		return badRequest(c, "invalid task_id")
	}
	var req GradeUpdateRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	if req.StudentID == uuid.Nil {
		return badRequest(c, "invalid student id")
	}
	if req.Score == nil {
		return badRequest(c, "invalid score")
	}

	score, err := h.gradeUpdateService.UpdateGrade(c.Request().Context(), user.ID, service.UpdateGradeInput{
		StudentID: req.StudentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     req.Score,
	})
	if err != nil {
		return serviceError(c, err)
	}
	return c.JSON(http.StatusOK, score)
}
