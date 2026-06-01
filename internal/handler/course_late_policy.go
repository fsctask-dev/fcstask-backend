package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type CourseLateHandler struct {
	svc *service.CourseLatePolicy
}

func NewCourseLateHandler(svc *service.CourseLatePolicy) *CourseLateHandler {
	return &CourseLateHandler{svc: svc}
}

type courseLateRequest struct {
	PolicyType        string   `json:"policy_type"`
	SoftPenalty       float64  `json:"soft_penalty"`
	HardDeadlineScore float64  `json:"hard_deadline_score"`
	StepPercent       *float64 `json:"step_percent,omitempty"`
	Coefficient       *float64 `json:"coefficient,omitempty"`
}

// PUT /admin/courses/:courseId/late-policy
func (h *CourseLateHandler) CreateOrUpdate(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}
	var req courseLateRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	policy, err := h.svc.CreateOrUpdate(c.Request().Context(), user.ID, courseID, service.CourseLateInput{
		PolicyType:        model.PolicyType(req.PolicyType),
		SoftPenalty:       req.SoftPenalty,
		HardDeadlineScore: req.HardDeadlineScore,
		StepPercent:       req.StepPercent,
		Coefficient:       req.Coefficient,
	})
	if err != nil {
		return serviceError(c, err)
	}
	return c.JSON(http.StatusOK, policy)
}
