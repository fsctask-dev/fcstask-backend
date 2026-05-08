package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AdminAuditLog struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AdminID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"admin_id"`
	Admin       *User          `gorm:"foreignKey:AdminID;constraint:OnDelete:SET NULL" json:"admin,omitempty"`
	TargetType  *string        `gorm:"type:varchar(50)" json:"target_type,omitempty"`
	TargetID    *uuid.UUID     `gorm:"type:uuid;index:idx_audit_target" json:"target_id,omitempty"`
	Action      string         `gorm:"type:varchar(50);not null;index" json:"action"`
	OldData     datatypes.JSON `gorm:"type:jsonb" json:"old_data,omitempty"`
	NewData     datatypes.JSON `gorm:"type:jsonb" json:"new_data,omitempty"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	IPAddress   *string        `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
}

func (a *AdminAuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
