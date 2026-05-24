package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LatePolicy struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HwID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"hw_id"`
	SoftDeadline time.Time `gorm:"type:timestamp with time zone;not null" json:"soft_deadline"`
	HardDeadline time.Time `gorm:"type:timestamp with time zone;not null" json:"hard_deadline"`
	SoftPenalty  float64   `gorm:"type:numeric(4,2);not null;default:0.7" json:"soft_penalty"`
	HardPenalty  float64   `gorm:"type:numeric(4,2);not null;default:0.0" json:"hard_penalty"`
}

func (lp *LatePolicy) BeforeCreate(tx *gorm.DB) error {
	if lp.ID == uuid.Nil {
		lp.ID = uuid.New()
	}
	return nil
}
