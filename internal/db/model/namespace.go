package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Namespace struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug          string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	Description   *string        `gorm:"type:text" json:"description,omitempty"`
	GitlabGroupID *string        `gorm:"type:varchar(255)" json:"gitlabGroupId,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type NamespaceUser struct {
	NamespaceID uuid.UUID `gorm:"type:uuid;not null" json:"namespace_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Role        string    `gorm:"type:varchar(50);not null;default:'student'" json:"role"`
}

func (NamespaceUser) TableName() string { return "namespace_users" }

type NamespaceCourse struct {
	NamespaceID uuid.UUID `gorm:"type:uuid;not null" json:"namespace_id"`
	CourseID    uuid.UUID `gorm:"type:uuid;not null" json:"course_id"`
}

func (NamespaceCourse) TableName() string { return "namespace_courses" }
