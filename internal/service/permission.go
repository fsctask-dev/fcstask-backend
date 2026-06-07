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

	PermissionDeadlineRead   = "deadline.read"
	PermissionDeadlineCreate = "deadline.create"
	PermissionDeadlineUpdate = "deadline.update"
	PermissionDeadlineDelete = "deadline.delete"

	PermissionTaskCreate      = "task.create"
	PermissionTaskRead        = "task.read"
	PermissionTaskUpdate      = "task.update"
	PermissionTaskPublish     = "task.publish"
	PermissionTaskDelete      = "task.delete"
	PermissionTaskScoreUpdate = "task.score.update"
	PermissionTaskSubmit      = "task.submit"

	PermissionLeaderboardRead = "leaderboard.read"
	PermissionCourseRead      = "course.read"
	PermissionCourseHiddenRead = "course.hidden.read"

	PermissionCourseRoleAssign = "course.roles.assign"
	PermissionCourseRoleRevoke = "course.roles.revoke"
	PermissionCourseRoleList   = "course.roles.list"

	PermissionCoursePermissionAdd    = "course.permissions.add"
	PermissionCoursePermissionRemove = "course.permissions.remove"
	PermissionCoursePermissionList   = "course.permissions.list"
	PermissionCourseInviteRegenerate = "course.invite.regenerate"

	PermissionCourseCreate     = "course.create"
	PermissionCourseUpdate     = "course.update"
	PermissionSuperAdminCreate = "super_admin.create"
	PermissionIsSuperAdmin     = "is_super_admin"

	PermissionGradeUpdate = "grade.update"

	PermissionLatePolicyCreate = "create.late_policy"
	PermissionStatsRead = "stats.read"
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

	// Super admin is identified by a dedicated global permission.
	globalRoleID, err := roleRepo.GetRoleIDByUserAndCourse(ctx, userID, uuid.Nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return HasPermission(ctx, roleRepo, globalRoleID, PermissionIsSuperAdmin)
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

// CourseStudentPermissions are granted when a user joins a course.
func CourseStudentPermissions() []string {
	return []string{
		PermissionHomeworkRead,
		PermissionDeadlineRead,
		PermissionTaskRead,
		PermissionTaskSubmit,
		PermissionLeaderboardRead,
		PermissionCourseRead,
	}
}

// CourseAdminPermissions are granted when a course owner adds a course admin.
func CourseAdminPermissions() []string {
	return []string{
		PermissionHomeworkCreate,
		PermissionHomeworkUpdate,
		PermissionHomeworkDelete,
		PermissionHomeworkPublish,
		PermissionDeadlineCreate,
		PermissionDeadlineUpdate,
		PermissionDeadlineDelete,
		PermissionTaskCreate,
		PermissionTaskUpdate,
		PermissionTaskPublish,
		PermissionTaskDelete,
		PermissionTaskScoreUpdate,
		PermissionCourseRead,
		PermissionCourseHiddenRead,
		PermissionCourseUpdate,
		PermissionLatePolicyCreate,
		PermissionGradeUpdate,
		PermissionStatsRead,
	}
}

// CourseOwnerPermissions are granted to the creator/owner of a course.
func CourseOwnerPermissions() []string {
	return append(append(CourseStudentPermissions(), CourseAdminPermissions()...),
		PermissionCourseRoleAssign,
		PermissionCourseRoleRevoke,
		PermissionCourseRoleList,
		PermissionCoursePermissionAdd,
		PermissionCoursePermissionRemove,
		PermissionCoursePermissionList,
		PermissionCourseInviteRegenerate,
		PermissionStatsRead,
	)
}

// ServiceSuperAdminPermissions are granted to global service admins.
func ServiceSuperAdminPermissions() []string {
	return []string{
		PermissionIsSuperAdmin,
		PermissionCourseCreate,
		PermissionSuperAdminCreate,
		PermissionStatsRead,
	}
}

func EnsureUserRoleWithPermissions(ctx context.Context, roleRepo repo.IRoleRepo, userID, courseID uuid.UUID, permissions []string) (*model.UserRole, error) {
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
		if err := roleRepo.AssignRoleWithPermissions(ctx, userRole, permissions); err != nil {
			return nil, Internal("Failed to assign role", err)
		}
		return &model.UserRole{
			UserID:   userID,
			CourseID: courseID,
			RoleID:   roleID,
		}, nil
	default:
		return nil, Internal("Failed to load role", err)
	}

	if err := roleRepo.AddPermissions(ctx, roleID, permissions); err != nil {
		return nil, Internal("Failed to add permissions", err)
	}

	return &model.UserRole{
		UserID:   userID,
		CourseID: courseID,
		RoleID:   roleID,
	}, nil
}

func GrantRolePermissions(ctx context.Context, roleRepo repo.IRoleRepo, roleID uuid.UUID, permissions []string) error {
	if roleID == uuid.Nil {
		return BadRequest("role_id is required")
	}
	if len(permissions) == 0 {
		return BadRequest("permissions are required")
	}

	if err := roleRepo.AddPermissions(ctx, roleID, permissions); err != nil {
		return Internal("Failed to add permissions", err)
	}

	return nil
}

func RevokeRolePermissions(ctx context.Context, roleRepo repo.IRoleRepo, roleID uuid.UUID, permissions []string) error {
	if roleID == uuid.Nil {
		return BadRequest("role_id is required")
	}
	if len(permissions) == 0 {
		return BadRequest("permissions are required")
	}

	if err := roleRepo.RemovePermissions(ctx, roleID, permissions); err != nil {
		return Internal("Failed to remove permissions", err)
	}

	return nil
}
