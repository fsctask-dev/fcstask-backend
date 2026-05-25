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
	GitlabGroupID *string        `gorm:"type:varchar(255)" json:"gitlab_group_id,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (n *Namespace) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

type NamespaceUser struct {
	NamespaceID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_namespace_user" json:"namespace_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_namespace_user" json:"user_id"`
	Role        string    `gorm:"type:varchar(50);not null;default:'student'" json:"role"`
}
