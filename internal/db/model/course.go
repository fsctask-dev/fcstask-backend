package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CourseType string

const (
	CourseTypePublic  CourseType = "public"
	CourseTypePrivate CourseType = "private"
)

type Course struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug         string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	Description  *string        `gorm:"type:text" json:"description,omitempty"`
	Status       string         `gorm:"type:varchar(50);not null;default:'created'" json:"status"`
	Type         CourseType     `gorm:"type:varchar(20);not null;default:'private'" json:"type"`
	InviteCode   *string        `gorm:"type:varchar(50);uniqueIndex" json:"invite_code,omitempty"`
	StartDate    *time.Time     `gorm:"type:timestamp" json:"start_date,omitempty"`
	EndDate      *time.Time     `gorm:"type:timestamp" json:"end_date,omitempty"`
	RepoTemplate *string        `gorm:"type:varchar(500)" json:"repo_template,omitempty"`
	URL          string         `gorm:"type:varchar(255)" json:"url"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (c *Course) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
