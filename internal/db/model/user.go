package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	Username     string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"username"` // с гитлаба
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	FirstName    *string        `gorm:"type:varchar(100)" json:"first_name,omitempty"`
	LastName     *string        `gorm:"type:varchar(100)" json:"last_name,omitempty"`
	TgUID        *int64         `gorm:"uniqueIndex" json:"tg_uid,omitempty"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"` // внутренний user_id
	Sessions     []Session      `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

