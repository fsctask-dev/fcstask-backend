package mailer

import (
	"context"
	"log"
)

// LogMailer prints codes to the application log. Useful for local development
// and tests; never wire it up in production.
type LogMailer struct{}

func NewLogMailer() *LogMailer { return &LogMailer{} }

func (m *LogMailer) SendVerificationCode(_ context.Context, to, username, code string) error {
	log.Printf("[mailer:dev] verification to=%s username=%s code=%s", to, username, code)
	return nil
}

func (m *LogMailer) SendPasswordResetCode(_ context.Context, to, code string) error {
	log.Printf("[mailer:dev] password_reset to=%s code=%s", to, code)
	return nil
}
