package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	OAuthProviderGitLab   = "gitlab"
	OAuthProviderGoogle   = "google"
	OAuthProviderTelegram = "telegram"
)

// OAuthIdentity is a permanent link between a User and a remote provider account.
// Rows in this table only exist for accounts that finished registration; the
// short-lived "user authenticated with provider but hasn't picked a username
// yet" state lives in registration_sessions instead.
type OAuthIdentity struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User         *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Provider     string     `gorm:"type:varchar(32);not null;uniqueIndex:idx_oauth_provider_uid" json:"provider"`
	ProviderUID  string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_oauth_provider_uid" json:"provider_uid"`
	AccessToken  string     `gorm:"type:text" json:"-"`
	RefreshToken string     `gorm:"type:text" json:"-"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	RawProfile   []byte     `gorm:"type:jsonb" json:"-"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (i *OAuthIdentity) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
