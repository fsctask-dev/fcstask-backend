package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type NamespaceHandler struct {
	namespaceService *service.NamespaceService
}

func NewNamespaceHandler(namespaceService *service.NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{namespaceService: namespaceService}
}

func (h *NamespaceHandler) GetNamespaces(ctx echo.Context) error {
	namespaces, err := h.namespaceService.GetNamespaces(ctx.Request().Context())
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, namespaces)
}

func (h *NamespaceHandler) GetNamespace(ctx echo.Context) error {
	namespaceID := ctx.Param("namespaceId")
	if namespaceID == "" {
		return badRequest(ctx, "namespace ID is required")
	}

	namespace, err := h.namespaceService.GetNamespace(ctx.Request().Context(), namespaceID)
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, namespace)
}

func (h *NamespaceHandler) GetInstanceSummary(ctx echo.Context) error {
	summary, err := h.namespaceService.GetInstanceSummary(ctx.Request().Context())
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, summary)
}

func (h *NamespaceHandler) GetCourseScores(ctx echo.Context) error {
	courseID := ctx.Param("courseId")
	if courseID == "" {
		return badRequest(ctx, "course ID is required")
	}

	scores, err := h.namespaceService.GetCourseScores(ctx.Request().Context(), courseID)
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, scores)
}

type ScoreResponse struct {
	ID        int    `json:"id"`
	Student   string `json:"student"`
	Score     int    `json:"score"`
	Submitted string `json:"submitted"`
}

type NamespaceResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Description   string `json:"description,omitempty"`
	GitlabGroupID string `json:"gitlabGroupId"`
	CoursesCount  int    `json:"coursesCount"`
	UsersCount    int    `json:"usersCount"`
}

type NamespaceDetailResponse struct {
	Namespace model.Namespace         `json:"namespace"`
	Users     []model.NamespaceUser   `json:"users"`
	Courses   []model.NamespaceCourse `json:"courses"`
}

type InstanceSummaryResponse struct {
	TotalCourses    int    `json:"totalCourses"`
	TotalUsers      int    `json:"totalUsers"`
	TotalNamespaces int    `json:"totalNamespaces"`
	HealthStatus    string `json:"healthStatus"`
}
