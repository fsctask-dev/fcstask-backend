package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/service"
)

type AdminRoleHandler struct {
	roleService IAdminRoleService
}

func NewAdminRoleHandler(roleService IAdminRoleService) *AdminRoleHandler {
	return &AdminRoleHandler{roleService: roleService}
}

type AssignCourseAdminRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type RevokeCourseAdminRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type AddPermissionRequest struct {
	Permission string `json:"permission"`
}

type CreateSuperAdminRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// POST /admin/super-admins
func (h *AdminRoleHandler) CreateSuperAdmin(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	var req CreateSuperAdminRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
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
func (h *AdminRoleHandler) AssignCourseAdmin(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}

	var req AssignCourseAdminRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	userRole, err := h.roleService.AssignCourseAdmin(c.Request().Context(), user.ID, service.AssignCourseAdminInput{
		UserID:   req.UserID,
		CourseID: courseID,
	})
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusCreated, userRole)
}

// DELETE /admin/courses/:courseId/roles
func (h *AdminRoleHandler) RevokeCourseAdmin(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}

	var req RevokeCourseAdminRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	if err := h.roleService.RevokeCourseAdmin(c.Request().Context(), user.ID, service.RevokeCourseAdminInput{
		UserID:   req.UserID,
		CourseID: courseID,
	}); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// DELETE /admin/courses/:courseId/participants
func (h *AdminRoleHandler) RemoveCourseParticipant(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}

	var req RevokeCourseAdminRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	if err := h.roleService.RemoveCourseParticipant(c.Request().Context(), user.ID, service.RemoveCourseParticipantInput{
		UserID:   req.UserID,
		CourseID: courseID,
	}); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /admin/courses/:courseId/roles
func (h *AdminRoleHandler) ListUserRoles(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}

	roles, err := h.roleService.ListUserRoles(c.Request().Context(), user.ID, courseID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, roles)
}

// POST /admin/courses/:courseId/roles/:roleId/permissions
func (h *AdminRoleHandler) AddPermission(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}

	roleID, ok := parseUUIDParam(c, "roleId", "Invalid role ID")
	if !ok {
		return nil
	}

	var req AddPermissionRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	perm, err := h.roleService.AddPermission(c.Request().Context(), user.ID, service.AddPermissionInput{
		CourseID:   courseID,
		RoleID:     roleID,
		Permission: req.Permission,
	})
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusCreated, perm)
}

// DELETE /admin/courses/:courseId/roles/:roleId/permissions/:permission
func (h *AdminRoleHandler) RemovePermission(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}

	roleID, ok := parseUUIDParam(c, "roleId", "Invalid role ID")
	if !ok {
		return nil
	}

	permission := c.Param("permission")

	if err := h.roleService.RemovePermission(c.Request().Context(), user.ID, courseID, roleID, permission); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /admin/courses/:courseId/roles/:roleId/permissions
func (h *AdminRoleHandler) ListPermissions(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	courseID, ok := parseUUIDParam(c, "courseId", "Invalid course ID")
	if !ok {
		return nil
	}

	roleID, ok := parseUUIDParam(c, "roleId", "Invalid role ID")
	if !ok {
		return nil
	}

	perms, err := h.roleService.ListPermissions(c.Request().Context(), user.ID, courseID, roleID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, perms)
}
