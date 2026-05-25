package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type NamespaceHandler struct {
	nsService *service.NamespaceService
}

func NewNamespaceHandler(nsService *service.NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{nsService: nsService}
}

func (h *NamespaceHandler) GetNamespaces(ctx echo.Context) error {
	user, ok := ctx.Get(UserContextKey).(*models.User)
	if !ok || user == nil {
		return unauthorized(ctx, "User not found in context")
	}

	namespaces, err := h.nsService.GetNamespaces(ctx.Request().Context(), user.ID)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, namespaces)
}

func (h *NamespaceHandler) GetNamespace(ctx echo.Context) error {
	nsID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return badRequest(ctx, "Invalid namespace ID")
	}

	ns, err := h.nsService.GetNamespace(ctx.Request().Context(), nsID)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, ns)
}

func (h *NamespaceHandler) GetNamespaceUsers(ctx echo.Context) error {
	nsID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return badRequest(ctx, "Invalid namespace ID")
	}

	users, err := h.nsService.GetNamespaceUsers(ctx.Request().Context(), nsID)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, users)
}

func (h *NamespaceHandler) GetNamespaceCourses(ctx echo.Context) error {
	nsID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return badRequest(ctx, "Invalid namespace ID")
	}

	courses, err := h.nsService.GetNamespaceCourses(ctx.Request().Context(), nsID)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, courses)
}
