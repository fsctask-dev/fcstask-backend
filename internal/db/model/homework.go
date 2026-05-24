package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	TaskID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"task_id"`
	HwID     uuid.UUID `gorm:"type:uuid;not null;index" json:"hw_id"`
	IsPublic *bool     `gorm:"type:bool;default:false" json:"is_public,omitempty"`
	RepoURL  *string   `gorm:"type:varchar(500)" json:"repo_url,omitempty"`
	TaskURL  *string   `gorm:"type:varchar(255);uniqueIndex" json:"task_url,omitempty"`
	Score    *int      `gorm:"type:int;default:null" json:"score,omitempty"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.TaskID == uuid.Nil {
		t.TaskID = uuid.New()
	}
	return nil
}

type Homework struct {
	HwID      uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"hw_id"`
	CourseID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"course_id"`
	IsPublic  *bool      `gorm:"type:bool;default:false" json:"is_public,omitempty"`
	StartDate *time.Time `gorm:"type:timestamp with time zone" json:"start_date,omitempty"`
	EndDate   *time.Time `gorm:"type:timestamp with time zone" json:"end_date,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (h *Homework) BeforeCreate(tx *gorm.DB) error {
	if h.HwID == uuid.Nil {
		h.HwID = uuid.New()
	}
	return nil
}
