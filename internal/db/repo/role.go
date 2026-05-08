package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"fcstask-backend/internal/db/model"
)

type IRoleRepo interface {
	AssignRole(ctx context.Context, userRole *model.UserRole) error
	RevokeRole(ctx context.Context, userID, courseID, roleID uuid.UUID) error
	GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.UserRole, error)

	AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error
	RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error
	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error)
}

type RoleRepository struct {
	db *gorm.DB
}

var _ IRoleRepo = (*RoleRepository)(nil)

func NewRoleRepository(db *gorm.DB) IRoleRepo {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) AssignRole(ctx context.Context, userRole *model.UserRole) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(userRole).Error
}

func (r *RoleRepository) RevokeRole(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ? AND role_id = ?", userID, courseID, roleID).
		Delete(&model.UserRole{}).Error
}

func (r *RoleRepository) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.UserRole, error) {
	var roles []model.UserRole
	err := r.db.WithContext(ctx).
		Where("course_id = ?", courseID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(perm).Error
}

func (r *RoleRepository) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission = ?", roleID, permission).
		Delete(&model.CourseAdminPermission{}).Error
}

func (r *RoleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error) {
	var perms []model.CourseAdminPermission
	err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}
