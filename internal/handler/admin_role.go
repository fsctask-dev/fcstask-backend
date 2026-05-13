package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type AdminRoleHandler struct {
	roleService IAdminRoleService
}

func NewAdminRoleHandler(roleService IAdminRoleService) *AdminRoleHandler {
	return &AdminRoleHandler{roleService: roleService}
}

type AssignRoleRequest struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}

type RevokeRoleRequest struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}

type AddPermissionRequest struct {
	Permission string `json:"permission"`
}

type CreateSuperAdminRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// POST /admin/super-admins
func (h *AdminRoleHandler) CreateSuperAdmin(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	var req CreateSuperAdminRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	userRole, err := h.roleService.CreateSuperAdmin(c.Request().Context(), user.ID, service.CreateSuperAdminInput{
		UserID: req.UserID,
	})
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusCreated, userRole)
}

// POST /admin/courses/:courseId/roles
func (h *AdminRoleHandler) AssignRole(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	var req AssignRoleRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	userRole, err := h.roleService.AssignRole(c.Request().Context(), user.ID, service.AssignRoleInput{
		UserID:   req.UserID,
		CourseID: courseID,
		RoleID:   req.RoleID,
	})
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusCreated, userRole)
}

// DELETE /admin/courses/:courseId/roles
func (h *AdminRoleHandler) RevokeRole(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	var req RevokeRoleRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if err := h.roleService.RevokeRole(c.Request().Context(), user.ID, service.RevokeRoleInput{
		UserID:   req.UserID,
		CourseID: courseID,
		RoleID:   req.RoleID,
	}); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /admin/courses/:courseId/roles
func (h *AdminRoleHandler) ListUserRoles(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	roles, err := h.roleService.ListUserRoles(c.Request().Context(), user.ID, courseID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, roles)
}

// POST /admin/roles/:roleId/permissions
func (h *AdminRoleHandler) AddPermission(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return badRequest(c, "Invalid role ID")
	}

	var req AddPermissionRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	perm, err := h.roleService.AddPermission(c.Request().Context(), user.ID, service.AddPermissionInput{
		RoleID:     roleID,
		Permission: req.Permission,
	})
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusCreated, perm)
}

// DELETE /admin/roles/:roleId/permissions/:permission
func (h *AdminRoleHandler) RemovePermission(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return badRequest(c, "Invalid role ID")
	}

	permission := c.Param("permission")

	if err := h.roleService.RemovePermission(c.Request().Context(), user.ID, roleID, permission); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /admin/roles/:roleId/permissions
func (h *AdminRoleHandler) ListPermissions(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return badRequest(c, "Invalid role ID")
	}

	perms, err := h.roleService.ListPermissions(c.Request().Context(), user.ID, roleID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, perms)
}
