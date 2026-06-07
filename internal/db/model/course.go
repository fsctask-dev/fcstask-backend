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
	InviteCode   *string        `gorm:"type:varchar(50);index" json:"invite_code,omitempty"`
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

type BoardDeadline struct {
	ID           string    `json:"id"`
	Label        string    `json:"label"`
	Percent      float64   `json:"percent"`
	SoftStatus   string    `json:"status"`
	HardStatus   string    `json:"hard_status"`
	SoftDeadline time.Time `json:"soft_deadline"`
	HardDeadline time.Time `json:"hard_deadline"`
}

type BoardTask struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Score       int     `json:"score"`
	ScoreEarned int     `json:"scoreEarned"`
	Stats       float64 `json:"stats"`
	IsBonus     *bool   `json:"isBonus,omitempty"`
	IsSpecial   *bool   `json:"isSpecial,omitempty"`
	URL         string  `json:"url,omitempty"`
}

type BoardGroup struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	IsSpecial *bool           `json:"isSpecial,omitempty"`
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

type TaskScore struct {
	TaskID uuid.UUID `json:"task_id"`
	Title  string    `json:"title"`
	Score  int       `json:"score"`
}

type HomeworkScore struct {
	HomeworkID    uuid.UUID   `json:"homework_id"`
	HomeworkTitle string      `json:"homework_title"`
	TotalScore    int         `json:"total_score"`
	Tasks         []TaskScore `json:"tasks"`
}

type LeaderboardEntry struct {
	Username   string          `json:"username"`
	TotalScore int             `json:"totalScore"`
	Homeworks  []HomeworkScore `json:"homeworks"`
	Rank       int             `json:"rank"`
}

type HomeworkWithTasks struct {
	Homework  `json:",inline"`
	Tasks     []Task     `json:"tasks"`
	Deadlines []Deadline `json:"deadlines,omitempty"`
}

type CourseInfo struct {
	Course    `json:",inline"`
	Homeworks []HomeworkWithTasks `json:"homework"`
}
