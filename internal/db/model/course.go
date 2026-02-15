package model

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Course struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	Code        string         `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	TeacherID   *uuid.UUID     `gorm:"type:uuid;index" json:"teacher_id,omitempty"`
	Teacher     *User          `gorm:"foreignKey:TeacherID;constraint:OnDelete:SET NULL" json:"teacher,omitempty"`
	StartDate   *time.Time     `gorm:"type:timestamp" json:"start_date,omitempty"`
	EndDate     *time.Time     `gorm:"type:timestamp" json:"end_date,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (c *Course) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}