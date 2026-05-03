package mailer

import "context"

// Mailer abstracts how the backend delivers transactional emails so handlers
// don't care whether SMTP is configured. Implementations: LogMailer (dev) and
// SMTPMailer (prod).
type Mailer interface {
	SendVerificationCode(ctx context.Context, to, username, code string) error
	SendPasswordResetCode(ctx context.Context, to, code string) error
}
