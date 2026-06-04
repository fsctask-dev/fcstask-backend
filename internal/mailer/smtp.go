package mailer

import (
	"context"
	"fcstask-backend/internal/config"

	"github.com/openframebox/gomail"
)

type SMTPMailer struct {
	cfg    config.MailerConfig
	mailer *gomail.GoMail
	pool   *gomail.Pool
}

func NewSMTPMailer(cfg config.MailerConfig) (*SMTPMailer, error) {
	mailer, err := gomail.New(&gomail.Config{
		SMTPHost: cfg.SMTP.Host,
		SMTPPort: cfg.SMTP.Port,
		SMTPUser: cfg.SMTP.Username,
		SMTPPass: cfg.SMTP.Password,
	})
	if err != nil {
		return nil, err
	}
	pool := mailer.NewPool(gomail.PoolConfig{
		MaxSize: cfg.SMTP.PoolSize,
	})
	return &SMTPMailer{cfg: cfg, pool: pool, mailer: mailer}, nil
}

func (m *SMTPMailer) Send(ctx context.Context, mail *gomail.Mail) error {
	return m.pool.Send(ctx, mail)
}

func (m *SMTPMailer) Close() error {
	return m.pool.Close()
}
