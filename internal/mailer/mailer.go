package mailer

import (
	"context"
	"fcstask-backend/internal/db/model"
	"fmt"

	"github.com/openframebox/gomail"
)

type Mailer interface {
	Send(ctx context.Context, to gomail.Address, subject, body string) error
}

func SendPasswordReset(mailer Mailer, ctx context.Context, data *model.PasswordReset, code string) error {
	return mailer.Send(ctx, gomail.Address{Name: data.User.Username, Email: data.User.Email}, "password reset", fmt.Sprintf("name: %s\ncode: %s\nurl: %s", data.User.Username, code, fmt.Sprintf("http://example.com/password_reset?id=%s&code=%s", data.ID, code)))
}

func SendEmailConfirmation(mailer Mailer, ctx context.Context, data *model.EmailRegistration, code string) error {
	return mailer.Send(ctx, gomail.Address{Name: data.Username, Email: data.Email}, "email confirmation", fmt.Sprintf("name: %s\ncode: %s\nurl:  %s\n", data.Username, code, fmt.Sprintf("http://example.com/password_reset?id=%s&code=%s", data.ID, code)))
}
