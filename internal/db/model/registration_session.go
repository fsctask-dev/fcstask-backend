package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegistrationSession is a short-lived row created the first time a user
// authenticates via OAuth and has no linked account yet. It carries the
// provider profile snapshot and tokens until the frontend posts the
// registration form to /api/oauth/complete-signup. Nothing in this row is
// authoritative — it gets discarded once the user is created.
type RegistrationSession struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Provider       string     `gorm:"type:varchar(32);not null;index"`
	ProviderUID    string     `gorm:"type:varchar(255);not null;index"`
	Email          string     `gorm:"type:varchar(255)"`
	Username       string     `gorm:"type:varchar(255)"`
	FirstName      string     `gorm:"type:varchar(255)"`
	LastName       string     `gorm:"type:varchar(255)"`
	AccessToken    string     `gorm:"type:text"`
	RefreshToken   string     `gorm:"type:text"`
	TokenExpiresAt *time.Time
	RawProfile     []byte    `gorm:"type:jsonb"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	ExpiresAt      time.Time `gorm:"not null;index"`
}

func (s *RegistrationSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
