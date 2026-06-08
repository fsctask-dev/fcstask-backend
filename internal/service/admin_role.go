package service

import (
	"context"

	"github.com/google/uuid"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/metrics"
)

type AdminRoleService struct {
	roleRepo repo.IRoleRepo
	userRepo repo.IUserRepo

	adminMetrics *metrics.AdminMetrics
}

func NewAdminRoleService(roleRepo repo.IRoleRepo, userRepo repo.IUserRepo) *AdminRoleService {
	return &AdminRoleService{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

func (s *AdminRoleService) WithMetrics(m *metrics.AdminMetrics) *AdminRoleService {
	s.adminMetrics = m
	return s
}

type AssignCourseAdminInput struct {
	UserID   uuid.UUID
	CourseID uuid.UUID
	Role     string
}

type RevokeCourseAdminInput struct {
	UserID   uuid.UUID
	CourseID uuid.UUID
}

type RemoveCourseParticipantInput struct {
	UserID   uuid.UUID
	CourseID uuid.UUID
}

type AddPermissionInput struct {
	CourseID   uuid.UUID
	RoleID     uuid.UUID
	Permission string
}

type CreateSuperAdminInput struct {
	UserID uuid.UUID
}

func (s *AdminRoleService) AssignCourseAdmin(ctx context.Context, userID uuid.UUID, input AssignCourseAdminInput) (role *model.UserRole, err error) {
    defer func() { s.adminMetrics.IncAction(metrics.AdminActionAssignRole, adminOutcome(err)) }()

    if input.UserID == uuid.Nil || input.CourseID == uuid.Nil {
        return nil, BadRequest("user_id or course_id is required")
    }
    if err = RequireScopedPermission(ctx, s.roleRepo, userID, input.CourseID, PermissionCourseRoleAssign); err != nil {
        return nil, err
    }

    if _, err = s.userRepo.GetUserByID(ctx, input.UserID); err != nil {
        return nil, NotFound("User not found")
    }

    switch input.Role {
    case "owner":
        // создаст роль если нет, выдаёт полный набор прав
        role, err = EnsureUserRoleWithPermissions(ctx, s.roleRepo, input.UserID, input.CourseID, CourseOwnerPermissions())
        if err != nil {
            return nil, Internal("Failed to assign owner permissions", err)
        }
        return role, nil

    default:
        // "admin" или "" требует участника, добавляет админ-права
        roleID, err := s.roleRepo.GetRoleIDByUserAndCourse(ctx, input.UserID, input.CourseID)
        if err != nil {
            return nil, NotFound("User is not a course participant")
        }
        if err = GrantRolePermissions(ctx, s.roleRepo, roleID, CourseAdminPermissions()); err != nil {
            return nil, err
        }
        return &model.UserRole{
            UserID:   input.UserID,
            CourseID: input.CourseID,
            RoleID:   roleID,
        }, nil
    }
}

func (s *AdminRoleService) CreateSuperAdmin(ctx context.Context, userID uuid.UUID, input CreateSuperAdminInput) (role *model.UserRole, err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionPromoteSuperAdmin, adminOutcome(err)) }()

	if err = RequireScopedPermission(ctx, s.roleRepo, userID, uuid.Nil, PermissionSuperAdminCreate); err != nil {
		return nil, err
	}
	if input.UserID == uuid.Nil {
		return nil, BadRequest("user_id is required")
	}

	if _, err = s.userRepo.GetUserByID(ctx, input.UserID); err != nil {
		return nil, NotFound("User not found")
	}

	role, err = EnsureUserRoleWithPermissions(ctx, s.roleRepo, input.UserID, uuid.Nil, ServiceSuperAdminPermissions())
	return role, err
}

func (s *AdminRoleService) GrantCourseCreatePermission(ctx context.Context, userID uuid.UUID, targetUserID uuid.UUID) (*model.UserRole, error) {
    if targetUserID == uuid.Nil {
        return nil, BadRequest("user_id is required")
    }

    if err := RequireScopedPermission(ctx, s.roleRepo, userID, uuid.Nil, PermissionCourseCreateGrant); err != nil {
        return nil, err
    }

    if _, err := s.userRepo.GetUserByID(ctx, targetUserID); err != nil {
        return nil, NotFound("User not found")
    }

    // если у пользователя нет глобальной роли, создаст ее и запишет право
    // если есть, добавит course.create к существующей роли
    role, err := EnsureUserRoleWithPermissions(ctx, s.roleRepo, targetUserID, uuid.Nil, []string{PermissionCourseCreate})
    if err != nil {
        return nil, Internal("Failed to grant course create permission", err)
    }

    return role, nil
}

func (s *AdminRoleService) RevokeCourseCreatePermission(ctx context.Context, userID uuid.UUID, targetUserID uuid.UUID) error {
    if targetUserID == uuid.Nil {
        return BadRequest("user_id is required")
    }

    if err := RequireScopedPermission(ctx, s.roleRepo, userID, uuid.Nil, PermissionCourseCreateGrant); err != nil {
        return err
    }

    roleID, err := s.roleRepo.GetRoleIDByUserAndCourse(ctx, targetUserID, uuid.Nil)
    if err != nil {
        return NotFound("User does not have a global role or course.create permission")
    }

	// удалит только право course.create
    if err := RevokeRolePermissions(ctx, s.roleRepo, roleID, []string{PermissionCourseCreate}); err != nil {
        return Internal("Failed to revoke course create permission", err)
    }

    return nil
}

