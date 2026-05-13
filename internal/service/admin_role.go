package service

import (
	"context"

	"github.com/google/uuid"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type AdminRoleService struct {
	roleRepo repo.IRoleRepo
	userRepo repo.IUserRepo
}

func NewAdminRoleService(roleRepo repo.IRoleRepo, userRepo repo.IUserRepo) *AdminRoleService {
	return &AdminRoleService{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

type AssignRoleInput struct {
	UserID   uuid.UUID
	CourseID uuid.UUID
	RoleID   uuid.UUID
}

type RevokeRoleInput struct {
	UserID   uuid.UUID
	CourseID uuid.UUID
	RoleID   uuid.UUID
}

type AddPermissionInput struct {
	RoleID     uuid.UUID
	Permission string
}

func (s *AdminRoleService) AssignRole(ctx context.Context, userID uuid.UUID, input AssignRoleInput) (*model.UserRole, error) {
	if input.UserID == uuid.Nil {
		return nil, BadRequest("user_id is required")
	}
	if input.CourseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	if input.RoleID == uuid.Nil {
		return nil, BadRequest("role_id is required")
	}
    isAdmin, err := IsCourseAdmin(ctx, s.roleRepo, userID, input.CourseID)
	if err != nil {
		return nil, Internal("Failed to check permissions", err)
	}
	if !isAdmin {
		return nil, Forbidden("You don't have permission to manage this course")
	}
	if _, err := s.userRepo.GetUserByID(ctx, input.UserID); err != nil {
		return nil, NotFound("User not found")
	}
	userRole := &model.UserRole{
		UserID:   input.UserID,
		CourseID: input.CourseID,
		RoleID:   input.RoleID,
	}
	if err := s.roleRepo.AssignRole(ctx, userRole); err != nil {
		return nil, Internal("Failed to assign role", err)
	}

	return userRole, nil
}

func (s *AdminRoleService) RevokeRole(ctx context.Context, userID uuid.UUID, input RevokeRoleInput) error {
	if input.UserID == uuid.Nil {
		return BadRequest("user_id is required")
	}
	if input.CourseID == uuid.Nil {
		return BadRequest("course_id is required")
	}
	if input.RoleID == uuid.Nil {
		return BadRequest("role_id is required")
	}
    isAdmin, err := IsCourseAdmin(ctx, s.roleRepo, userID, input.CourseID)
	if err != nil {
		return Internal("Failed to check permissions", err)
	}
	if !isAdmin {
		return Forbidden("You don't have permission to manage this course")
	}
	if err := s.roleRepo.RevokeRole(ctx, input.UserID, input.CourseID, input.RoleID); err != nil {
		return Internal("Failed to revoke role", err)
	}

	return nil
}

func (s *AdminRoleService) ListUserRoles(ctx context.Context, userID uuid.UUID, courseID uuid.UUID) ([]model.UserRole, error) {
	if courseID == uuid.Nil {
		return nil, BadRequest("course_id is required")
	}
	isAdmin, err := IsCourseAdmin(ctx, s.roleRepo, userID, courseID)
	if err != nil {
		return nil, Internal("Failed to check permissions", err)
	}
	if !isAdmin {
		return nil, Forbidden("You don't have permission to manage this course")
	}
	roles, err := s.roleRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, Internal("Failed to fetch roles", err)
	}

	return roles, nil
}

func (s *AdminRoleService) AddPermission(ctx context.Context, input AddPermissionInput) (*model.CourseAdminPermission, error) {
	if input.RoleID == uuid.Nil {
		return nil, BadRequest("role_id is required")
	}
	if input.Permission == "" {
		return nil, BadRequest("permission is required")
	}

	perm := &model.CourseAdminPermission{
		RoleID:     input.RoleID,
		Permission: input.Permission,
	}

	if err := s.roleRepo.AddPermission(ctx, perm); err != nil {
		return nil, Internal("Failed to add permission", err)
	}

	return perm, nil
}

func (s *AdminRoleService) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	if roleID == uuid.Nil {
		return BadRequest("role_id is required")
	}
	if permission == "" {
		return BadRequest("permission is required")
	}

	if err := s.roleRepo.RemovePermission(ctx, roleID, permission); err != nil {
		return Internal("Failed to remove permission", err)
	}

	return nil
}

func (s *AdminRoleService) ListPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error) {
	if roleID == uuid.Nil {
		return nil, BadRequest("role_id is required")
	}

	perms, err := s.roleRepo.GetPermissions(ctx, roleID)
	if err != nil {
		return nil, Internal("Failed to fetch permissions", err)
	}

	return perms, nil
}
