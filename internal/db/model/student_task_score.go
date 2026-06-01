package model

import (
    "time"
    "github.com/google/uuid"
)

type StudentTaskScore struct {
    ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    StudentID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_student_task" json:"student_id"`
    TaskID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_student_task" json:"task_id"`
    CourseID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"course_id"`
    Score     int        `gorm:"type:int;not null;default:0" json:"score"`
    IsPassed  bool       `gorm:"type:boolean;not null;default:false" json:"is_passed"`
    UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}