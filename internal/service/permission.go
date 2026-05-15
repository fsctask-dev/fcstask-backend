package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

const (
	PermissionHomeworkCreate  = "homework.create"
	PermissionHomeworkRead    = "homework.read"
	PermissionHomeworkUpdate  = "homework.update"
	PermissionHomeworkDelete  = "homework.delete"
	PermissionHomeworkPublish = "homework.publish"

	PermissionDeadlineCreate = "deadline.create"
	PermissionDeadlineUpdate = "deadline.update"
	PermissionDeadlineDelete = "deadline.delete"

	PermissionTaskCreate      = "task.create"
	PermissionTaskRead        = "task.read"
	PermissionTaskUpdate      = "task.update"
	PermissionTaskDelete      = "task.delete"
	PermissionTaskScoreUpdate = "task.score.update"

	PermissionAdminRoleAssign  = "admin.roles.assign"
	PermissionAdminRoleRevoke  = "admin.roles.revoke"
	PermissionAdminRoleList    = "admin.roles.list"
	PermissionAdminPermAdd     = "admin.permissions.add"
	PermissionAdminPermRemove  = "admin.permissions.remove"
	PermissionAdminPermList    = "admin.permissions.list"
	PermissionAdminSuperCreate = "admin.super_admins.create"
)

func HasPermission(ctx context.Context, roleRepo repo.IRoleRepo, roleID uuid.UUID, permission string) (bool, error) {
	if roleID == uuid.Nil || permission == "" {
		return false, nil
	}

	return roleRepo.HasPermission(ctx, roleID, permission)
}

func HasScopedPermission(ctx context.Context, roleRepo repo.IRoleRepo, userID, courseID uuid.UUID, permission string) (bool, error) {
	if userID == uuid.Nil || permission == "" {
		return false, nil
	}

	roleID, err := roleRepo.GetRoleIDByUserAndCourse(ctx, userID, courseID)
	if err == nil {
		allowed, err := HasPermission(ctx, roleRepo, roleID, permission)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	if courseID == uuid.Nil {
		return false, nil
	}

	// Super admin is identified by the existence of a global role scoped to uuid.Nil.
	_, err = roleRepo.GetRoleIDByUserAndCourse(ctx, userID, uuid.Nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func RequireScopedPermission(ctx context.Context, roleRepo repo.IRoleRepo, userID, courseID uuid.UUID, permission string) error {
	allowed, err := HasScopedPermission(ctx, roleRepo, userID, courseID, permission)
	if err != nil {
		return Internal("Failed to check permissions", err)
	}
	if !allowed {
		return Forbidden("You don't have permission to access this resource")
	}
	return nil
}

func CoursePermissions() []string {
	return []string{
		PermissionHomeworkCreate,
		PermissionHomeworkRead,
		PermissionHomeworkUpdate,
		PermissionHomeworkDelete,
		PermissionHomeworkPublish,
		PermissionDeadlineCreate,
		PermissionDeadlineUpdate,
		PermissionDeadlineDelete,
		PermissionTaskCreate,
		PermissionTaskRead,
		PermissionTaskUpdate,
		PermissionTaskDelete,
		PermissionTaskScoreUpdate,
	}
}

func AdminPermissions() []string {
	return []string{
		PermissionAdminRoleAssign,
		PermissionAdminRoleRevoke,
		PermissionAdminRoleList,
		PermissionAdminPermAdd,
		PermissionAdminPermRemove,
		PermissionAdminPermList,
		PermissionAdminSuperCreate,
	}
}

func EnsureRolePermissions(ctx context.Context, roleRepo repo.IRoleRepo, userID, courseID uuid.UUID, permissions []string) (*model.UserRole, error) {
	if userID == uuid.Nil {
		return nil, BadRequest("user_id is required")
	}
	if len(permissions) == 0 {
		return nil, BadRequest("permissions are required")
	}

	roleID, err := roleRepo.GetRoleIDByUserAndCourse(ctx, userID, courseID)
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		roleID = uuid.New()
		userRole := &model.UserRole{
			UserID:   userID,
			CourseID: courseID,
			RoleID:   roleID,
		}
		if err := roleRepo.AssignRole(ctx, userRole); err != nil {
			return nil, Internal("Failed to assign role", err)
		}
	default:
		return nil, Internal("Failed to load role", err)
	}

	for _, permission := range permissions {
		if err := roleRepo.AddPermission(ctx, &model.CourseAdminPermission{
			RoleID:     roleID,
			Permission: permission,
		}); err != nil {
			return nil, Internal("Failed to add permission", err)
		}
	}

	return &model.UserRole{
		UserID:   userID,
		CourseID: courseID,
		RoleID:   roleID,
	}, nil
}
