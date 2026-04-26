package model

import "time"

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
	Type         CourseType `gorm:"default:private" json:"type"`
	StartDate    string     `json:"startDate"`
	EndDate      string     `json:"endDate"`
	RepoTemplate string     `json:"repoTemplate"`
	Description  string     `json:"description"`
	URL          string     `json:"url"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type BoardDeadline struct {
	ID      string  `json:"id"`
	Label   string  `json:"label"`
	Percent float64 `json:"percent"`
	DueAt   string  `json:"dueAt"`
	Status  string  `json:"status"`
}

type BoardTask struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Score       int     `json:"score"`
	ScoreEarned int     `json:"scoreEarned"`
	Stats       float64 `json:"stats"`
	IsBonus     bool    `json:"isBonus,omitempty"`
	IsSpecial   bool    `json:"isSpecial,omitempty"`
	URL         string  `json:"url,omitempty"`
}

type BoardGroup struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	IsSpecial bool            `json:"isSpecial,omitempty"`
	StartedAt string          `json:"startedAt"`
	EndsAt    string          `json:"endsAt"`
	Deadlines []BoardDeadline `json:"deadlines"`
	Tasks     []BoardTask     `json:"tasks"`
}

type TaskBoardSummary struct {
	CourseName    string       `json:"courseName"`
	CourseStatus  string       `json:"courseStatus"`
	SolvedScore   int          `json:"solvedScore"`
	MaxScore      int          `json:"maxScore"`
	SolvedPercent int          `json:"solvedPercent"`
	Groups        []BoardGroup `json:"groups"`
}
