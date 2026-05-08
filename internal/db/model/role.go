package model

import "github.com/google/uuid"

type CourseAdminPermission struct {
	RoleID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_role_permission" json:"role_id"`
	Permission string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_role_permission" json:"permission"`
}

type UserRole struct {
	UserID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_course_role" json:"user_id"`
	CourseID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_course_role" json:"course_id"`
	RoleID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_course_role" json:"role_id"`
}
