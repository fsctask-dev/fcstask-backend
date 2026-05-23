package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"fcstask-backend/internal/db/model"
)

type IRoleRepo interface {
	AssignRoleWithPermissions(ctx context.Context, userRole *model.UserRole, permissions []string) error
	RevokeRoleWithPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) error
	GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.UserRole, error)
	GetRoleIDByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (uuid.UUID, error)
	RoleBelongsToCourse(ctx context.Context, roleID, courseID uuid.UUID) (bool, error)
	HasPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error)

	AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error
	AddPermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error
	RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error
	RemovePermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error
	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error)
}

type RoleRepository struct {
	db *gorm.DB
}

var _ IRoleRepo = (*RoleRepository)(nil)

func NewRoleRepository(db *gorm.DB) IRoleRepo {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) AssignRoleWithPermissions(ctx context.Context, userRole *model.UserRole, permissions []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "course_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"role_id"}),
			}).
			Create(userRole).Error; err != nil {
			return err
		}

		if len(permissions) == 0 {
			return nil
		}

		perms := make([]model.CourseAdminPermission, 0, len(permissions))
		for _, permission := range permissions {
			perms = append(perms, model.CourseAdminPermission{
				RoleID:     userRole.RoleID,
				Permission: permission,
			})
		}

		return tx.
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&perms).Error
	})
}

func (r *RoleRepository) RevokeRoleWithPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("role_id = ?", roleID).
			Delete(&model.CourseAdminPermission{}).Error; err != nil {
			return err
		}

		return tx.
			Where("user_id = ? AND course_id = ? AND role_id = ?", userID, courseID, roleID).
			Delete(&model.UserRole{}).Error
	})
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

func (r *RoleRepository) GetRoleIDByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (uuid.UUID, error) {
	var role model.UserRole
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, gorm.ErrRecordNotFound
		}
		return uuid.Nil, err
	}
	return role.RoleID, nil
}

func (r *RoleRepository) RoleBelongsToCourse(ctx context.Context, roleID, courseID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM user_roles
			WHERE role_id = ?
			  AND course_id = ?
		)
	`, roleID, courseID).Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *RoleRepository) HasPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM course_admin_permissions cap
			WHERE cap.role_id = ?
			  AND cap.permission = ?
		)
	`, roleID, permission).Scan(&exists).Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *RoleRepository) AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(perm).Error
}

func (r *RoleRepository) AddPermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	perms := make([]model.CourseAdminPermission, 0, len(permissions))
	for _, permission := range permissions {
		perms = append(perms, model.CourseAdminPermission{
			RoleID:     roleID,
			Permission: permission,
		})
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&perms).Error
	})
}

func (r *RoleRepository) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission = ?", roleID, permission).
		Delete(&model.CourseAdminPermission{}).Error
}

func (r *RoleRepository) RemovePermissions(ctx context.Context, roleID uuid.UUID, permissions []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.
			Where("role_id = ? AND permission IN ?", roleID, permissions).
			Delete(&model.CourseAdminPermission{}).Error
	})
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
