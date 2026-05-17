package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Score struct {
	ID          int          `gorm:"primaryKey;autoIncrement" json:"-"`
	Student     string       `gorm:"type:varchar(255);not null" json:"-"`
	Score       int          `gorm:"type:int;not null;default:0" json:"-"`
	SubmittedAt *time.Time   `gorm:"type:timestamptz" json:"-"`
	CourseID    uuid.UUID    `gorm:"type:uuid;not null" json:"-"`
}

func (s *Score) TableName() string {
	return "scores"
}

type Namespace struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"-"`
	Name          string         `gorm:"type:varchar(255);not null" json:"-"`
	Slug          string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	Description   *string        `gorm:"type:text" json:"-"`
	GitlabGroupID *string        `gorm:"type:varchar(255)" json:"-"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"-"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (n *Namespace) TableName() string {
	return "namespaces"
}

// API Response types

type ScoreResponse struct {
	ID        int    `json:"id"`
	Student   string `json:"student"`
	Score     int    `json:"score"`
	Submitted string `json:"submitted"`
}

type NamespaceResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Description   string `json:"description,omitempty"`
	GitlabGroupID string `json:"gitlabGroupId"`
	CoursesCount  int    `json:"coursesCount"`
	UsersCount    int    `json:"usersCount"`
}

type NamespaceUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	RmsID    string `json:"rmsId"`
	Role     string `json:"role"`
}

type NamespaceCourse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	GitlabGroup string   `json:"gitlabGroup"`
	Owners      []string `json:"owners"`
	URL         string   `json:"url"`
}

type NamespaceDetailResponse struct {
	Namespace NamespaceResponse   `json:"namespace"`
	Users     []NamespaceUser     `json:"users"`
	Courses   []NamespaceCourse   `json:"courses"`
}

type InstanceSummary struct {
	TotalCourses    int    `json:"totalCourses"`
	TotalUsers      int    `json:"totalUsers"`
	TotalNamespaces int    `json:"totalNamespaces"`
	HealthStatus    string `json:"healthStatus"`
}