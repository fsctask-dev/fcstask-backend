package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/service"
)

type NamespaceHandler struct {
	namespaceService *service.NamespaceService
}

func NewNamespaceHandler(namespaceService *service.NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{namespaceService: namespaceService}
}

// GET /api/namespaces
func (h *NamespaceHandler) GetNamespaces(c echo.Context) error {
	namespaces, err := h.namespaceService.ListNamespaces(c.Request().Context())
	if err != nil {
		return serviceError(c, err)
	}
	return c.JSON(http.StatusOK, namespaces)
}

// GET /api/namespaces/:id
func (h *NamespaceHandler) GetNamespace(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "Invalid namespace ID")
	}

	ns, err := h.namespaceService.GetNamespace(c.Request().Context(), id)
	if err != nil {
		return serviceError(c, err)
	}
	return c.JSON(http.StatusOK, ns)
}

// GET /api/namespaces/:id/users
func (h *NamespaceHandler) GetNamespaceUsers(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "Invalid namespace ID")
	}

	users, err := h.namespaceService.GetNamespaceUsers(c.Request().Context(), id)
	if err != nil {
		return serviceError(c, err)
	}
	return c.JSON(http.StatusOK, users)
}

// GET /api/namespaces/:id/courses
func (h *NamespaceHandler) GetNamespaceCourses(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "Invalid namespace ID")
	}

	courses, err := h.namespaceService.GetNamespaceCourses(c.Request().Context(), id)
	if err != nil {
		return serviceError(c, err)
	}
	return c.JSON(http.StatusOK, courses)
}
