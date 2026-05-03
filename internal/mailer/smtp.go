package mailer

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"fcstask-backend/internal/config"
)

// SMTPMailer sends mail through a plain SMTP server using STARTTLS when the
// server advertises it. Authentication is PLAIN if username/password are set.
type SMTPMailer struct {
	cfg config.MailerConfig
}

func NewSMTPMailer(cfg config.MailerConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) SendVerificationCode(ctx context.Context, to, username, code string) error {
	subject := "Подтверждение регистрации"
	body := fmt.Sprintf("Здравствуйте, %s!\r\n\r\nВаш код подтверждения: %s\r\nКод действителен 15 минут.\r\n",
		username, code)
	return m.send(ctx, to, subject, body)
}

func (m *SMTPMailer) SendPasswordResetCode(ctx context.Context, to, code string) error {
	subject := "Восстановление пароля"
	body := fmt.Sprintf("Ваш код для сброса пароля: %s\r\nКод действителен 15 минут.\r\nЕсли вы не запрашивали сброс — проигнорируйте письмо.\r\n", code)
	return m.send(ctx, to, subject, body)
}

func (m *SMTPMailer) send(_ context.Context, to, subject, body string) error {
	addr := net.JoinHostPort(m.cfg.SMTP.Host, strconv.Itoa(m.cfg.SMTP.Port))
	from := m.cfg.From
	if from == "" {
		from = m.cfg.SMTP.Username
	}

	msg := strings.Builder{}
	msg.WriteString("From: ")
	msg.WriteString(from)
	msg.WriteString("\r\nTo: ")
	msg.WriteString(to)
	msg.WriteString("\r\nSubject: ")
	msg.WriteString(subject)
	msg.WriteString("\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
	msg.WriteString(body)

	var auth smtp.Auth
	if m.cfg.SMTP.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTP.Username, m.cfg.SMTP.Password, m.cfg.SMTP.Host)
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg.String()))
}
