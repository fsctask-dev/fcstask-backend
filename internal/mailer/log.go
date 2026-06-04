package mailer

import (
	"context"
	"log"

	"github.com/openframebox/gomail"
)

// LogMailer is the development mailer: instead of delivering anything it logs
// the message. Selected when MailerConfig.Enabled is false.
type LogMailer struct{}

var _ Mailer = (*LogMailer)(nil)

func NewLogMailer() *LogMailer {
	return &LogMailer{}
}

func (m *LogMailer) Send(ctx context.Context, to gomail.Address, subject, body string) error {
	log.Printf("[mailer] to=%s <%s> subject=%q\n%s", to.Name, to.Email, subject, body)
	return nil
}
