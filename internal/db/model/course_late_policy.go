package model

import (
	"time"

	"github.com/google/uuid"
)

type PolicyType string

const (
	PolicyTypeLinear      PolicyType = "linear"
	PolicyTypeStep        PolicyType = "step"
	PolicyTypeCoefficient PolicyType = "coefficient"
)

type CourseLatePolicy struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CourseID          uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"`
	PolicyType        PolicyType `gorm:"type:text;not null"`
	SoftPenalty       float64    `gorm:"type:numeric(5,4);not null;default:0"`
	HardDeadlineScore float64    `gorm:"type:numeric(5,4);not null;default:0"`
	StepPercent       *float64   `gorm:"type:numeric(5,4)"`
	Coefficient       *float64   `gorm:"type:numeric(5,4)"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (CourseLatePolicy) TableName() string { return "course_late_policies" }
