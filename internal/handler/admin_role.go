package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type AdminRoleHandler struct {
	roleRepo repo.IRoleRepo
	userRepo repo.IUserRepo
}

func NewAdminRoleHandler(roleRepo repo.IRoleRepo, userRepo repo.IUserRepo) *AdminRoleHandler {
	return &AdminRoleHandler{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
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

// POST /admin/courses/:courseId/roles
func (h *AdminRoleHandler) AdminAssignRoleHandler(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	var req AssignRoleRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.UserID == uuid.Nil {
		return badRequest(c, "user_id is required")
	}
	if req.RoleID == uuid.Nil {
		return badRequest(c, "role_id is required")
	}

	if _, err := h.userRepo.GetUserByID(c.Request().Context(), req.UserID); err != nil {
		return notFound(c, "User not found")
	}

	userRole := &model.UserRole{
		UserID:   req.UserID,
		CourseID: courseID,
		RoleID:   req.RoleID,
	}

	if err := h.roleRepo.AssignRole(c.Request().Context(), userRole); err != nil {
		return internalError(c, "Failed to assign role")
	}

	return c.JSON(http.StatusCreated, userRole)
}

// DELETE /admin/courses/:courseId/roles
func (h *AdminRoleHandler) AdminRevokeRoleHandler(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	var req RevokeRoleRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.UserID == uuid.Nil {
		return badRequest(c, "user_id is required")
	}
	if req.RoleID == uuid.Nil {
		return badRequest(c, "role_id is required")
	}

	if err := h.roleRepo.RevokeRole(c.Request().Context(), req.UserID, courseID, req.RoleID); err != nil {
		return internalError(c, "Failed to revoke role")
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /admin/courses/:courseId/roles
func (h *AdminRoleHandler) AdminListUserRolesHandler(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	roles, err := h.roleRepo.GetByCourseID(c.Request().Context(), courseID)
	if err != nil {
		return internalError(c, "Failed to fetch roles")
	}

	return c.JSON(http.StatusOK, roles)
}

// POST /admin/roles/:roleId/permissions
func (h *AdminRoleHandler) AdminAddPermissionHandler(c echo.Context) error {
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return badRequest(c, "Invalid role ID")
	}

	var req AddPermissionRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.Permission == "" {
		return badRequest(c, "Permission is required")
	}

	perm := &model.CourseAdminPermission{
		RoleID:     roleID,
		Permission: req.Permission,
	}

	if err := h.roleRepo.AddPermission(c.Request().Context(), perm); err != nil {
		return internalError(c, "Failed to add permission")
	}

	return c.JSON(http.StatusCreated, perm)
}

// DELETE /admin/roles/:roleId/permissions/:permission
func (h *AdminRoleHandler) AdminRemovePermissionHandler(c echo.Context) error {
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return badRequest(c, "Invalid role ID")
	}

	permission := c.Param("permission")
	if permission == "" {
		return badRequest(c, "Permission is required")
	}

	if err := h.roleRepo.RemovePermission(c.Request().Context(), roleID, permission); err != nil {
		return internalError(c, "Failed to remove permission")
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /admin/roles/:roleId/permissions
func (h *AdminRoleHandler) AdminListPermissionsHandler(c echo.Context) error {
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return badRequest(c, "Invalid role ID")
	}

	perms, err := h.roleRepo.GetPermissions(c.Request().Context(), roleID)
	if err != nil {
		return internalError(c, "Failed to fetch permissions")
	}

	return c.JSON(http.StatusOK, perms)
}
