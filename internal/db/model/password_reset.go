package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PasswordReset is a pending "I forgot my password" flow. We don't allow
// arbitrarily many simultaneous outstanding resets per user — old rows are
// overwritten by Upsert in the repo.
type PasswordReset struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	User       *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	CodeHash   string    `gorm:"type:varchar(64);not null"`
	Attempts   int       `gorm:"not null;default:0"`
	LastSentAt time.Time `gorm:"not null"`
	ExpiresAt  time.Time `gorm:"not null;index"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (r *PasswordReset) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
