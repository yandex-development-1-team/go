package service

import (
	"context"
	"fmt"

	"github.com/go-mail/mail/v2"

	"github.com/yandex-development-1-team/go/internal/config"
)

type EmailService struct {
	dialer  *mail.Dialer
	from    string
	baseURL string
}

func NewEmailService(cfg config.EmailConfig) *EmailService {
	dialer := mail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	return &EmailService{
		dialer:  dialer,
		from:    cfg.FromEmail,
		baseURL: cfg.BaseURL,
	}
}

func (s *EmailService) SendPasswordResetEmail(ctx context.Context, toEmail, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, resetToken)

	m := mail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Восстановление пароля")
	m.SetBody("text/plain", fmt.Sprintf("Для восстановления пароля перейдите по ссылке: %s\n\nСсылка действительна 1 час.", resetURL))

	return s.dialer.DialAndSend(m)
}
