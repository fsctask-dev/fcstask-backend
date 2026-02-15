package model

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Deadline struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	CourseID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"course_id"`
	Course      Course         `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"course"`
	DueDate     time.Time      `gorm:"type:timestamp;not null" json:"due_date"`
	AssignedBy  *uuid.UUID     `gorm:"type:uuid;index" json:"assigned_by,omitempty"`
	Assignee    *User          `gorm:"foreignKey:AssignedBy;constraint:OnDelete:SET NULL" json:"assignee,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (d *Deadline) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}