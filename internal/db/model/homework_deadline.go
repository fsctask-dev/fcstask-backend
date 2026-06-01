package model

import (
	"time"

	"github.com/google/uuid"
)

type HomeworkDeadline struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	HwID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	SoftDeadline time.Time `gorm:"type:timestamptz;not null"`
	HardDeadline time.Time `gorm:"type:timestamptz;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (HomeworkDeadline) TableName() string { return "homework_deadlines" }