func (s *AdminRoleService) RevokeCourseAdmin(ctx context.Context, userID uuid.UUID, input RevokeCourseAdminInput) (err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionRevokeRole, adminOutcome(err)) }()

	if input.UserID == uuid.Nil || input.CourseID == uuid.Nil {
		return BadRequest("user_id or course_id is required")
	}
	if err = RequireScopedPermission(ctx, s.roleRepo, userID, input.CourseID, PermissionCourseRoleRevoke); err != nil {
		return err
	}

	roleID, err := s.roleRepo.GetRoleIDByUserAndCourse(ctx, input.UserID, input.CourseID)
	if err != nil {
		return NotFound("User is not a course participant")
	}
	if err = RevokeRolePermissions(ctx, s.roleRepo, roleID, CourseAdminPermissions()); err != nil {
		return err
	}

	return nil
}

func (s *AdminRoleService) RemoveCourseParticipant(ctx context.Context, userID uuid.UUID, input RemoveCourseParticipantInput) (err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionRemoveParticipant, adminOutcome(err)) }()

	if input.UserID == uuid.Nil || input.CourseID == uuid.Nil {
		return BadRequest("user_id or course_id is required")
	}
	if err = RequireScopedPermission(ctx, s.roleRepo, userID, input.CourseID, PermissionCourseRoleRevoke); err != nil {
		return err
	}

	roleID, err := s.roleRepo.GetRoleIDByUserAndCourse(ctx, input.UserID, input.CourseID)
	if err != nil {
		return NotFound("User is not a course participant")
	}

	if err = s.roleRepo.RevokeRoleWithPermissions(ctx, input.UserID, input.CourseID, roleID); err != nil {
		return Internal("Failed to remove course participant", err)
	}

	return nil
}

func (s *AdminRoleService) ListUserRoles(ctx context.Context, userID, courseID uuid.UUID) ([]model.UserRole, error) {
	if courseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	if err := RequireScopedPermission(ctx, s.roleRepo, userID, courseID, PermissionCourseRoleList); err != nil {
		return nil, err
	}

	roles, err := s.roleRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, Internal("Failed to fetch roles", err)
	}

	return roles, nil
}

func (s *AdminRoleService) AddPermission(ctx context.Context, userID uuid.UUID, input AddPermissionInput) (perm *model.CourseAdminPermission, err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionGrantPermission, adminOutcome(err)) }()

	if input.CourseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	} else if input.RoleID == uuid.Nil {
		return nil, BadRequest("role_id is required")
	} else if input.Permission == "" {
		return nil, BadRequest("permission is required")
	}
	if err = RequireScopedPermission(ctx, s.roleRepo, userID, input.CourseID, PermissionCoursePermissionAdd); err != nil {
		return nil, err
	}
	if err = s.requireRoleInCourse(ctx, input.RoleID, input.CourseID); err != nil {
		return nil, err
	}

	perm = &model.CourseAdminPermission{
		RoleID:     input.RoleID,
		Permission: input.Permission,
	}

	if err = s.roleRepo.AddPermission(ctx, perm); err != nil {
		return nil, Internal("Failed to add permission", err)
	}

	return perm, nil
}

func (s *AdminRoleService) RemovePermission(ctx context.Context, userID, courseID, roleID uuid.UUID, permission string) (err error) {
	defer func() { s.adminMetrics.IncAction(metrics.AdminActionRevokePermission, adminOutcome(err)) }()

	if courseID == uuid.Nil {
		return BadRequest("course_id is required")
	}
	if roleID == uuid.Nil {
		return BadRequest("role_id is required")
	}
	if permission == "" {
		return BadRequest("permission is required")
	}
	if err = RequireScopedPermission(ctx, s.roleRepo, userID, courseID, PermissionCoursePermissionRemove); err != nil {
		return err
	}
	if err = s.requireRoleInCourse(ctx, roleID, courseID); err != nil {
		return err
	}

	if err = s.roleRepo.RemovePermission(ctx, roleID, permission); err != nil {
		return Internal("Failed to remove permission", err)
	}

	return nil
}

func (s *AdminRoleService) ListPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) ([]model.CourseAdminPermission, error) {
	if courseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	if roleID == uuid.Nil {
		return nil, BadRequest("role_id is required")
	}
	if err := RequireScopedPermission(ctx, s.roleRepo, userID, courseID, PermissionCoursePermissionList); err != nil {
		return nil, err
	}
	if err := s.requireRoleInCourse(ctx, roleID, courseID); err != nil {
		return nil, err
	}

	perms, err := s.roleRepo.GetPermissions(ctx, roleID)
	if err != nil {
		return nil, Internal("Failed to fetch permissions", err)
	}

	return perms, nil
}

func (s *AdminRoleService) requireRoleInCourse(ctx context.Context, roleID, courseID uuid.UUID) error {
	ok, err := s.roleRepo.RoleBelongsToCourse(ctx, roleID, courseID)
	if err != nil {
		return Internal("Failed to check role course", err)
	}
	if !ok {
		return NotFound("Role not found in course")
	}
	return nil
}
