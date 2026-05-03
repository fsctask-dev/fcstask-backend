package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailRegistration is a pending email/password signup. The User row is not
// created until the user proves they own the email by entering the code.
// Stores credentials_hash already bcrypt'd (never plain).
type EmailRegistration struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email          string    `gorm:"type:varchar(255);not null;index"`
	Username       string    `gorm:"type:varchar(255);not null"`
	PasswordHash   string    `gorm:"type:varchar(255);not null"`
	FirstName      *string   `gorm:"type:varchar(255)"`
	LastName       *string   `gorm:"type:varchar(255)"`
	CodeHash       string    `gorm:"type:varchar(64);not null"`
	Attempts       int       `gorm:"not null;default:0"`
	LastSentAt     time.Time `gorm:"not null"`
	ExpiresAt      time.Time `gorm:"not null;index"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

func (r *EmailRegistration) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
