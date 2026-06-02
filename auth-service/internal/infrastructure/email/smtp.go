package email

import (
	"context"
	"fmt"

	"github.com/DoMinhHHung/auth-service/internal/domain/port"
	"github.com/DoMinhHHung/auth-service/internal/infrastructure/config"
	"gopkg.in/gomail.v2"
)

type smtpService struct {
	cfg config.EmailConfig
}

func NewSMTPService(cfg config.EmailConfig) port.EmailService {
	return &smtpService{cfg: cfg}
}

func (s *smtpService) SendSignupOTP(ctx context.Context, to, otp string) error {
	subject := "Verify your email - Rental Platform"
	body := fmt.Sprintf(`
        <html><body>
        <h2>Email Verification</h2>
        <p>Your OTP code is:</p>
        <h1 style="letter-spacing:5px;color:#2563eb">%s</h1>
        <p>This code expires in <strong>5 minutes</strong>.</p>
        <p>If you didn't request this, please ignore this email.</p>
        </body></html>
    `, otp)
	return s.send(to, subject, body)
}

func (s *smtpService) SendRecoveryOTP(ctx context.Context, to, otp string) error {
	subject := "Reset your password - Rental Platform"
	body := fmt.Sprintf(`
        <html><body>
        <h2>Password Reset</h2>
        <p>Your OTP code is:</p>
        <h1 style="letter-spacing:5px;color:#dc2626">%s</h1>
        <p>This code expires in <strong>5 minutes</strong>.</p>
        <p>If you didn't request this, please change your password immediately.</p>
        </body></html>
    `, otp)
	return s.send(to, subject, body)
}

func (s *smtpService) send(to, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(s.cfg.Host, s.cfg.Port, s.cfg.Username, s.cfg.Password)
	d.SSL = s.cfg.Port == 465

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email to %s: %w", to, err)
	}
	return nil
}
