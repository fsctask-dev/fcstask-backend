package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	Username  string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"username"` // с гитлаба
	FirstName *string        `gorm:"type:varchar(100)" json:"first_name,omitempty"`
	LastName  *string        `gorm:"type:varchar(100)" json:"last_name,omitempty"`
	TgUID     *int64         `gorm:"uniqueIndex" json:"tg_uid,omitempty"`
	UserID    int64          `gorm:"not null;uniqueIndex" json:"user_id"` // внутренний user_id
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
