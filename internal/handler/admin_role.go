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

type AssignCourseAdminRequest struct {
	UserID uuid.UUID `json:"user_id"`
    Role     string    `json:"role,omitempty"`   // дефолтный админ или owner
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

type GrantCourseCreateRequest struct {
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

// POST /admin/users/:userId/grant-course-create
func (h *AdminRoleHandler) GrantCourseCreatePermission(c echo.Context) error {
    user, ok := c.Get(UserContextKey).(*model.User)
    if !ok || user == nil {
        return unauthorized(c, "User not found")
    }

    targetUserID, err := uuid.Parse(c.Param("userId"))
    if err != nil {
        return badRequest(c, "Invalid user ID")
    }

    userRole, err := h.roleService.GrantCourseCreatePermission(c.Request().Context(), user.ID, targetUserID)
    if err != nil {
        return serviceError(c, err)
    }

    return c.JSON(http.StatusCreated, userRole)
}

// DELETE /admin/users/:userId/grant-course-create
func (h *AdminRoleHandler) RevokeCourseCreatePermission(c echo.Context) error {
    user, ok := c.Get(UserContextKey).(*model.User)
    if !ok || user == nil {
        return unauthorized(c, "User not found")
    }

    targetUserID, err := uuid.Parse(c.Param("userId"))
    if err != nil {
        return badRequest(c, "Invalid user ID")
    }

    if err := h.roleService.RevokeCourseCreatePermission(c.Request().Context(), user.ID, targetUserID); err != nil {
        return serviceError(c, err)
    }

    return c.NoContent(http.StatusNoContent)
}

// POST /admin/courses/:courseId/roles
func (h *AdminRoleHandler) AssignCourseAdmin(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	var req AssignCourseAdminRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	userRole, err := h.roleService.AssignCourseAdmin(c.Request().Context(), user.ID, service.AssignCourseAdminInput{
		UserID:   req.UserID,
		CourseID: courseID,
		Role:     req.Role,
	})
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusCreated, userRole)
}

// DELETE /admin/courses/:courseId/roles
func (h *AdminRoleHandler) RevokeCourseAdmin(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	var req RevokeCourseAdminRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
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
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	var req RevokeCourseAdminRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
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

// POST /admin/courses/:courseId/roles/:roleId/permissions
func (h *AdminRoleHandler) AddPermission(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
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
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return badRequest(c, "Invalid role ID")
	}

	permission := c.Param("permission")

	if err := h.roleService.RemovePermission(c.Request().Context(), user.ID, courseID, roleID, permission); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /admin/courses/:courseId/roles/:roleId/permissions
func (h *AdminRoleHandler) ListPermissions(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return badRequest(c, "Invalid course ID")
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return badRequest(c, "Invalid role ID")
	}

	perms, err := h.roleService.ListPermissions(c.Request().Context(), user.ID, courseID, roleID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, perms)
}
