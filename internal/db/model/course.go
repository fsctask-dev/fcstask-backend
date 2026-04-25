package model

import (
	"time"
)

type CourseType string

const (
	CourseTypePublic  CourseType = "public"
	CourseTypePrivate CourseType = "private"
)

type Course struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	Name         string     `json:"name"`
	Slug         string     `gorm:"uniqueIndex" json:"slug"`
	Status       string     `json:"status"`
	Type         CourseType `gorm:"default:private" json:"type"` // новое поле
	StartDate    string     `json:"startDate"`
	EndDate      string     `json:"endDate"`
	RepoTemplate string     `json:"repoTemplate"`
	Description  string     `json:"description"`
	URL          string     `json:"url"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
